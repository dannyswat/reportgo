package engine

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/dannyswat/reportgo/internal/models"
)

func TestAddFuncMapExtendsTemplates(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.AddFuncMap(template.FuncMap{
		"surround": func(input string) string { return "[" + input + "]" },
	})

	got := engine.processTemplate("{{surround \"ok\"}}")
	if got != "[ok]" {
		t.Fatalf("expected custom function result, got %q", got)
	}
}

func TestDefaultFuncMapProvidesFormattingHelpers(t *testing.T) {
	engine := newTestEngine(t, map[string]interface{}{
		"Date":   "2026-03-07",
		"Amount": 12.5,
		"Rate":   0.125,
	})

	if got := engine.processTemplate("{{formatDate .Date \"Jan 2006\"}}"); got != "Mar 2026" {
		t.Fatalf("expected formatted date, got %q", got)
	}
	if got := engine.processTemplate("{{formatCurrency .Amount \"EUR \"}}"); got != "EUR 12.50" {
		t.Fatalf("expected formatted currency, got %q", got)
	}
	if got := engine.processTemplate("{{formatPercent .Rate 1}}"); got != "12.5%" {
		t.Fatalf("expected formatted percent, got %q", got)
	}
	if got := engine.processTemplate("{{ifelse true \"yes\" \"no\"}}"); got != "yes" {
		t.Fatalf("expected ifelse result, got %q", got)
	}
	if got := engine.processTemplate("{{truncate \"abcdef\" 5}}"); got != "ab..." {
		t.Fatalf("expected truncated value, got %q", got)
	}
	if got := engine.processTemplate("{{replace \"a-b\" \"-\" \"_\"}}"); got != "a_b" {
		t.Fatalf("expected replaced value, got %q", got)
	}
	engine.SetData(map[string]interface{}{"Parts": []string{"a", "b"}})
	if got := engine.processTemplate("{{join .Parts \",\"}}"); got != "a,b" {
		t.Fatalf("expected joined value, got %q", got)
	}
}

func TestFormattingHelpersSupportGroupingAndDecimals(t *testing.T) {
	engine := newTestEngine(t, map[string]interface{}{
		"Big":    1234567.891,
		"Amount": 1234.5,
		"Rate":   0.12345,
	})

	if got := engine.processTemplate("{{formatNumber .Big 2}}"); got != "1234567.89" {
		t.Fatalf("expected ungrouped number, got %q", got)
	}
	if got := engine.processTemplate("{{formatNumber .Big 2 true}}"); got != "1,234,567.89" {
		t.Fatalf("expected grouped number, got %q", got)
	}
	if got := engine.processTemplate("{{formatNumber .Big 0 true}}"); got != "1,234,568" {
		t.Fatalf("expected grouped integer, got %q", got)
	}
	if got := engine.processTemplate("{{formatCurrency .Amount \"\" 0 true}}"); got != "$1,234" {
		t.Fatalf("expected grouped currency with explicit decimals, got %q", got)
	}
	if got := engine.processTemplate("{{formatCurrency .Amount \"EUR \" 2 true}}"); got != "EUR 1,234.50" {
		t.Fatalf("expected grouped currency, got %q", got)
	}
	if got := engine.processTemplate("{{formatPercent .Rate 3 true}}"); got != "12.345%" {
		t.Fatalf("expected percent with explicit decimals, got %q", got)
	}
}

func TestFormatValueSupportsExplicitDecimalsAndGrouping(t *testing.T) {
	zeroDecimals := 0
	threeDecimals := 3

	tests := []struct {
		name string
		val  interface{}
		col  models.Column
		want string
	}{
		{"auto currency", 1234.5, models.Column{Format: "currency"}, "$1234.50"},
		{"grouped currency", 1234.5, models.Column{Format: "currency", Grouping: true}, "$1,234.50"},
		{"explicit decimals grouped currency", 1234.567, models.Column{Format: "currency", Decimals: &threeDecimals, Grouping: true}, "$1,234.567"},
		{"explicit zero decimals grouped number", 1234567.4, models.Column{Format: "number", Decimals: &zeroDecimals, Grouping: true}, "1,234,567"},
		{"auto whole number", 1000.0, models.Column{Format: "number"}, "1000"},
		{"grouped percent with explicit decimals", 0.12345, models.Column{Format: "percent", Decimals: &threeDecimals, Grouping: true}, "12.345%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatValue(tt.val, tt.col); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestEmbeddedFontsAreRegistered(t *testing.T) {
	fontData := mustReadFontFile(t)
	engine := newTestEngine(t, nil)
	engine.AddEmbeddedFont(models.EmbeddedFont{
		Name:   "noto-embedded",
		Family: "Noto Sans",
		Style:  "",
		Data:   fontData,
	})
	engine.SetReport(&models.Report{
		Document: models.Document{
			Orientation: "portrait",
			Unit:        "mm",
			Format:      "A4",
			Margins:     &models.Margins{Top: 10, Right: 10, Bottom: 10, Left: 10},
		},
		Styles: &models.Styles{Styles: []models.Style{{
			Name:       "body",
			FontFamily: "Noto Sans",
			FontSize:   12,
			LineHeight: 5,
		}}},
	})
	engine.initPDF()
	engine.pdf.AddPage()

	engine.renderText(&models.Text{Style: "body", Content: "embedded font text"})

	output := renderedPDF(t, engine)
	if output == "" {
		t.Fatalf("expected generated PDF output")
	}
}

func mustReadFontFile(t *testing.T) []byte {
	t.Helper()

	fontPath := filepath.Join("..", "..", "fonts", "Noto_Sans", "static", "NotoSans-Regular.ttf")
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("failed to read test font file %s: %v", fontPath, err)
	}

	return fontData
}
