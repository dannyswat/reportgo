// Package models defines the style structures for ReportGo templates.
package models

import "fmt"

// Fonts contains font definitions.
type Fonts struct {
	Fonts []Font `xml:"font"`
}

// Font represents a custom font definition.
type Font struct {
	Name   string `xml:"name,attr"`
	Family string `xml:"family,attr"`
	Style  string `xml:"style,attr"`
	File   string `xml:"file,attr"`
}

// Styles contains style definitions.
type Styles struct {
	Styles []Style `xml:"style"`
}

// Style represents a reusable style definition.
type Style struct {
	Name       string    `xml:"name,attr"`
	FontFamily string    `xml:"fontFamily"`
	FontStyle  string    `xml:"fontStyle"`
	FontSize   float64   `xml:"fontSize"`
	TextColor  *RGBColor `xml:"textColor"`
	FillColor  *RGBColor `xml:"fillColor"`
	Align      string    `xml:"align"`
	LineHeight float64   `xml:"lineHeight"`
}

// RGBColor represents an RGB color value.
type RGBColor struct {
	R int `xml:"r,attr"`
	G int `xml:"g,attr"`
	B int `xml:"b,attr"`
}

// ToRGB returns the color as R, G, B values.
func (c *RGBColor) ToRGB() (int, int, int) {
	if c == nil {
		return 0, 0, 0
	}
	return c.R, c.G, c.B
}

// ParseColor parses a color string (hex or named) to RGB values.
func ParseColor(color string) (int, int, int) {
	namedColors := map[string][3]int{
		"black":     {0, 0, 0},
		"white":     {255, 255, 255},
		"red":       {255, 0, 0},
		"green":     {0, 128, 0},
		"blue":      {0, 0, 255},
		"gray":      {128, 128, 128},
		"lightgray": {211, 211, 211},
		"darkgray":  {169, 169, 169},
		"yellow":    {255, 255, 0},
		"orange":    {255, 165, 0},
		"purple":    {128, 0, 128},
		"cyan":      {0, 255, 255},
		"magenta":   {255, 0, 255},
	}

	if rgb, ok := namedColors[color]; ok {
		return rgb[0], rgb[1], rgb[2]
	}

	if len(color) == 7 && color[0] == '#' {
		var r, g, b int
		fmt.Sscanf(color, "#%02x%02x%02x", &r, &g, &b)
		return r, g, b
	}

	return 0, 0, 0
}
