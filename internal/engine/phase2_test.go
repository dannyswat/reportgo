package engine

import (
	"strings"
	"testing"

	"github.com/dannyswat/reportgo/internal/models"
)

func TestRenderSpacerAdvancesY(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.pdf.SetY(20)

	engine.renderSpacer(&models.Spacer{Height: 3})

	if got := engine.pdf.GetY(); got != 23 {
		t.Fatalf("expected spacer to advance Y to 23.00, got %.2f", got)
	}
}

func TestRenderSectionPaddingAffectsWrapWidth(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.applyStyle("body")

	content := "pad"
	for {
		width := engine.pdf.GetStringWidth(content)
		if width > 171 && width < 186 {
			break
		}
		if width >= 186 {
			t.Fatalf("failed to build padding test content width in expected range, got %.2f", width)
		}
		content += " pad"
	}

	startY := engine.pdf.GetY()
	section := models.Section{
		Name:        "padded",
		PaddingLeft: 20,
		Elements: []models.SectionElement{{
			Type: "text",
			Text: &models.Text{Style: "body", Wrap: true, Content: content},
		}},
	}

	if err := engine.renderSection(&section); err != nil {
		t.Fatalf("renderSection returned error: %v", err)
	}

	if got := engine.pdf.GetY(); got <= startY+5 {
		t.Fatalf("expected padded wrapped text to span multiple lines, got Y %.2f", got)
	}
}

func TestRenderRowUsesTallestChildHeight(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.applyStyle("body")
	left := "row"
	for {
		width := engine.pdf.GetStringWidth(left)
		if width > 42 && width < 75 {
			break
		}
		if width >= 75 {
			t.Fatalf("failed to build row test content width in expected range, got %.2f", width)
		}
		left += " row"
	}

	row := models.Row{
		BaseElement: models.BaseElement{SpacingAfter: 2},
		Elements: []models.SectionElement{
			{
				Type: "text",
				Text: &models.Text{Style: "body", Width: 40, Content: left},
			},
			{
				Type: "text",
				Text: &models.Text{Style: "body", Width: 30, Align: "R", Content: "right"},
			},
		},
	}

	startY := engine.pdf.GetY()
	if err := engine.renderRow(&row); err != nil {
		t.Fatalf("renderRow returned error: %v", err)
	}

	if got := engine.pdf.GetY(); got != startY+12 {
		t.Fatalf("expected row to advance by tallest child plus spacing to %.2f, got %.2f", startY+12, got)
	}

	output := renderedPDF(t, engine)
	if !strings.Contains(output, "right") {
		t.Fatalf("expected row child content to be rendered")
	}
}

func TestRenderRowAllowsAutoWidthText(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.applyStyle("body")

	row := models.Row{
		Elements: []models.SectionElement{
			{
				Type: "text",
				Text: &models.Text{Style: "body", Content: "left"},
			},
			{
				Type: "text",
				Text: &models.Text{Style: "body", Align: "R", Content: "right"},
			},
		},
	}

	if err := engine.renderRow(&row); err != nil {
		t.Fatalf("renderRow returned error for auto-width text: %v", err)
	}

	if got := engine.pdf.GetY(); got <= 10 {
		t.Fatalf("expected auto-width row text to advance Y, got %.2f", got)
	}
	if got := engine.pdf.GetX(); got != engine.flowLeftMargin() {
		t.Fatalf("expected row to restore X to flow margin, got %.2f", got)
	}
}

func TestRenderRowSupportsRelativeOffsets(t *testing.T) {
	engine := newTestEngine(t, nil)
	engine.applyStyle("body")

	row := models.Row{
		Elements: []models.SectionElement{
			{
				Type: "text",
				Text: &models.Text{Style: "body", Width: 20, Content: "first"},
			},
			{
				Type: "text",
				Text: &models.Text{Style: "body", X: 8, Y: 4, Width: 25, Content: "second"},
			},
		},
	}

	startY := engine.pdf.GetY()
	if err := engine.renderRow(&row); err != nil {
		t.Fatalf("renderRow returned error for row-relative offsets: %v", err)
	}

	if got := engine.pdf.GetY(); got <= startY+4 {
		t.Fatalf("expected row-relative Y offset to affect row height, got %.2f", got)
	}
}

func TestRenderSectionRejectsUnsupportedRowChild(t *testing.T) {
	engine := newTestEngine(t, nil)
	section := models.Section{
		Name: "bad-row",
		Elements: []models.SectionElement{{
			Type: "row",
			Row: &models.Row{
				Elements: []models.SectionElement{{
					Type: "list",
					List: &models.List{Items: "{{.Items}}"},
				}},
			},
		}},
	}

	if err := engine.renderSection(&section); err == nil {
		t.Fatalf("expected unsupported row child to return an error")
	}
}
