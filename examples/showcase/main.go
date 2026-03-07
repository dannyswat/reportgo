package main

import (
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/dannyswat/reportgo/pkg/reportgo"
)

func main() {
	regularFont := mustReadFile("fonts/Noto_Sans/static/NotoSans-Regular.ttf")
	boldFont := mustReadFile("fonts/Noto_Sans/static/NotoSans-Bold.ttf")

	engine := reportgo.New(
		reportgo.WithEmbeddedFont("noto-regular", "Noto Sans", "", regularFont),
		reportgo.WithEmbeddedFont("noto-bold", "Noto Sans", "B", boldFont),
		reportgo.WithFuncMap(template.FuncMap{
			"badge": func(input string) string {
				return "[" + strings.ToUpper(strings.TrimSpace(input)) + "]"
			},
		}),
	)

	if err := engine.LoadTemplate("templates/examples/showcase.xml"); err != nil {
		log.Fatalf("load template: %v", err)
	}

	outputPath := "examples/showcase/showcase.pdf"
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer outputFile.Close()

	data := map[string]interface{}{
		"Theme":            "Enhancement Review",
		"Title":            "ReportGo Enhancement Showcase",
		"GeneratedAt":      "2026-03-07",
		"ExecutiveSummary": "This example exercises the newer layout, template, and API capabilities in a single document so the PDF can be reviewed manually. It demonstrates style inheritance, conditional sections and elements, section loops, padding, spacers, row-based horizontal layout, line spacing, built-in helper functions, custom template functions, and embedded fonts.",
		"Revenue":          184250.75,
		"DeliveryRate":     0.972,
		"PrimarySlug":      "showcase-enhancement-review",
		"Highlights": []string{
			"Rows can mix auto-width and right-aligned text without dropping into absolute page coordinates.",
			"Section padding and spacer elements make document flow templates much easier to author.",
			"Style inheritance removes repetitive font and color declarations across related styles.",
		},
		"Projects": []map[string]interface{}{
			{
				"Visible":     true,
				"Featured":    true,
				"Name":        "signal redesign",
				"LastUpdated": "2026-02-17",
				"Description": "Reworked the weekly operations report into a tighter narrative format with clearer layout hierarchy, more legible metrics, and a cleaner section flow across desktop and print output.",
				"Tags":        []string{"row layout", "spacing", "style inheritance"},
			},
			{
				"Visible":     true,
				"Featured":    false,
				"Name":        "ops snapshot",
				"LastUpdated": "2026-03-01",
				"Description": "Added conditional summaries, loop-driven project cards, and compact overview rows so large data sets can still render as concise PDFs without pre-expanding XML on the caller side.",
				"Tags":        []string{"loops", "conditions", "helpers"},
			},
			{
				"Visible":     false,
				"Featured":    false,
				"Name":        "hidden validation card",
				"LastUpdated": "2026-03-05",
				"Description": "This item should not appear because section-level loop conditions now work correctly.",
				"Tags":        []string{"hidden", "condition"},
			},
		},
		"Metrics": []map[string]interface{}{
			{"label": "Cycle Time", "owner": "Platform", "value": "6.4d", "delta": "-12%"},
			{"label": "Release Confidence", "owner": "Delivery", "value": "94", "delta": "+4"},
			{"label": "Escaped Issues", "owner": "QA", "value": "3", "delta": "-2"},
		},
		"ShowConfidential": true,
		"ConfidentialNote": "This page intentionally uses a conditional section so reviewers can confirm optional narrative blocks render only when enabled in the input data.",
	}

	if err := engine.GenerateToWriter(outputFile, data); err != nil {
		log.Fatalf("generate pdf: %v", err)
	}

	log.Printf("generated %s", outputPath)
}

func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}
	return data
}
