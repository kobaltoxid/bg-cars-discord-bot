package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// CarBrand is a cars.bg brand identifier. Unknown brands deliberately remain zero.
type CarBrand int

const (
	BrandUnknown    CarBrand = 0
	BrandBMW        CarBrand = 10
	BrandAudi       CarBrand = 8
	BrandVW         CarBrand = 86
	BrandMercedes   CarBrand = 54
	BrandToyota     CarBrand = 80
	BrandMitsubishi CarBrand = 57
	BrandHonda      CarBrand = 31
	BrandFord       CarBrand = 26
	BrandOpel       CarBrand = 63
	BrandRenault    CarBrand = 69
	BrandMazda      CarBrand = 53
	BrandCitroen    CarBrand = 17
	BrandPeugeot    CarBrand = 64
	BrandNissan     CarBrand = 60
	BrandSkoda      CarBrand = 73
	BrandFiat       CarBrand = 25
	BrandHyundai    CarBrand = 33
	BrandKia        CarBrand = 38
	BrandVolvo      CarBrand = 85
	BrandSuzuki     CarBrand = 77
)

// Offer represents a car listing.
type Offer struct {
	DataItem string
	Title    string
	ImageURL string
	ListLink string
	Price    string
}

var brandIDs = map[string]CarBrand{
	"bmw": BrandBMW, "bayerische motoren werke": BrandBMW,
	"audi": BrandAudi,
	"vw":   BrandVW, "volkswagen": BrandVW, "volks wagen": BrandVW,
	"mercedes": BrandMercedes, "mercedes-benz": BrandMercedes, "mercedes benz": BrandMercedes,
	"toyota": BrandToyota, "mitsubishi": BrandMitsubishi, "honda": BrandHonda,
	"ford": BrandFord, "opel": BrandOpel, "renault": BrandRenault,
	"mazda": BrandMazda, "citroen": BrandCitroen, "citroën": BrandCitroen, "peugeot": BrandPeugeot,
	"nissan": BrandNissan, "skoda": BrandSkoda, "škoda": BrandSkoda,
	"fiat": BrandFiat, "hyundai": BrandHyundai, "kia": BrandKia,
	"volvo": BrandVolvo, "suzuki": BrandSuzuki,
}

func BrandNameToID(brand string) CarBrand { return brandIDs[strings.ToLower(strings.TrimSpace(brand))] }

