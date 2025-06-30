package bot

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"bg-cars-discord-bot/pkg/commands"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot instance
type Bot struct {
	session *discordgo.Session
	token   string
}

// New creates a new Bot instance
func New(token string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}

	bot := &Bot{
		session: session,
		token:   token,
	}

	// Register event handlers
	bot.registerHandlers()

	return bot, nil
}

// Start starts the bot and blocks until interrupted
func (b *Bot) Start() error {
	// Set bot intents
	b.session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Open WebSocket connection
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %v", err)
	}

	fmt.Println("🤖 Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	b.waitForInterrupt()

	return nil
}

// Stop gracefully stops the bot
func (b *Bot) Stop() error {
	fmt.Println("🛑 Shutting down bot...")
	return b.session.Close()
}

// registerHandlers registers all event handlers for the bot
func (b *Bot) registerHandlers() {
	b.session.AddHandler(b.onReady)
	b.session.AddHandler(b.onMessageCreate)
}

// onReady handles the ready event when bot connects
func (b *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("✅ Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)
	fmt.Printf("🔗 Bot ID: %v\n", s.State.User.ID)
	fmt.Printf("🌐 Connected to %d guilds\n", len(event.Guilds))

	// Set bot status
	err := s.UpdateGameStatus(0, "🚗 !cars <brand> <model>")
	if err != nil {
		fmt.Printf("Error setting status: %v\n", err)
	}
}

// onMessageCreate handles incoming messages
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots (including ourselves)
	if m.Author.Bot {
		return
	}

	// Process the message
	b.processMessage(s, m)
}

// processMessage processes incoming messages and handles commands
func (b *Bot) processMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := strings.TrimSpace(m.Content)

	// Check if message starts with command prefix
	if !strings.HasPrefix(content, "!") {
		return
	}

	// Parse command and arguments
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	// Route commands
	switch command {
	case "!cars":
		response := commands.CarsCommand(s, m.ChannelID, args)
		if response != "" {
			s.ChannelMessageSend(m.ChannelID, response)
		}
	case "!help":
		b.sendHelpMessage(s, m.ChannelID)
	case "!ping":
		s.ChannelMessageSend(m.ChannelID, "🏓 Pong!")
	default:
		// Unknown command
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❓ Unknown command: `%s`\nType `!help` for available commands.", command))
	}
}

// sendHelpMessage sends the help message
func (b *Bot) sendHelpMessage(s *discordgo.Session, channelID string) {
	helpText := `🤖 **Car Search Bot Commands**

**!cars** [brand] [model] [pages]
Search for cars on cars.bg
• **brand** - Car brand (optional, e.g., BMW, Audi)
• **model** - Car model (optional, e.g., X5, A4)
• **pages** - Number of pages to search (1-10, default: 2)

**Examples:**
• ` + "`!cars`" + ` - Search all cars (first 2 pages)
• ` + "`!cars BMW`" + ` - Search all BMW cars
• ` + "`!cars BMW X5`" + ` - Search BMW X5 specifically
• ` + "`!cars BMW X5 5`" + ` - Search BMW X5 (first 5 pages)

**Other Commands:**
• **!help** - Show this help message
• **!ping** - Test if bot is responsive

*Results are displayed as rich embeds with car images and details.*`

	s.ChannelMessageSend(channelID, helpText)
}

// waitForInterrupt waits for an interrupt signal to gracefully shutdown
func (b *Bot) waitForInterrupt() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
