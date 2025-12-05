// Package models defines the document configuration structures.
package models

import "encoding/xml"

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

// SectionElement represents any element that can appear in a section.
type SectionElement struct {
	Type      string
	Text      *Text
	Image     *Image
	Table     *Table
	List      *List
	KVList    *KeyValueList
	Line      *Line
	Rectangle *Rectangle
	PageBreak *PageBreak
}

// Section represents a content section in the report.
type Section struct {
	Name            string `xml:"name,attr"`
	PageBreakBefore bool   `xml:"pageBreakBefore,attr"`
	PageBreakAfter  bool   `xml:"pageBreakAfter,attr"`
	Condition       string `xml:"condition,attr"`
	Loop            string `xml:"loop,attr"`
	LoopVariable    string `xml:"loopVariable,attr"`

	// Elements in document order
	Elements []SectionElement

	// Legacy fields for backward compatibility (populated during UnmarshalXML)
	Texts         []Text         `xml:"-"`
	Images        []Image        `xml:"-"`
	Tables        []Table        `xml:"-"`
	Lists         []List         `xml:"-"`
	KeyValueLists []KeyValueList `xml:"-"`
	Lines         []Line         `xml:"-"`
	Rectangles    []Rectangle    `xml:"-"`
	PageBreaks    []PageBreak    `xml:"-"`
}

// UnmarshalXML implements custom XML unmarshaling to preserve element order.
func (s *Section) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Parse attributes
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "name":
			s.Name = attr.Value
		case "pageBreakBefore":
			s.PageBreakBefore = attr.Value == "true"
		case "pageBreakAfter":
			s.PageBreakAfter = attr.Value == "true"
		case "condition":
			s.Condition = attr.Value
		case "loop":
			s.Loop = attr.Value
		case "loopVariable":
			s.LoopVariable = attr.Value
		}
	}

	// Parse child elements in order
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			var elem SectionElement
			switch t.Name.Local {
			case "text":
				var text Text
				if err := d.DecodeElement(&text, &t); err != nil {
					return err
				}
				elem.Type = "text"
				elem.Text = &text
				s.Texts = append(s.Texts, text)
			case "image":
				var img Image
				if err := d.DecodeElement(&img, &t); err != nil {
					return err
				}
				elem.Type = "image"
				elem.Image = &img
				s.Images = append(s.Images, img)
			case "table":
				var table Table
				if err := d.DecodeElement(&table, &t); err != nil {
					return err
				}
				elem.Type = "table"
				elem.Table = &table
				s.Tables = append(s.Tables, table)
			case "list":
				var list List
				if err := d.DecodeElement(&list, &t); err != nil {
					return err
				}
				elem.Type = "list"
				elem.List = &list
				s.Lists = append(s.Lists, list)
			case "keyValueList":
				var kvList KeyValueList
				if err := d.DecodeElement(&kvList, &t); err != nil {
					return err
				}
				elem.Type = "keyValueList"
				elem.KVList = &kvList
				s.KeyValueLists = append(s.KeyValueLists, kvList)
			case "line":
				var line Line
				if err := d.DecodeElement(&line, &t); err != nil {
					return err
				}
				elem.Type = "line"
				elem.Line = &line
				s.Lines = append(s.Lines, line)
			case "rectangle":
				var rect Rectangle
				if err := d.DecodeElement(&rect, &t); err != nil {
					return err
				}
				elem.Type = "rectangle"
				elem.Rectangle = &rect
				s.Rectangles = append(s.Rectangles, rect)
			case "pageBreak":
				var pb PageBreak
				if err := d.DecodeElement(&pb, &t); err != nil {
					return err
				}
				elem.Type = "pageBreak"
				elem.PageBreak = &pb
				s.PageBreaks = append(s.PageBreaks, pb)
			default:
				// Skip unknown elements
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}
			s.Elements = append(s.Elements, elem)
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}
