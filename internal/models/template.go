// Package models defines the data structures for ReportGo templates.
package models

// Report represents the root element of a report template.
type Report struct {
	Version  string    `xml:"version,attr"`
	Metadata *Metadata `xml:"metadata"`
	Document Document  `xml:"document"`
	Fonts    *Fonts    `xml:"fonts"`
	Styles   *Styles   `xml:"styles"`
	Header   *Header   `xml:"header"`
	Footer   *Footer   `xml:"footer"`
	Sections Sections  `xml:"sections"`
}

// Metadata contains template metadata.
type Metadata struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	Created     string `xml:"created"`
	Modified    string `xml:"modified"`
}
