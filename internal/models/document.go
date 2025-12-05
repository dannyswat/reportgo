// Package models defines the document configuration structures.
package models

// Document represents the document configuration.
type Document struct {
	Orientation string      `xml:"orientation,attr"`
	Unit        string      `xml:"unit,attr"`
	Format      string      `xml:"format,attr"`
	Margins     *Margins    `xml:"margins"`
	CustomSize  *CustomSize `xml:"customSize"`
}

// Margins defines the page margins.
type Margins struct {
	Top    float64 `xml:"top,attr"`
	Right  float64 `xml:"right,attr"`
	Bottom float64 `xml:"bottom,attr"`
	Left   float64 `xml:"left,attr"`
}

// CustomSize defines custom page dimensions.
type CustomSize struct {
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
}

// Header represents the page header configuration.
type Header struct {
	Enabled bool    `xml:"enabled,attr"`
	Height  float64 `xml:"height,attr"`
	Texts   []Text  `xml:"text"`
	Images  []Image `xml:"image"`
	Lines   []Line  `xml:"line"`
}

// Footer represents the page footer configuration.
type Footer struct {
	Enabled bool    `xml:"enabled,attr"`
	Height  float64 `xml:"height,attr"`
	Texts   []Text  `xml:"text"`
	Images  []Image `xml:"image"`
	Lines   []Line  `xml:"line"`
}

// Sections contains all report sections.
type Sections struct {
	Sections []Section `xml:"section"`
}

// Section represents a content section in the report.
type Section struct {
	Name            string `xml:"name,attr"`
	PageBreakBefore bool   `xml:"pageBreakBefore,attr"`
	PageBreakAfter  bool   `xml:"pageBreakAfter,attr"`
	Condition       string `xml:"condition,attr"`
	Loop            string `xml:"loop,attr"`
	LoopVariable    string `xml:"loopVariable,attr"`

	// Elements
	Texts         []Text         `xml:"text"`
	Images        []Image        `xml:"image"`
	Tables        []Table        `xml:"table"`
	Lists         []List         `xml:"list"`
	KeyValueLists []KeyValueList `xml:"keyValueList"`
	Lines         []Line         `xml:"line"`
	Rectangles    []Rectangle    `xml:"rectangle"`
	PageBreaks    []PageBreak    `xml:"pageBreak"`
}
