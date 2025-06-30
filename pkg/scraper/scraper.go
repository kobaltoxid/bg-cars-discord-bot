package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// CarBrand represents different car brands
type CarBrand int

const (
	BrandUnknown CarBrand = 0
	BrandBMW     CarBrand = 10
	BrandAudi    CarBrand = 20
	BrandVW      CarBrand = 30
)

// Offer represents a car listing
type Offer struct {
	DataItem string
	Title    string
	ImageURL string
	ListLink string
	Price    string
}

// BrandNameToID maps a string (case-insensitive) to a CarBrand type
func BrandNameToID(brand string) CarBrand {
	switch strings.ToLower(brand) {
	case "bmw":
		return BrandBMW
	case "audi":
		return BrandAudi
	case "vw", "volkswagen":
		return BrandVW
	default:
		return BrandUnknown
	}
}

// ModelNameToIDs maps model names to their IDs
var modelNameToIDs = map[string][]string{
	"5series":  {"1000003", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132"},
	"5-series": {"1000003", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132"},
	"5":        {"1000003", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132"},
}

func ModelNameToIDs(model string) []string {
	return modelNameToIDs[strings.ToLower(model)]
}

// calcUrl builds the search URL
func calcUrl(brand CarBrand, page int, modelIDs []string) string {
	base := "https://www.cars.bg/carslist.php"
	params := url.Values{}
	if brand != 0 {
		params.Set("subm", "1")
		params.Set("add_search", "1")
		params.Set("typeoffer", "1")
		params.Set("brandId", fmt.Sprintf("%d", brand))
	}
	params.Set("page", fmt.Sprintf("%d", page))
	for _, mid := range modelIDs {
		params.Add("models[]", mid)
	}
	return base + "?" + params.Encode()
}

// SearchCars performs the car search and returns results
func SearchCars(ctx context.Context, maxPages int, brand string, model string) ([]Offer, error) {
	var allOffers []Offer

	carBrandId := BrandNameToID(brand)
	modelIDs := ModelNameToIDs(model)

	for page := 1; page <= maxPages; page++ {
		url := calcUrl(carBrandId, page, modelIDs)
		fmt.Printf("Scraping page %d: %s\n", page, url)

		offersOnPage, err := GetOffersByUrl(ctx, url)
		if err != nil {
			fmt.Printf("Error parsing offers on page %d: %v\n", page, err)
			continue
		}
		allOffers = append(allOffers, offersOnPage...)

		// Stop if no offers found (likely reached end)
		if len(offersOnPage) == 0 {
			fmt.Printf("No offers found on page %d, stopping search\n", page)
			break
		}
	}

	return allOffers, nil
}

// GetOffersByUrl fetches and parses offers from a URL
func GetOffersByUrl(ctx context.Context, url string) ([]Offer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ExtractAllOffers(string(body))
}

// ExtractAllOffers parses HTML and extracts car offers
func ExtractAllOffers(htmlStr string) ([]Offer, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, err
	}

	var offers []Offer

	var findOffers func(*html.Node)
	findOffers = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			var dataItem, title string
			for _, attr := range n.Attr {
				if attr.Key == "data-item" {
					dataItem = attr.Val
				}
				if attr.Key == "title" {
					title = attr.Val
				}
			}
			if dataItem != "" {
				offer := Offer{
					DataItem: dataItem,
					Title:    title,
					ImageURL: findImageURL(n),
					ListLink: findListLink(n),
					Price:    findPrice(n),
				}
				offers = append(offers, offer)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findOffers(c)
		}
	}
	findOffers(doc)
	return offers, nil
}

// parsePrice cleans and formats price text properly
func parsePrice(rawPrice string) string {
	// Clean up basic formatting
	price := strings.TrimSpace(rawPrice)
	price = strings.ReplaceAll(price, "\n", " ")
	price = strings.ReplaceAll(price, "\t", " ")

	// Replace multiple spaces with single space
	for strings.Contains(price, "  ") {
		price = strings.ReplaceAll(price, "  ", " ")
	}

	if price == "" {
		return "Price not available"
	}

	// Use regex to extract price amounts and currencies
	priceRegex := regexp.MustCompile(`(\d+(?:[,\s]\d{3})*(?:\.\d{2})?)\s*(BGN|EUR|лв\.?)`)
	matches := priceRegex.FindAllStringSubmatch(price, -1)

	if len(matches) > 0 {
		// Prefer BGN price if available, otherwise use the first match
		for _, match := range matches {
			if len(match) >= 3 {
				amount := match[1]
				currency := match[2]

				// Clean up amount formatting
				amount = strings.ReplaceAll(amount, " ", ",")

				// Prefer BGN or лв (Bulgarian Lev)
				if currency == "BGN" || strings.Contains(currency, "лв") {
					if strings.Contains(currency, "лв") {
						currency = "BGN"
					}
					return fmt.Sprintf("%s %s", amount, currency)
				}
			}
		}

		// If no BGN found, return the first valid price
		if len(matches[0]) >= 3 {
			amount := matches[0][1]
			currency := matches[0][2]
			amount = strings.ReplaceAll(amount, " ", ",")
			return fmt.Sprintf("%s %s", amount, currency)
		}
	}

	// Fallback: try to extract just numbers and common currency indicators
	numberRegex := regexp.MustCompile(`(\d+(?:[,\s]\d{3})*(?:\.\d{2})?)`)
	currencyRegex := regexp.MustCompile(`(BGN|EUR|лв\.?)`)

	numberMatches := numberRegex.FindAllString(price, -1)
	currencyMatches := currencyRegex.FindAllString(price, -1)

	if len(numberMatches) > 0 && len(currencyMatches) > 0 {
		amount := numberMatches[0]
		currency := currencyMatches[0]

		// Clean up amount
		amount = strings.ReplaceAll(amount, " ", ",")

		// Normalize currency
		if strings.Contains(currency, "лв") {
			currency = "BGN"
		}

		return fmt.Sprintf("%s %s", amount, currency)
	}

	// Final fallback: return cleaned original text
	return price
}

// Helper functions for parsing HTML elements
func findPrice(n *html.Node) string {
	var price string
	var f func(*html.Node)
	f = func(nn *html.Node) {
		if nn.Type == html.ElementNode && nn.Data == "h6" {
			for _, attr := range nn.Attr {
				if attr.Key == "class" &&
					strings.Contains(attr.Val, "card__title") &&
					strings.Contains(attr.Val, "mdc-typography") &&
					strings.Contains(attr.Val, "mdc-typography--headline6") &&
					strings.Contains(attr.Val, "price") {
					price = getTextContent(nn)
					return
				}
			}
		}
		for c := nn.FirstChild; c != nil; c = c.NextSibling {
			if price == "" {
				f(c)
			}
		}
	}
	f(n)
	// Parse and format the price properly
	return parsePrice(price)
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}

func findImageURL(n *html.Node) string {
	var imgURL string
	var f func(*html.Node)
	f = func(nn *html.Node) {
		if nn.Type == html.ElementNode && nn.Data == "div" {
			for _, attr := range nn.Attr {
				if attr.Key == "style" && strings.Contains(attr.Val, "background-image") {
					re := regexp.MustCompile(`url\(['"]?([^'")]+)['"]?\)`)
					matches := re.FindStringSubmatch(attr.Val)
					if len(matches) > 1 {
						imgURL = matches[1]
						return
					}
				}
			}
		}
		for c := nn.FirstChild; c != nil; c = c.NextSibling {
			if imgURL == "" {
				f(c)
			}
		}
	}
	f(n)
	return imgURL
}

func findListLink(n *html.Node) string {
	var link string
	var f func(*html.Node)
	f = func(nn *html.Node) {
		if nn.Type == html.ElementNode && nn.Data == "a" {
			for _, attr := range nn.Attr {
				if attr.Key == "list-link" {
					link = attr.Val
					return
				}
			}
		}
		for c := nn.FirstChild; c != nil; c = c.NextSibling {
			if link == "" {
				f(c)
			}
		}
	}
	f(n)
	return link
}
