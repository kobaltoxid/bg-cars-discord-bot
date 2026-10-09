package discord

import (
	"testing"

	"bg-cars-discord-bot/pkg/scraper"
)

func TestCreateCarEmbedIncludesOfferDetailsAndFallbacks(t *testing.T) {
	embed := CreateCarEmbed(scraper.Offer{
		Title:    "  BMW X5  ",
		Price:    " 30 000 EUR ",
		ImageURL: "https://example.com/car.jpg",
		ListLink: " https://cars.bg/listing/1 ",
		DataItem: " 123 ",
	}, 2, 4)

	if embed.Title != "BMW X5" || embed.URL != "https://cars.bg/listing/1" || embed.Image.URL != "https://example.com/car.jpg" {
		t.Fatalf("CreateCarEmbed() lost offer details: %#v", embed)
	}
	if len(embed.Fields) != 2 || embed.Fields[0].Value != "30 000 EUR" || embed.Fields[1].Value != "123" {
		t.Fatalf("CreateCarEmbed() fields = %#v", embed.Fields)
	}
	if embed.Footer == nil || embed.Footer.Text != "Result 2 of 4" {
		t.Fatalf("CreateCarEmbed() footer = %#v", embed.Footer)
	}
}

func TestCreateCarEmbedAndFallbackUseSafeMissingValues(t *testing.T) {
	embed := CreateCarEmbed(scraper.Offer{}, 1, 1)
	if embed.Title != "Car listing" || embed.Fields[0].Value != "Price not available" {
		t.Fatalf("CreateCarEmbed() missing-value fallback = %#v", embed)
	}

	fallback := CreateCarFallbackMessage(scraper.Offer{})
	if fallback != "🚗 **Car listing**\n💰 Price not available" {
		t.Fatalf("CreateCarFallbackMessage() = %q", fallback)
	}
}

func TestSearchMessagesDescribeTruncation(t *testing.T) {
	if got := CreateSearchSummaryMessage(12, 10); got != "🎉 **Found 12 car(s)** (showing first 10)" {
		t.Fatalf("summary = %q", got)
	}
	if got := CreateSearchCompleteMessage(10, 12); got != "✅ **Search complete!** Showing 10 results out of 12 total found" {
		t.Fatalf("complete = %q", got)
	}
}
