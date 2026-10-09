package bot

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"bg-cars-discord-bot/pkg/commands"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot instance
type Bot struct {
	session  *discordgo.Session
	token    string
	stopOnce sync.Once
	stopDone chan struct{}
	stopErr  error

	lifecycleMu sync.Mutex
	started     bool
	stopped     bool
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
		session:  session,
		token:    token,
		stopDone: make(chan struct{}),
	}

	// Register event handlers
	bot.registerHandlers()

	return bot, nil
}

// Start starts the bot and blocks until interrupted
func (b *Bot) Start() error {
	b.lifecycleMu.Lock()
	if b.stopped {
		b.lifecycleMu.Unlock()
		return fmt.Errorf("bot has been stopped")
	}
	if b.started {
		b.lifecycleMu.Unlock()
		return fmt.Errorf("bot is already started")
	}
	b.started = true

	// Set bot intents
	// MessageContent is required for Discord to populate m.Content. Guilds is
	// included because the ready event and guild state are used by the bot.
	b.session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	// Open WebSocket connection
	err := b.session.Open()
	if err != nil {
		b.started = false
		b.lifecycleMu.Unlock()
		return fmt.Errorf("error opening connection: %v", err)
	}
	b.lifecycleMu.Unlock()

	fmt.Println("🤖 Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	return b.waitForInterrupt()
}

// Stop gracefully stops the bot
func (b *Bot) Stop() error {
	b.stopOnce.Do(func() {
		fmt.Println("🛑 Shutting down bot...")

		b.lifecycleMu.Lock()
		b.stopped = true
		started := b.started
		b.lifecycleMu.Unlock()

		if started {
			b.stopErr = b.session.Close()
		}
		close(b.stopDone)
	})
	return b.stopErr
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

	// Process the message.
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
			b.sendMessage(s, m.ChannelID, response)
		}
	case "!help":
		b.sendHelpMessage(s, m.ChannelID)
	case "!ping":
		b.sendMessage(s, m.ChannelID, "🏓 Pong!")
	default:
		// Unknown command
		b.sendMessage(s, m.ChannelID, fmt.Sprintf("❓ Unknown command: `%s`\nType `!help` for available commands.", command))
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
• ` + "`!cars BMW 5`" + ` - Search all BMW cars across the first 5 pages
• ` + "`!cars BMW X5 5`" + ` - Search BMW X5 (first 5 pages)

**Other Commands:**
• **!help** - Show this help message
• **!ping** - Test if bot is responsive

*Results are displayed as rich embeds with car images and details.*`

	b.sendMessage(s, channelID, helpText)
}

// waitForInterrupt waits for an interrupt signal to gracefully shutdown
func (b *Bot) waitForInterrupt() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case <-stop:
		return b.Stop()
	case <-b.stopDone:
		return b.stopErr
	}
}

func (b *Bot) sendMessage(s *discordgo.Session, channelID, content string) {
	if _, err := s.ChannelMessageSend(channelID, content); err != nil {
		fmt.Printf("Error sending message to channel %s: %v\n", channelID, err)
	}
}
