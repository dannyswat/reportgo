package engine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dannyswat/reportgo/internal/models"
)

func TestRenderSectionSkipsSectionCondition(t *testing.T) {
	engine := newTestEngine(t, map[string]interface{}{"ShowSection": false})
	section := models.Section{
		Name:      "conditional",
		Condition: "{{.ShowSection}}",
		Elements: []models.SectionElement{
			{
				Type: "text",
				Text: &models.Text{Style: "body", Content: "hidden section"},
			},
		},
	}

	startY := engine.pdf.GetY()
	if err := engine.renderSection(&section); err != nil {
		t.Fatalf("renderSection returned error: %v", err)
	}

	if got := engine.pdf.GetY(); got != startY {
		t.Fatalf("expected section condition to skip rendering and keep Y at %.2f, got %.2f", startY, got)
	}

	if output := renderedPDF(t, engine); strings.Contains(output, "hidden section") {
		t.Fatalf("expected skipped section text to be absent from output")
	}
}

func TestRenderSectionSkipsConditionalElement(t *testing.T) {
	engine := newTestEngine(t, map[string]interface{}{"ShowHidden": false})
	section := models.Section{
		Name: "elements",
		Elements: []models.SectionElement{
			{
				Type: "text",
				Text: &models.Text{Style: "body", Content: "visible text"},
			},
			{
				Type: "text",
				Text: &models.Text{BaseElement: models.BaseElement{Condition: "{{.ShowHidden}}"}, Style: "body", Content: "hidden text"},
			},
		},
	}

	if err := engine.renderSection(&section); err != nil {
		t.Fatalf("renderSection returned error: %v", err)
	}

	output := renderedPDF(t, engine)
	if !strings.Contains(output, "visible text") {
		t.Fatalf("expected visible text to be rendered")
	}
	if strings.Contains(output, "hidden text") {
		t.Fatalf("expected conditional element text to be skipped")
	}
}

func TestRenderSectionExpandsLoops(t *testing.T) {
	engine := newTestEngine(t, map[string]interface{}{
		"Products": []map[string]interface{}{
			{"Name": "One"},
			{"Name": "Two"},
		},
	})
	section := models.Section{
		Name:         "products",
		Loop:         "{{.Products}}",
		LoopVariable: "product",
		Elements: []models.SectionElement{
			{
				Type: "text",
				Text: &models.Text{Style: "body", Content: "{{.product.Name}}"},
			},
		},
	}

	if err := engine.renderSection(&section); err != nil {
		t.Fatalf("renderSection returned error: %v", err)
	}

	output := renderedPDF(t, engine)
	if !strings.Contains(output, "One") || !strings.Contains(output, "Two") {
		t.Fatalf("expected looped content for both items, got output %q", output)
	}
}

func TestRenderLineAppliesSpacingAfter(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.pdf.SetY(20)

	engine.renderLine(&models.Line{
		BaseElement: models.BaseElement{SpacingAfter: 4},
		X1:          10,
		Y1:          0,
		X2:          100,
		Y2:          0,
	})

	if got := engine.pdf.GetY(); got != 24 {
		t.Fatalf("expected Y to advance to 24.00 after line spacing, got %.2f", got)
	}
}

func TestRenderTextWrapUsesCurrentX(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.applyStyle("body")

	content := "wrap"
	for {
		width := engine.pdf.GetStringWidth(content)
		if width > 110 && width < 180 {
			break
		}
		if width >= 180 {
			t.Fatalf("failed to build test content width in expected range, got %.2f", width)
		}
		content += " wrap"
	}

	startY := 20.0
	engine.renderText(&models.Text{
		Style:   "body",
		X:       100,
		Y:       startY,
		Wrap:    true,
		Content: content,
	})

	if got := engine.pdf.GetY(); got <= startY+5 {
		t.Fatalf("expected wrapped text to advance beyond one line from X offset, got Y %.2f", got)
	}
}

func TestRenderHeaderFooterElementsOffsetsFooterY(t *testing.T) {
	engine := newTestEngine(t, map[string]interface{}{"DocumentRef": "REF-1"})
	_, pageHeight := engine.pdf.GetPageSize()
	footerTop := pageHeight - 15
	engine.pdf.SetY(footerTop)

	engine.renderHeaderFooterElements(footerTop, []models.Text{{
		Style:   "body",
		X:       15,
		Y:       4,
		Content: "{{.DocumentRef}}",
	}}, nil, nil)

	if got := engine.pdf.GetY(); got <= footerTop {
		t.Fatalf("expected footer text to render within footer area, got Y %.2f with footer top %.2f", got, footerTop)
	}
	if got := engine.pdf.GetY(); got >= pageHeight {
		t.Fatalf("expected footer text to stay on page, got Y %.2f with page height %.2f", got, pageHeight)
	}
	if _, exists := engine.data["PageNumber"]; exists {
		t.Fatalf("expected page number variables to be scoped to header/footer rendering")
	}
}

func TestSetupHeaderFooterRendersPageNumbersInFooter(t *testing.T) {
	engine := New()
	engine.SetReport(&models.Report{
		Document: models.Document{
			Orientation: "portrait",
			Unit:        "mm",
			Format:      "A4",
			Margins:     &models.Margins{Top: 10, Right: 10, Bottom: 10, Left: 10},
		},
		Styles: &models.Styles{Styles: []models.Style{{
			Name:       "body",
			FontFamily: "Arial",
			FontSize:   12,
			LineHeight: 5,
		}}},
		Footer: &models.Footer{
			Enabled: true,
			Height:  15,
			Texts: []models.Text{{
				Style:   "body",
				X:       15,
				Y:       4,
				Content: "Page {{.PageNumber}} of {{.TotalPages}}",
			}},
		},
	})
	engine.SetData(nil)
	engine.initPDF()
	engine.pdf.SetCompression(false)
	engine.setupHeaderFooter()
	engine.pdf.AddPage()

	output := renderedPDF(t, engine)
	if !strings.Contains(output, "Page 1 of 1") {
		t.Fatalf("expected footer page numbers in output, got %q", output)
	}
}

func newTestEngine(t *testing.T, data map[string]interface{}) *Engine {
	t.Helper()

	engine := New()
	engine.SetReport(&models.Report{
		Document: models.Document{
			Orientation: "portrait",
			Unit:        "mm",
			Format:      "A4",
			Margins: &models.Margins{
				Top:    10,
				Right:  10,
				Bottom: 10,
				Left:   10,
			},
		},
		Styles: &models.Styles{
			Styles: []models.Style{{
				Name:       "body",
				FontFamily: "Arial",
				FontSize:   12,
				LineHeight: 5,
			}},
		},
	})
	engine.SetData(data)
	engine.initPDF()
	engine.pdf.SetCompression(false)
	engine.pdf.AddPage()

	return engine
}

func renderedPDF(t *testing.T, engine *Engine) string {
	t.Helper()

	var output bytes.Buffer
	if err := engine.pdf.Output(&output); err != nil {
		t.Fatalf("pdf output failed: %v", err)
	}

	return output.String()
}
