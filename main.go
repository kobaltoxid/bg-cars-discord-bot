package main

import (
	"log"
	"os"
	"strings"

	"bg-cars-discord-bot/pkg/bot"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load .env: %v", err)
	}

	// Get Discord bot token from environment
	token := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	if token == "" {
		log.Fatal("❌ DISCORD_BOT_TOKEN environment variable is required")
	}

	// Create and start the bot
	discordBot, err := bot.New(token)
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	// Start the bot (this blocks until interrupted)
	if err := discordBot.Start(); err != nil {
		log.Fatalf("❌ Failed to start bot: %v", err)
	}

	// Graceful shutdown
	if err := discordBot.Stop(); err != nil {
		log.Printf("⚠️ Error during shutdown: %v", err)
	}

	log.Println("👋 Bot stopped successfully")
}
