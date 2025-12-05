// Package models defines the element structures for ReportGo templates.
package models

// Element is the interface for all renderable elements.
type Element interface {
	GetType() string
	GetCondition() string
}

// BaseElement contains common element attributes.
type BaseElement struct {
	Condition    string  `xml:"condition,attr"`
	SpacingAfter float64 `xml:"spacingAfter,attr"`
}

// Text represents a text element.
type Text struct {
	BaseElement
	Content string  `xml:",chardata"`
	Style   string  `xml:"style,attr"`
	X       float64 `xml:"x,attr"`
	Y       float64 `xml:"y,attr"`
	Width   float64 `xml:"width,attr"`
	Align   string  `xml:"align,attr"`
	Wrap    bool    `xml:"wrap,attr"`
}

// GetType returns the element type.
func (t Text) GetType() string { return "text" }

// GetCondition returns the condition for rendering.
func (t Text) GetCondition() string { return t.Condition }

// Image represents an image element.
type Image struct {
	BaseElement
	Path   string  `xml:"path,attr"`
	X      float64 `xml:"x,attr"`
	Y      float64 `xml:"y,attr"`
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
	Align  string  `xml:"align,attr"`
}

// GetType returns the element type.
func (i Image) GetType() string { return "image" }

// GetCondition returns the condition for rendering.
func (i Image) GetCondition() string { return i.Condition }

// Table represents a table element.
type Table struct {
	BaseElement
	DataSource        string    `xml:"dataSource,attr"`
	HeaderStyle       string    `xml:"headerStyle,attr"`
	CellStyle         string    `xml:"cellStyle,attr"`
	Border            bool      `xml:"border,attr"`
	AlternateRowColor *RGBColor `xml:"alternateRowColor"`
	Columns           Columns   `xml:"columns"`
}

// GetType returns the element type.
func (t Table) GetType() string { return "table" }

// GetCondition returns the condition for rendering.
func (t Table) GetCondition() string { return t.Condition }

// Columns contains table column definitions.
type Columns struct {
	Columns []Column `xml:"column"`
}

// Column represents a table column definition.
type Column struct {
	Header string  `xml:"header,attr"`
	Field  string  `xml:"field,attr"`
	Width  float64 `xml:"width,attr"`
	Align  string  `xml:"align,attr"`
	Format string  `xml:"format,attr"`
}

// List represents a list element.
type List struct {
	BaseElement
	Items  string  `xml:"items,attr"`
	Style  string  `xml:"style,attr"`
	Bullet string  `xml:"bullet,attr"`
	Indent float64 `xml:"indent,attr"`
}

// GetType returns the element type.
func (l List) GetType() string { return "list" }

// GetCondition returns the condition for rendering.
func (l List) GetCondition() string { return l.Condition }

// KeyValueList represents a key-value list element.
type KeyValueList struct {
	BaseElement
	Style      string         `xml:"style,attr"`
	KeyWidth   float64        `xml:"keyWidth,attr"`
	ValueWidth float64        `xml:"valueWidth,attr"`
	ValueAlign string         `xml:"valueAlign,attr"`
	Items      []KeyValueItem `xml:"item"`
}

// GetType returns the element type.
func (k KeyValueList) GetType() string { return "keyValueList" }

// GetCondition returns the condition for rendering.
func (k KeyValueList) GetCondition() string { return k.Condition }

// KeyValueItem represents a key-value pair.
type KeyValueItem struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

// Line represents a line element.
type Line struct {
	BaseElement
	X1    float64 `xml:"x1,attr"`
	Y1    float64 `xml:"y1,attr"`
	X2    float64 `xml:"x2,attr"`
	Y2    float64 `xml:"y2,attr"`
	Color string  `xml:"color,attr"`
	Width float64 `xml:"width,attr"`
}

// GetType returns the element type.
func (l Line) GetType() string { return "line" }

// GetCondition returns the condition for rendering.
func (l Line) GetCondition() string { return l.Condition }

// Rectangle represents a rectangle element.
type Rectangle struct {
	BaseElement
	X           float64   `xml:"x,attr"`
	Y           float64   `xml:"y,attr"`
	Width       float64   `xml:"width,attr"`
	Height      float64   `xml:"height,attr"`
	Radius      float64   `xml:"radius,attr"`
	FillColor   *RGBColor `xml:"fillColor"`
	BorderColor *RGBColor `xml:"borderColor"`
	BorderWidth float64   `xml:"borderWidth"`
}

// GetType returns the element type.
func (r Rectangle) GetType() string { return "rectangle" }

// GetCondition returns the condition for rendering.
func (r Rectangle) GetCondition() string { return r.Condition }

// PageBreak represents a page break element.
type PageBreak struct {
	Condition string `xml:"condition,attr"`
}

// GetType returns the element type.
func (p PageBreak) GetType() string { return "pageBreak" }

// GetCondition returns the condition for rendering.
func (p PageBreak) GetCondition() string { return p.Condition }
