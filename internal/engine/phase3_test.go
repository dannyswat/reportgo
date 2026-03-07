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
