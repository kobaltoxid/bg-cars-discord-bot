package discord

import (
	"fmt"
	"strings"

	"bg-cars-discord-bot/pkg/scraper"

	"github.com/bwmarrin/discordgo"
)

// CreateCarEmbed creates a Discord embed for a car listing
func CreateCarEmbed(offer scraper.Offer, resultIndex, totalResults int) *discordgo.MessageEmbed {
	// Clean and format the price
	cleanPrice := strings.TrimSpace(offer.Price)
	if cleanPrice == "" {
		cleanPrice = "Price not available"
	}

	// Create Discord embed for the car
	embed := &discordgo.MessageEmbed{
		Title:       offer.Title,
		Description: "Click the title to view the full listing",
		Color:       0x3498db, // Blue color
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Result %d of %d", resultIndex, totalResults),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "💰 Price",
				Value:  cleanPrice,
				Inline: true,
			},
		},
	}

	// Add image if available
	if offer.ImageURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: offer.ImageURL,
		}
	}

	// Add link if available
	if offer.ListLink != "" {
		embed.URL = offer.ListLink
	}

	// Add data item as a field if available
	if offer.DataItem != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📋 Listing ID",
			Value:  offer.DataItem,
			Inline: true,
		})
	}

	return embed
}

// CreateSearchSummaryMessage creates a summary message for search results
func CreateSearchSummaryMessage(totalResults, maxResults int) string {
	summaryMsg := fmt.Sprintf("🎉 **Found %d car(s)**", totalResults)
	if totalResults > maxResults {
		summaryMsg += fmt.Sprintf(" (showing first %d)", maxResults)
	}
	return summaryMsg
}

// CreateSearchCompleteMessage creates a completion message for search results
func CreateSearchCompleteMessage(resultCount, totalResults int) string {
	footerMsg := fmt.Sprintf("✅ **Search complete!** Showing %d results", resultCount)
	if totalResults > resultCount {
		footerMsg += fmt.Sprintf(" out of %d total found", totalResults)
	}
	return footerMsg
}

// CreateCarFallbackMessage creates a fallback text message if embed fails
func CreateCarFallbackMessage(offer scraper.Offer) string {
	fallbackMsg := fmt.Sprintf("🚗 **%s**\n💰 %s", offer.Title, offer.Price)
	if offer.ListLink != "" {
		fallbackMsg += fmt.Sprintf("\n🔗 %s", offer.ListLink)
	}
	return fallbackMsg
}