var modelNameToIDs = map[string][]string{
	"5series":   {"1000003", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132"},
	"5-series":  {"1000003", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132"},
	"5":         {"1000003", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132"},
	"a4":        {"76"},
	"outlander": {"1015"},
	"insignia":  {"1118"},
	"berlingo":  {"255"},
	"mazda3":    {"763"},
	"mazda-3":   {"763"},
	"mazda6":    {"766"},
	"mazda-6":   {"766"},
	"350z":      {"1058"},
	"octavia":   {"1308"},
	"santa-fe":  {"540"},
	"santa fe":  {"540"},
	"rio":       {"608"},
	"v50":       {"1456"},
	"liana":     {"1343"},
	"corsa":     {"1114"},
	"c3":        {"260"},
}

func ModelNameToIDs(model string) []string {
	return modelNameToIDs[strings.ToLower(strings.TrimSpace(model))]
}

func calcUrl(brand CarBrand, page int, modelIDs []string) string {
	params := url.Values{}
	if brand != BrandUnknown {
		params.Set("subm", "1")
		params.Set("add_search", "1")
		params.Set("typeoffer", "1")
		params.Set("brandId", strconv.Itoa(int(brand)))
	}
	params.Set("page", strconv.Itoa(page))
	for _, modelID := range modelIDs {
		if strings.TrimSpace(modelID) != "" {
			params.Add("models[]", modelID)
		}
	}
	return "https://www.cars.bg/carslist.php?" + params.Encode()
}

func SearchCars(ctx context.Context, maxPages int, brand string, model string) ([]Offer, error) {
	brand = strings.TrimSpace(brand)
	brandID := BrandNameToID(brand)
	if brand != "" && brandID == BrandUnknown {
		return nil, fmt.Errorf("unsupported car brand %q", brand)
	}
	if maxPages <= 0 {
		return []Offer{}, nil
	}
	var all []Offer
	for page := 1; page <= maxPages; page++ {
		offers, err := GetOffersByUrl(ctx, calcUrl(brandID, page, ModelNameToIDs(model)))
		if err != nil {
			return nil, fmt.Errorf("scrape page %d: %w", page, err)
		}
		all = append(all, offers...)
		if len(offers) == 0 {
			break
		}
	}
	return all, nil
}

func GetOffersByUrl(ctx context.Context, target string) ([]Offer, error) {
	return GetOffersByURLWithClient(ctx, target, &http.Client{Timeout: 10 * time.Second})
}

// GetOffersByURLWithClient is the injectable form of GetOffersByUrl.
func GetOffersByURLWithClient(ctx context.Context, target string, client *http.Client) ([]Offer, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "bg-cars-discord-bot/1.0 (+https://www.cars.bg/)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("cars.bg returned HTTP %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ExtractAllOffers(string(body))
}

func ExtractAllOffers(htmlStr string) ([]Offer, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, err
	}
	var offers []Offer
	seen := map[string]bool{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "div" || n.Data == "article" || n.Data == "li") {
			if item := parseOffer(n); item != nil {
				key := item.ListLink
				if key == "" {
					key = item.DataItem
				}
				if key != "" && !seen[key] {
					seen[key] = true
					offers = append(offers, *item)
					return
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return offers, nil
}

func parseOffer(n *html.Node) *Offer {
	dataItem, link := attr(n, "data-item"), findListLink(n)
	if dataItem == "" {
		dataItem = offerIDFromURL(link)
	}
	if dataItem == "" && link == "" {
		return nil
	}
	return &Offer{DataItem: dataItem, Title: findTitle(n), ImageURL: findImageURL(n), ListLink: link, Price: findPrice(n)}
}

func findTitle(n *html.Node) string {
	if title := strings.TrimSpace(attr(n, "title")); title != "" {
		return title
	}
	var best string
	walkElements(n, func(node *html.Node) {
		class := strings.ToLower(attr(node, "class"))
		if best == "" && (hasClass(class, "title") || hasClass(class, "name") || hasClass(class, "card__title")) {
			text := cleanText(getTextContent(node))
			if text != "" && !looksLikePrice(text) {
				best = text
			}
		}
	})
	return best
}

func findPrice(n *html.Node) string {
	var candidates []string
	walkElements(n, func(node *html.Node) {
		class := strings.ToLower(attr(node, "class"))
		if classContains(class, "price") || attr(node, "data-price") != "" {
			text := cleanText(attr(node, "data-price") + " " + getTextContent(node))
			if text != "" {
				candidates = append(candidates, text)
			}
		}
	})
	for _, candidate := range candidates {
		if price := parsePrice(candidate); price != "Price not available" {
			return price
		}
	}
	if price := parsePrice(getTextContent(n)); price != "Price not available" {
		return price
	}
	return "Price not available"
}

var pricePattern = regexp.MustCompile(`(?i)([0-9]+(?:[.,][0-9]{3})*(?:[.,][0-9]{1,2})?|[0-9]+(?:[ .][0-9]{3})+(?:[.,][0-9]{1,2})?)\s*(EUR|€|BGN|лв\.?|лева?)`)
var backgroundURLPattern = regexp.MustCompile(`(?i)background-image\s*:\s*url\(\s*['"]?([^'")\s]+)['"]?\s*\)`)

func parsePrice(raw string) string {
	matches := pricePattern.FindAllStringSubmatch(cleanText(raw), -1)
	if len(matches) == 0 {
		return "Price not available"
	}
	chosen := matches[0]
	for _, match := range matches {
		if strings.EqualFold(match[2], "EUR") || match[2] == "€" {
			chosen = match
			break
		}
	}
	amount := normalizeAmount(chosen[1])
	currency := strings.ToUpper(chosen[2])
	if currency == "€" {
		currency = "EUR"
	}
	if strings.HasPrefix(strings.ToLower(currency), "лв") || strings.EqualFold(currency, "лева") {
		currency = "BGN"
	}
	return amount + " " + currency
}

func normalizeAmount(value string) string {
	value = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), "\u00a0", ""), " ", "")
	if strings.Contains(value, ",") && strings.Contains(value, ".") {
		if dot := strings.LastIndex(value, "."); len(value)-dot-1 == 2 {
			integer := strings.ReplaceAll(value[:dot], ",", "")
			return formatThousands(integer) + "." + value[dot+1:]
		}
		if comma := strings.LastIndex(value, ","); len(value)-comma-1 == 2 {
			integer := strings.ReplaceAll(value[:comma], ".", "")
			return formatThousands(integer) + "." + value[comma+1:]
		}
	}
	if strings.Contains(value, ",") && strings.Contains(value, ".") {
		if strings.LastIndex(value, ",") > strings.LastIndex(value, ".") {
			value = strings.ReplaceAll(value, ".", "")
			value = strings.ReplaceAll(value, ",", ".")
		} else {
			value = strings.ReplaceAll(value, ",", "")
		}
	} else if strings.Count(value, ",") == 1 && len(value)-strings.LastIndex(value, ",")-1 <= 2 {
		value = strings.ReplaceAll(value, ",", ".")
	} else if strings.Count(value, ".") == 1 && len(value)-strings.LastIndex(value, ".")-1 == 3 {
		// cars.bg commonly uses a dot as the thousands separator.
		value = strings.ReplaceAll(value, ".", "")
	} else {
		value = strings.ReplaceAll(value, ",", "")
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil && f == float64(int64(f)) {
		return formatThousands(strconv.FormatInt(int64(f), 10))
	}
	return strings.ReplaceAll(value, ".", ",")
}

func formatThousands(value string) string {
	for i := len(value) - 3; i > 0; i -= 3 {
		value = value[:i] + "," + value[i:]
	}
	return value
}

func findImageURL(n *html.Node) string {
	var result string
	walkElements(n, func(node *html.Node) {
		if result != "" {
			return
		}
		if node.Data == "img" {
			result = firstNonEmpty(attr(node, "src"), firstSrcset(attr(node, "srcset")))
		}
		if result == "" {
			result = backgroundURL(attr(node, "style"))
		}
	})
	return absoluteURL(result)
}

func findListLink(n *html.Node) string {
	var offerLink, explicitFallback string
	walkElements(n, func(node *html.Node) {
		if node.Data != "a" && attr(node, "list-link") == "" && attr(node, "href") == "" && attr(node, "data-href") == "" {
			return
		}
		for _, candidate := range []struct {
			raw      string
			explicit bool
		}{
			{attr(node, "list-link"), true},
			{attr(node, "href"), false},
			{attr(node, "data-href"), true},
		} {
			rawLink := candidate.raw
			link := absoluteURL(rawLink)
			if link == "" {
				continue
			}
			if strings.Contains(strings.ToLower(linkPath(link)), "/offer/") {
				if offerLink == "" {
					offerLink = link
				}
			} else if candidate.explicit && explicitFallback == "" {
				explicitFallback = link
			}
		}
	})
	if offerLink != "" {
		return offerLink
	}
	return explicitFallback
}

func linkPath(link string) string {
	parsed, err := url.Parse(link)
	if err != nil {
		return link
	}
	return parsed.Path
}

func offerIDFromURL(link string) string {
	parsed, err := url.Parse(link)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i := 0; i+1 < len(parts); i++ {
		if strings.EqualFold(parts[i], "offer") && parts[i+1] != "" {
			return parts[i+1]
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
func firstSrcset(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return strings.Split(fields[0], ",")[0]
}
func backgroundURL(style string) string {
	match := backgroundURLPattern.FindStringSubmatch(style)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
func absoluteURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	base, _ := url.Parse("https://www.cars.bg/")
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	return base.ResolveReference(parsed).String()
}
func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}
func hasClass(classes, wanted string) bool {
	for _, class := range strings.Fields(classes) {
		if class == wanted {
			return true
		}
	}
	return false
}
func classContains(classes, wanted string) bool {
	for _, class := range strings.Fields(classes) {
		if strings.Contains(class, wanted) {
			return true
		}
	}
	return false
}
func cleanText(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(value, "\u00a0", " ")), " ")
}
func looksLikePrice(value string) bool { return pricePattern.MatchString(value) }
func walkElements(n *html.Node, fn func(*html.Node)) {
	if n.Type == html.ElementNode {
		fn(n)
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		walkElements(child, fn)
	}
}
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		b.WriteString(getTextContent(child))
	}
	return b.String()
}
