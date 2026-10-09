package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bg-cars-discord-bot/pkg/discord"
	"bg-cars-discord-bot/pkg/scraper"

	"github.com/bwmarrin/discordgo"
)

const (
	defaultMaxPages       = 2
	maxPagesLimit         = 10
	maxResults            = 10
	searchTimeout         = 2 * time.Minute
	maxConcurrentSearches = 4
	supportedBrandList    = "BMW, Audi, VW, Mercedes, Toyota, Mitsubishi, Honda, Ford, Opel, Renault, Mazda, Citroen, Peugeot, Nissan, Skoda, Fiat, Hyundai, Kia, Volvo, Suzuki"
)

var searchSlots = make(chan struct{}, maxConcurrentSearches)

// ParseCarsArgs validates the positional arguments accepted by !cars.
func ParseCarsArgs(args []string) (brand, model string, maxPages int, err error) {
	if len(args) > 3 {
		return "", "", 0, fmt.Errorf("too many arguments")
	}
	if len(args) > 0 {
		brand = strings.TrimSpace(args[0])
	}
	if len(args) > 1 {
		second := strings.TrimSpace(args[1])
		if len(args) == 2 {
			if pages, parseErr := strconv.Atoi(second); parseErr == nil {
				if pages < 1 || pages > maxPagesLimit {
					return "", "", 0, fmt.Errorf("pages must be a whole number from 1 to %d", maxPagesLimit)
				}
				return brand, "", pages, nil
			}
		}
		model = second
	}
	maxPages = defaultMaxPages
	if len(args) == 3 {
		maxPages, err = strconv.Atoi(strings.TrimSpace(args[2]))
		if err != nil || maxPages < 1 || maxPages > maxPagesLimit {
			return "", "", 0, fmt.Errorf("pages must be a whole number from 1 to %d", maxPagesLimit)
		}
	}
	return brand, model, maxPages, nil
}

func carsUsage() string {
	return "Usage: `!cars [brand] [model] [pages]`\nPages must be a whole number from 1 to 10 (default: 2)."
}

// CarsCommand handles the !cars command
func CarsCommand(s *discordgo.Session, channelID string, args []string) string {
	brand, model, maxPages, err := ParseCarsArgs(args)
	if err != nil {
		return fmt.Sprintf("❌ **Invalid !cars arguments:** %v\n%s", err, carsUsage())
	}
	if brand != "" && scraper.BrandNameToID(brand) == scraper.BrandUnknown {
		return fmt.Sprintf("❌ **Unsupported brand:** `%s`\nSupported brands: %s.\n%s", brand, supportedBrandList, carsUsage())
	}
	if s == nil || strings.TrimSpace(channelID) == "" {
		return "❌ I couldn't start that search because the Discord channel is unavailable."
	}
	select {
	case searchSlots <- struct{}{}:
		go searchAndSendResults(s, channelID, brand, model, maxPages)
	default:
		return "⏳ Too many searches are already running. Please try again shortly."
	}

	// Return immediate response
	return buildSearchStartMessage(brand, model, maxPages)
}

// searchAndSendResults performs the car search and sends results to Discord
func searchAndSendResults(s *discordgo.Session, channelID, brand, model string, maxPages int) {
	defer func() { <-searchSlots }()
	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()
	fmt.Printf("Starting car search: brand=%s, model=%s, maxPages=%d\n", brand, model, maxPages)

	offers, err := scraper.SearchCars(ctx, maxPages, brand, model)
	if err != nil {
		log.Printf("Error searching cars: %v", err)
		sendMessage(s, channelID, fmt.Sprintf("❌ **Error searching cars:** %v", err))
		return
	}

	fmt.Printf("Found %d car offers\n", len(offers))

	if len(offers) == 0 {
		sendMessage(s, channelID, "🔍 **No cars found** matching your criteria.")
		return
	}

	// Send results to Discord
	sendCarResults(s, channelID, offers)
}

// sendCarResults sends car search results to Discord channel using embeds
func sendCarResults(s *discordgo.Session, channelID string, offers []scraper.Offer) {
	// Send summary first
	summaryMsg := discord.CreateSearchSummaryMessage(len(offers), maxResults)
	sendMessage(s, channelID, summaryMsg)

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
			log.Printf("Error sending embed for car %d: %v", i+1, err)
			// Fallback to text message if embed fails
			fallbackMsg := discord.CreateCarFallbackMessage(offer)
			sendMessage(s, channelID, fallbackMsg)
		}

		resultCount++
	}

	// Send completion message
	footerMsg := discord.CreateSearchCompleteMessage(resultCount, len(offers))
	sendMessage(s, channelID, footerMsg)
}

func sendMessage(s *discordgo.Session, channelID, content string) {
	if _, err := s.ChannelMessageSend(channelID, content); err != nil {
		log.Printf("Error sending message to channel %s: %v", channelID, err)
	}
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
