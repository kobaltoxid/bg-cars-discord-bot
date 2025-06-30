package main

import (
	"fmt"
	"log"
	"os"

	"bg-cars-discord-bot/pkg/bot"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get Discord bot token from environment
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ DISCORD_BOT_TOKEN environment variable is required")
	}

	// Create and start the bot
	discordBot, err := bot.New(token)
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	// Start the bot (this blocks until interrupted)
	err = discordBot.Start()
	if err != nil {
		log.Fatalf("❌ Failed to start bot: %v", err)
	}

	// Graceful shutdown
	err = discordBot.Stop()
	if err != nil {
		log.Printf("⚠️ Error during shutdown: %v", err)
	}

	fmt.Println("👋 Bot stopped successfully")
}
