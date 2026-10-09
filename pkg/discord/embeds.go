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

	title := strings.TrimSpace(offer.Title)
	if title == "" {
		title = "Car listing"
	}

	// Create Discord embed for the car
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: "Open the listing for full details.",
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
	if strings.TrimSpace(offer.ImageURL) != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: strings.TrimSpace(offer.ImageURL),
		}
	}

	// Add link if available
	if strings.TrimSpace(offer.ListLink) != "" {
		embed.URL = strings.TrimSpace(offer.ListLink)
	}

	// Add data item as a field if available
	if strings.TrimSpace(offer.DataItem) != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📋 Listing ID",
			Value:  strings.TrimSpace(offer.DataItem),
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
	title := strings.TrimSpace(offer.Title)
	if title == "" {
		title = "Car listing"
	}
	price := strings.TrimSpace(offer.Price)
	if price == "" {
		price = "Price not available"
	}
	fallbackMsg := fmt.Sprintf("🚗 **%s**\n💰 %s", title, price)
	if strings.TrimSpace(offer.ListLink) != "" {
		fallbackMsg += fmt.Sprintf("\n🔗 %s", strings.TrimSpace(offer.ListLink))
	}
	if strings.TrimSpace(offer.DataItem) != "" {
		fallbackMsg += fmt.Sprintf("\n📋 Listing ID: %s", strings.TrimSpace(offer.DataItem))
	}
	return fallbackMsg
}
