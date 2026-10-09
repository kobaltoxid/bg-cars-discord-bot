package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractAllOffersUsesLiveCardFallbacksAndDeduplicates(t *testing.T) {
	markup := `<div data-item="42" title="BMW 530d"><a href="/offer/42"><img srcset="/img/low.jpg 1x, /img/high.jpg 2x"></a><span class="card-price">30 000 EUR <small>58 674 BGN</small></span><h3 class="card__title">BMW 530d</h3></div>` +
		`<article data-item="42"><a href="https://www.cars.bg/offer/42">duplicate</a></article>` +
		`<div><a href="/offer/99"><span class="name">Audi A4</span><span class="price">12.500 €</span><div style="background-image:url('/fallback.jpg')"></div></a></div>` +
		`<div data-item=""></div><div class="empty-card"></div>`
	offers, err := ExtractAllOffers(markup)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 2 {
		t.Fatalf("got %d offers, want 2: %#v", len(offers), offers)
	}
	if got, want := offers[0].ListLink, "https://www.cars.bg/offer/42"; got != want {
		t.Errorf("link = %q, want %q", got, want)
	}
	if got, want := offers[0].ImageURL, "https://www.cars.bg/img/low.jpg"; got != want {
		t.Errorf("image = %q, want %q", got, want)
	}
	if got, want := offers[0].Price, "30,000 EUR"; got != want {
		t.Errorf("price = %q, want %q", got, want)
	}
	if got, want := offers[1].ImageURL, "https://www.cars.bg/fallback.jpg"; got != want {
		t.Errorf("background image = %q, want %q", got, want)
	}
}

func TestExtractAllOffersEmptyAndMalformedMarkup(t *testing.T) {
	offers, err := ExtractAllOffers(`<html><body><div class="not-an-offer">nothing</div></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 0 {
		t.Fatalf("got %#v, want no offers", offers)
	}
}

func TestCalcURLAndBrandAliases(t *testing.T) {
	query := calcUrl(BrandNameToID(" Mercedes-Benz "), 2, []string{"m 1", ""})
	if !strings.Contains(query, "brandId=54") || !strings.Contains(query, "page=2") || !strings.Contains(query, "models%5B%5D=m+1") {
		t.Fatalf("unexpected query: %s", query)
	}
	if BrandNameToID("not-a-brand") != BrandUnknown {
		t.Fatal("unknown brand must be safe")
	}
	if BrandNameToID("Volkswagen") != BrandVW {
		t.Fatal("Volkswagen alias not recognized")
	}
	if BrandNameToID("Citroën") != BrandCitroen {
		t.Fatal("Citroën alias not recognized")
	}
	verified := map[string]CarBrand{
		"BMW": BrandBMW, "Audi": BrandAudi, "VW": BrandVW, "Toyota": BrandToyota,
		"Mitsubishi": BrandMitsubishi, "Honda": BrandHonda, "Ford": BrandFord,
		"Opel": BrandOpel, "Renault": BrandRenault, "Mazda": BrandMazda,
		"Citroen": BrandCitroen, "Peugeot": BrandPeugeot, "Nissan": BrandNissan,
		"Skoda": BrandSkoda, "Fiat": BrandFiat, "Hyundai": BrandHyundai,
		"Kia": BrandKia, "Volvo": BrandVolvo, "Suzuki": BrandSuzuki,
	}
	for name, want := range verified {
		if got := BrandNameToID(name); got != want {
			t.Errorf("BrandNameToID(%q) = %d, want %d", name, got, want)
		}
	}
}

func TestSearchCarsRejectsUnsupportedBrand(t *testing.T) {
	_, err := SearchCars(context.Background(), 1, "Tesla", "")
	if err == nil || !strings.Contains(err.Error(), "unsupported car brand") {
		t.Fatalf("error = %v, want unsupported brand error", err)
	}
}

func TestSearchCarsAllowsEmptyBrandWithoutRequestsWhenNoPages(t *testing.T) {
	offers, err := SearchCars(context.Background(), 0, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 0 {
		t.Fatalf("got %d offers, want none", len(offers))
	}
}

func TestModelNameToIDsIncludesVerifiedAliases(t *testing.T) {
	want := map[string]string{
		"a4": "76", "outlander": "1015", "insignia": "1118", "berlingo": "255",
		"mazda3": "763", "mazda6": "766", "350z": "1058", "octavia": "1308",
		"santa-fe": "540", "rio": "608", "v50": "1456", "liana": "1343",
		"corsa": "1114", "c3": "260",
	}
	for model, id := range want {
		ids := ModelNameToIDs(model)
		if len(ids) != 1 || ids[0] != id {
			t.Errorf("ModelNameToIDs(%q) = %#v, want [%q]", model, ids, id)
		}
	}
	if len(ModelNameToIDs("5-series")) == 0 {
		t.Fatal("existing 5-series aliases were lost")
	}
	for model, want := range map[string]string{"mazda-3": "763", "mazda-6": "766", "santa fe": "540"} {
		ids := ModelNameToIDs(model)
		if len(ids) != 1 || ids[0] != want {
			t.Errorf("ModelNameToIDs(%q) = %#v, want [%q]", model, ids, want)
		}
	}
}

func TestOfferLinkPreferenceDerivesDataItemAndPriceFromCardText(t *testing.T) {
	markup := "<div><a href=\"/seller/88\">seller</a><a data-href=\"/offer/987654\">car</a><span class=\"specs\">2018, 9 900 EUR, diesel</span></div>"
	offers, err := ExtractAllOffers(markup)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 1 {
		t.Fatalf("got %d offers, want 1", len(offers))
	}
	if got, want := offers[0].ListLink, "https://www.cars.bg/offer/987654"; got != want {
		t.Errorf("link = %q, want %q", got, want)
	}
	if got, want := offers[0].DataItem, "987654"; got != want {
		t.Errorf("data item = %q, want %q", got, want)
	}
	if got, want := offers[0].Price, "9,900 EUR"; got != want {
		t.Errorf("price = %q, want %q", got, want)
	}
}

func TestExtractAllOffersIgnoresSellerOnlyHref(t *testing.T) {
	offers, err := ExtractAllOffers(`<div><a href="/seller/company-88">Seller Company</a></div>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 0 {
		t.Fatalf("got %#v, want no offers", offers)
	}
}

func TestGetOffersByURLWithClientSetsBrowserHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header is empty")
		}
		if r.Header.Get("Accept") == "" {
			t.Error("Accept header is empty")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	_, _ = GetOffersByURLWithClient(context.Background(), server.URL, server.Client())
}

func TestGetOffersByURLWithClientRejectsNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusServiceUnavailable) }))
	defer server.Close()
	_, err := GetOffersByURLWithClient(context.Background(), server.URL, server.Client())
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("error = %v, want HTTP 503 error", err)
	}
}

func TestParsePricePreservesDecimalThousands(t *testing.T) {
	if got, want := parsePrice("10,559.53 EUR BGN"), "10,559.53 EUR"; got != want {
		t.Fatalf("parsePrice() = %q, want %q", got, want)
	}
}
