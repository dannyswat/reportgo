// Package models defines the document configuration structures.
package models

import (
	"encoding/xml"
	"fmt"
)

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
	Row       *Row
	RowGrid   *RowGrid
	Spacer    *Spacer
	PageBreak *PageBreak
}

// Section represents a content section in the report.
type Section struct {
	Name            string  `xml:"name,attr"`
	PageBreakBefore bool    `xml:"pageBreakBefore,attr"`
	PageBreakAfter  bool    `xml:"pageBreakAfter,attr"`
	Condition       string  `xml:"condition,attr"`
	Loop            string  `xml:"loop,attr"`
	LoopVariable    string  `xml:"loopVariable,attr"`
	PaddingLeft     float64 `xml:"paddingLeft,attr"`

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
	Rows          []Row          `xml:"-"`
	RowGrids      []RowGrid      `xml:"-"`
	Spacers       []Spacer       `xml:"-"`
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
		case "paddingLeft":
			if _, err := fmt.Sscanf(attr.Value, "%f", &s.PaddingLeft); err != nil {
				s.PaddingLeft = 0
			}
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
			elem, ok, err := decodeSectionElement(d, &t)
			if err != nil {
				return err
			}
			if !ok {
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}
			s.appendLegacyElement(elem)
			s.Elements = append(s.Elements, elem)
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

func (s *Section) appendLegacyElement(elem SectionElement) {
	switch elem.Type {
	case "text":
		if elem.Text != nil {
			s.Texts = append(s.Texts, *elem.Text)
		}
	case "image":
		if elem.Image != nil {
			s.Images = append(s.Images, *elem.Image)
		}
	case "table":
		if elem.Table != nil {
			s.Tables = append(s.Tables, *elem.Table)
		}
	case "list":
		if elem.List != nil {
			s.Lists = append(s.Lists, *elem.List)
		}
	case "keyValueList":
		if elem.KVList != nil {
			s.KeyValueLists = append(s.KeyValueLists, *elem.KVList)
		}
	case "line":
		if elem.Line != nil {
			s.Lines = append(s.Lines, *elem.Line)
		}
	case "rectangle":
		if elem.Rectangle != nil {
			s.Rectangles = append(s.Rectangles, *elem.Rectangle)
		}
	case "row":
		if elem.Row != nil {
			s.Rows = append(s.Rows, *elem.Row)
		}
	case "rowgrid":
		if elem.RowGrid != nil {
			s.RowGrids = append(s.RowGrids, *elem.RowGrid)
		}
	case "spacer":
		if elem.Spacer != nil {
			s.Spacers = append(s.Spacers, *elem.Spacer)
		}
	case "pageBreak":
		if elem.PageBreak != nil {
			s.PageBreaks = append(s.PageBreaks, *elem.PageBreak)
		}
	}
}

// UnmarshalXML implements custom XML unmarshaling for rows.
func (r *Row) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "condition":
			r.Condition = attr.Value
		case "spacingAfter":
			if _, err := fmt.Sscanf(attr.Value, "%f", &r.SpacingAfter); err != nil {
				r.SpacingAfter = 0
			}
		}
	}

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
			case "image":
				var img Image
				if err := d.DecodeElement(&img, &t); err != nil {
					return err
				}
				elem.Type = "image"
				elem.Image = &img
			default:
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}
			r.Elements = append(r.Elements, elem)
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

// UnmarshalXML implements custom XML unmarshaling for rowgrids.
func (r *RowGrid) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "condition":
			r.Condition = attr.Value
		case "spacingAfter":
			if _, err := fmt.Sscanf(attr.Value, "%f", &r.SpacingAfter); err != nil {
				r.SpacingAfter = 0
			}
		case "columns":
			if _, err := fmt.Sscanf(attr.Value, "%d", &r.Columns); err != nil {
				r.Columns = 0
			}
		}
	}

	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local != "col" {
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}

			var col RowGridColumn
			if err := d.DecodeElement(&col, &t); err != nil {
				return err
			}
			r.Cols = append(r.Cols, col)
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

// UnmarshalXML implements custom XML unmarshaling for rowgrid columns.
func (c *RowGridColumn) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			elem, ok, err := decodeSectionElement(d, &t)
			if err != nil {
				return err
			}
			if !ok {
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}
			c.Elements = append(c.Elements, elem)
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

func decodeSectionElement(d *xml.Decoder, start *xml.StartElement) (SectionElement, bool, error) {
	var elem SectionElement

	switch start.Name.Local {
	case "text":
		var text Text
		if err := d.DecodeElement(&text, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "text"
		elem.Text = &text
	case "image":
		var img Image
		if err := d.DecodeElement(&img, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "image"
		elem.Image = &img
	case "table":
		var table Table
		if err := d.DecodeElement(&table, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "table"
		elem.Table = &table
	case "list":
		var list List
		if err := d.DecodeElement(&list, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "list"
		elem.List = &list
	case "keyValueList":
		var kvList KeyValueList
		if err := d.DecodeElement(&kvList, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "keyValueList"
		elem.KVList = &kvList
	case "line":
		var line Line
		if err := d.DecodeElement(&line, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "line"
		elem.Line = &line
	case "rectangle":
		var rect Rectangle
		if err := d.DecodeElement(&rect, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "rectangle"
		elem.Rectangle = &rect
	case "row":
		var row Row
		if err := d.DecodeElement(&row, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "row"
		elem.Row = &row
	case "rowgrid":
		var rowGrid RowGrid
		if err := d.DecodeElement(&rowGrid, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "rowgrid"
		elem.RowGrid = &rowGrid
	case "spacer":
		var spacer Spacer
		if err := d.DecodeElement(&spacer, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "spacer"
		elem.Spacer = &spacer
	case "pageBreak":
		var pb PageBreak
		if err := d.DecodeElement(&pb, start); err != nil {
			return SectionElement{}, false, err
		}
		elem.Type = "pageBreak"
		elem.PageBreak = &pb
	default:
		return SectionElement{}, false, nil
	}

	return elem, true, nil
}
