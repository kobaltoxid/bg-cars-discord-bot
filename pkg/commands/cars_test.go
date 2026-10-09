package commands

import (
	"strings"
	"testing"
)

func TestParseCarsArgs(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		wantBrand, wantModel string
		wantPages            int
	}{
		{name: "defaults", wantPages: 2},
		{name: "full search", args: []string{"BMW", "X5", "5"}, wantBrand: "BMW", wantModel: "X5", wantPages: 5},
		{name: "brand and pages shorthand", args: []string{"BMW", "5"}, wantBrand: "BMW", wantPages: 5},
		{name: "trimmed values", args: []string{" Audi ", " A4 "}, wantBrand: "Audi", wantModel: "A4", wantPages: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brand, model, pages, err := ParseCarsArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseCarsArgs() error = %v", err)
			}
			if brand != tt.wantBrand || model != tt.wantModel || pages != tt.wantPages {
				t.Fatalf("ParseCarsArgs() = (%q, %q, %d), want (%q, %q, %d)", brand, model, pages, tt.wantBrand, tt.wantModel, tt.wantPages)
			}
		})
	}
}

func TestParseCarsArgsRejectsInvalidPagesAndExtraArguments(t *testing.T) {
	for _, args := range [][]string{{"BMW", "0"}, {"BMW", "11"}, {"BMW", "X5", "0"}, {"BMW", "X5", "11"}, {"BMW", "X5", "many"}, {"BMW", "X5", "5", "extra"}} {
		if _, _, _, err := ParseCarsArgs(args); err == nil {
			t.Errorf("ParseCarsArgs(%q) expected an error", args)
		}
	}
}

func TestCarsCommandInvalidArgumentsIncludesUsage(t *testing.T) {
	got := CarsCommand(nil, "", []string{"BMW", "X5", "0"})
	if got == "" || !containsAll(got, "Invalid !cars arguments", "!cars [brand] [model] [pages]", "1 to 10") {
		t.Fatalf("CarsCommand() = %q, want a helpful validation message", got)
	}
}

func TestCarsCommandRejectsUnsupportedBrandBeforeStartingSearch(t *testing.T) {
	got := CarsCommand(nil, "", []string{"Tesla"})
	if !containsAll(got, "Unsupported brand", "Supported brands: "+supportedBrandList, "!cars [brand] [model] [pages]") {
		t.Fatalf("CarsCommand() = %q, want supported-brand guidance", got)
	}
}

func TestBuildSearchStartMessage(t *testing.T) {
	got := buildSearchStartMessage("bmw", "X5", 3)
	if !containsAll(got, "Brand:** BMW", "Model:** X5", "Max Pages:** 3") {
		t.Fatalf("buildSearchStartMessage() = %q", got)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
