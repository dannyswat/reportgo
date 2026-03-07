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
