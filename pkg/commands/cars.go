package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bg-cars-discord-bot/pkg/discord"
	"bg-cars-discord-bot/pkg/scraper"

	"github.com/bwmarrin/discordgo"
)

// CarsCommand handles the !cars command
func CarsCommand(s *discordgo.Session, channelID string, args []string) string {
	// Default values
	brand := ""
	model := ""
	maxPages := 2

	// Parse arguments
	if len(args) >= 1 {
		brand = args[0]
	}
	if len(args) >= 2 {
		model = args[1]
	}
	if len(args) >= 3 {
		if pages, err := strconv.Atoi(args[2]); err == nil && pages > 0 && pages <= 10 {
			maxPages = pages
		}
	}

	// Start the search asynchronously
	go searchAndSendResults(s, channelID, brand, model, maxPages)

	// Return immediate response
	return buildSearchStartMessage(brand, model, maxPages)
}

// searchAndSendResults performs the car search and sends results to Discord
func searchAndSendResults(s *discordgo.Session, channelID, brand, model string, maxPages int) {
	ctx := context.Background()
	fmt.Printf("Starting car search: brand=%s, model=%s, maxPages=%d\n", brand, model, maxPages)

	offers, err := scraper.SearchCars(ctx, maxPages, brand, model)
	if err != nil {
		fmt.Printf("Error searching cars: %v\n", err)
		s.ChannelMessageSend(channelID, fmt.Sprintf("❌ **Error searching cars:** %v", err))
		return
	}

	fmt.Printf("Found %d car offers\n", len(offers))

	if len(offers) == 0 {
		s.ChannelMessageSend(channelID, "🔍 **No cars found** matching your criteria.")
		return
	}

	// Send results to Discord
	sendCarResults(s, channelID, offers)
}

// sendCarResults sends car search results to Discord channel using embeds
func sendCarResults(s *discordgo.Session, channelID string, offers []scraper.Offer) {
	const maxResults = 10 // Limit to prevent Discord message spam

	// Send summary first
	summaryMsg := discord.CreateSearchSummaryMessage(len(offers), maxResults)
	s.ChannelMessageSend(channelID, summaryMsg)

	// Send each car as a rich embed
	resultCount := 0
	for i, offer := range offers {
		if i >= maxResults {
			break
		}

		// Create embed for this car
		embed := discord.CreateCarEmbed(offer, i+1, len(offers))

		// Send the embed
		_, err := s.ChannelMessageSendEmbed(channelID, embed)
		if err != nil {
			fmt.Printf("Error sending embed for car %d: %v\n", i+1, err)
			// Fallback to text message if embed fails
			fallbackMsg := discord.CreateCarFallbackMessage(offer)
			s.ChannelMessageSend(channelID, fallbackMsg)
		}

		resultCount++
	}

	// Send completion message
	footerMsg := discord.CreateSearchCompleteMessage(resultCount, len(offers))
	s.ChannelMessageSend(channelID, footerMsg)
}

// buildSearchStartMessage creates the initial response message
func buildSearchStartMessage(brand, model string, maxPages int) string {
	var searchDetails strings.Builder
	searchDetails.WriteString("🚗 **Searching for cars...**\n")

	if brand != "" {
		searchDetails.WriteString(fmt.Sprintf("**Brand:** %s\n", strings.ToUpper(brand)))
	} else {
		searchDetails.WriteString("**Brand:** All brands\n")
	}

	if model != "" {
		searchDetails.WriteString(fmt.Sprintf("**Model:** %s\n", model))
	} else {
		searchDetails.WriteString("**Model:** All models\n")
	}

	searchDetails.WriteString(fmt.Sprintf("**Max Pages:** %d\n", maxPages))
	searchDetails.WriteString("\n*Results will be sent here shortly...*")

	return searchDetails.String()
}
