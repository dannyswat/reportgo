package parser

import "testing"

func TestParseTemplateResolvesStyleInheritance(t *testing.T) {
	report, err := ParseTemplateFromString(`<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4"/>
    <styles>
        <style name="base">
            <fontFamily>Arial</fontFamily>
            <fontSize>11</fontSize>
            <textColor r="10" g="20" b="30"/>
            <lineHeight>5</lineHeight>
        </style>
        <style name="child" extends="base">
            <fontStyle>B</fontStyle>
            <fontSize>12</fontSize>
        </style>
    </styles>
    <sections>
        <section name="main"/>
    </sections>
</report>`)
	if err != nil {
		t.Fatalf("ParseTemplateFromString returned error: %v", err)
	}

	if len(report.Styles.Styles) != 2 {
		t.Fatalf("expected 2 styles, got %d", len(report.Styles.Styles))
	}

	var childFound bool
	for _, style := range report.Styles.Styles {
		if style.Name != "child" {
			continue
		}
		childFound = true
		if style.FontFamily != "Arial" {
			t.Fatalf("expected inherited font family, got %q", style.FontFamily)
		}
		if style.FontStyle != "B" {
			t.Fatalf("expected overriding font style, got %q", style.FontStyle)
		}
		if style.FontSize != 12 {
			t.Fatalf("expected overriding font size, got %.2f", style.FontSize)
		}
		if style.TextColor == nil || style.TextColor.R != 10 || style.TextColor.G != 20 || style.TextColor.B != 30 {
			t.Fatalf("expected inherited text color, got %#v", style.TextColor)
		}
		if style.LineHeight != 5 {
			t.Fatalf("expected inherited line height, got %.2f", style.LineHeight)
		}
	}

	if !childFound {
		t.Fatalf("expected to find resolved child style")
	}
}

func TestParseTemplateRejectsStyleInheritanceCycles(t *testing.T) {
	_, err := ParseTemplateFromString(`<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4"/>
    <styles>
        <style name="a" extends="b"><fontFamily>Arial</fontFamily></style>
        <style name="b" extends="a"><fontFamily>Arial</fontFamily></style>
    </styles>
    <sections>
        <section name="main"/>
    </sections>
</report>`)
	if err == nil {
		t.Fatalf("expected cycle detection error")
	}
}

func TestParseTemplateParsesRowGridColumns(t *testing.T) {
	report, err := ParseTemplateFromString(`<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4"/>
    <sections>
        <section name="main">
            <rowgrid columns="3" spacingAfter="4">
                <col>
                    <text style="body">First</text>
                    <spacer height="2"/>
                    <text>Second</text>
                </col>
                <col>
                    <row>
                        <text width="20">Nested</text>
                        <text width="20">Row</text>
                    </row>
                </col>
            </rowgrid>
        </section>
    </sections>
</report>`)
	if err != nil {
		t.Fatalf("ParseTemplateFromString returned error: %v", err)
	}

	if len(report.Sections.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(report.Sections.Sections))
	}
	section := report.Sections.Sections[0]
	if len(section.Elements) != 1 {
		t.Fatalf("expected 1 section element, got %d", len(section.Elements))
	}
	if section.Elements[0].Type != "rowgrid" || section.Elements[0].RowGrid == nil {
		t.Fatalf("expected rowgrid element, got %#v", section.Elements[0])
	}
	rowGrid := section.Elements[0].RowGrid
	if rowGrid.Columns != 3 {
		t.Fatalf("expected rowgrid columns to be 3, got %d", rowGrid.Columns)
	}
	if len(rowGrid.Cols) != 2 {
		t.Fatalf("expected 2 rowgrid cols, got %d", len(rowGrid.Cols))
	}
	if len(rowGrid.Cols[0].Elements) != 3 {
		t.Fatalf("expected first col to preserve 3 child elements, got %d", len(rowGrid.Cols[0].Elements))
	}
	if rowGrid.Cols[1].Elements[0].Type != "row" {
		t.Fatalf("expected nested row child in second col, got %s", rowGrid.Cols[1].Elements[0].Type)
	}
}
