// Package engine provides the core report generation engine.
package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/dannyswat/reportgo/internal/models"
	"github.com/phpdave11/gofpdf"
)

// Engine is the core report generation engine.
type Engine struct {
	report  *models.Report
	data    map[string]interface{}
	pdf     *gofpdf.Fpdf
	styles  map[string]*models.Style
	funcMap template.FuncMap
}

// New creates a new Engine instance.
func New() *Engine {
	return &Engine{
		styles:  make(map[string]*models.Style),
		funcMap: defaultFuncMap(),
	}
}

// SetReport sets the parsed report template.
func (e *Engine) SetReport(report *models.Report) {
	e.report = report
	e.buildStyleMap()
}

// SetData sets the data for template rendering.
func (e *Engine) SetData(data map[string]interface{}) {
	e.data = data
}

// Generate generates the PDF and writes it to the given writer.
func (e *Engine) Generate(w io.Writer) error {
	if e.report == nil {
		return fmt.Errorf("no report template loaded")
	}

	// Initialize PDF
	e.initPDF()

	// Set up header/footer
	e.setupHeaderFooter()

	// Add first page
	e.pdf.AddPage()

	// Render sections
	for _, section := range e.report.Sections.Sections {
		if err := e.renderSection(&section); err != nil {
			return fmt.Errorf("failed to render section %s: %w", section.Name, err)
		}
	}

	// Check for errors
	if e.pdf.Err() {
		return fmt.Errorf("PDF generation error: %w", e.pdf.Error())
	}

	// Write output
	return e.pdf.Output(w)
}

// GenerateToFile generates the PDF and writes it to a file.
func (e *Engine) GenerateToFile(filepath string) error {
	if e.report == nil {
		return fmt.Errorf("no report template loaded")
	}

	e.initPDF()
	e.setupHeaderFooter()
	e.pdf.AddPage()

	for _, section := range e.report.Sections.Sections {
		if err := e.renderSection(&section); err != nil {
			return fmt.Errorf("failed to render section %s: %w", section.Name, err)
		}
	}

	if e.pdf.Err() {
		return fmt.Errorf("PDF generation error: %w", e.pdf.Error())
	}

	return e.pdf.OutputFileAndClose(filepath)
}

// initPDF initializes the PDF document.
func (e *Engine) initPDF() {
	doc := e.report.Document
	e.pdf = gofpdf.New(
		string(doc.Orientation[0:1]),
		doc.Unit,
		doc.Format,
		"",
	)

	if doc.Margins != nil {
		e.pdf.SetMargins(doc.Margins.Left, doc.Margins.Top, doc.Margins.Right)
		e.pdf.SetAutoPageBreak(true, doc.Margins.Bottom)
	}

	// Load custom fonts
	if e.report.Fonts != nil {
		for _, font := range e.report.Fonts.Fonts {
			// Read font file ourselves to avoid gofpdf path.Join bug with absolute paths
			fontBytes, err := os.ReadFile(font.File)
			if err != nil {
				// Log warning but continue - font may not be critical
				fmt.Fprintf(os.Stderr, "Warning: could not load font %s: %v\n", font.File, err)
				continue
			}
			e.pdf.AddUTF8FontFromBytes(font.Family, font.Style, fontBytes)
		}
	}
}

// buildStyleMap builds a map of style names to styles for quick lookup.
func (e *Engine) buildStyleMap() {
	if e.report.Styles == nil {
		return
	}
	for i := range e.report.Styles.Styles {
		style := &e.report.Styles.Styles[i]
		e.styles[style.Name] = style
	}
}

// setupHeaderFooter configures page headers and footers.
func (e *Engine) setupHeaderFooter() {
	e.pdf.SetHeaderFuncMode(func() {
		if e.report.Header != nil && e.report.Header.Enabled {
			e.renderHeaderFooterElements(e.report.Header.Texts, e.report.Header.Images, e.report.Header.Lines)
		}
	}, true)

	e.pdf.SetFooterFunc(func() {
		if e.report.Footer != nil && e.report.Footer.Enabled {
			e.pdf.SetY(-e.report.Footer.Height)
			e.renderHeaderFooterElements(e.report.Footer.Texts, e.report.Footer.Images, e.report.Footer.Lines)
		}
	})
}

// renderHeaderFooterElements renders header/footer elements.
func (e *Engine) renderHeaderFooterElements(texts []models.Text, images []models.Image, lines []models.Line) {
	for _, text := range texts {
		e.renderText(&text)
	}
	for _, img := range images {
		e.renderImage(&img)
	}
	for _, line := range lines {
		e.renderLine(&line)
	}
}

// renderSection renders a section and its elements.
func (e *Engine) renderSection(section *models.Section) error {
	// Handle page break before
	if section.PageBreakBefore {
		e.pdf.AddPage()
	}

	// Render elements
	for _, text := range section.Texts {
		e.renderText(&text)
	}
	for _, img := range section.Images {
		e.renderImage(&img)
	}
	for _, table := range section.Tables {
		e.renderTable(&table)
	}
	for _, list := range section.Lists {
		e.renderList(&list)
	}
	for _, kvList := range section.KeyValueLists {
		e.renderKeyValueList(&kvList)
	}
	for _, line := range section.Lines {
		e.renderLine(&line)
	}
	for _, rect := range section.Rectangles {
		e.renderRectangle(&rect)
	}
	for range section.PageBreaks {
		e.pdf.AddPage()
	}

	// Handle page break after
	if section.PageBreakAfter {
		e.pdf.AddPage()
	}

	return nil
}

// processTemplate processes a template string with data.
func (e *Engine) processTemplate(tmplStr string) string {
	if !strings.Contains(tmplStr, "{{") {
		return tmplStr
	}

	tmpl, err := template.New("").Funcs(e.funcMap).Parse(tmplStr)
	if err != nil {
		return tmplStr
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, e.data); err != nil {
		return tmplStr
	}

	return buf.String()
}

// applyStyle applies a named style to the PDF.
func (e *Engine) applyStyle(styleName string) {
	style, ok := e.styles[styleName]
	if !ok {
		return
	}

	fontStyle := style.FontStyle
	if style.FontFamily != "" {
		e.pdf.SetFont(style.FontFamily, fontStyle, style.FontSize)
	}

	if style.TextColor != nil {
		r, g, b := style.TextColor.ToRGB()
		e.pdf.SetTextColor(r, g, b)
	}

	if style.FillColor != nil {
		r, g, b := style.FillColor.ToRGB()
		e.pdf.SetFillColor(r, g, b)
	}
}

// defaultFuncMap returns the default template function map.
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"title":   strings.Title,
		"trim":    strings.TrimSpace,
		"default": defaultValue,
		"add":     add,
		"sub":     sub,
		"mul":     mul,
		"div":     div,
	}
}

func defaultValue(def, val interface{}) interface{} {
	if val == nil || val == "" {
		return def
	}
	return val
}

func add(a, b float64) float64 { return a + b }
func sub(a, b float64) float64 { return a - b }
func mul(a, b float64) float64 { return a * b }
func div(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
