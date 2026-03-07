// Package engine provides the core report generation engine.
package engine

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dannyswat/reportgo/internal/models"
	"github.com/phpdave11/gofpdf"
)

// Engine is the core report generation engine.
type Engine struct {
	report         *models.Report
	data           map[string]interface{}
	pdf            *gofpdf.Fpdf
	styles         map[string]*models.Style
	funcMap        template.FuncMap
	flowOffsetLeft float64
	embeddedFonts  []models.EmbeddedFont
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

// AddFuncMap merges additional template functions into the engine.
func (e *Engine) AddFuncMap(funcs template.FuncMap) {
	if len(funcs) == 0 {
		return
	}
	for name, fn := range funcs {
		e.funcMap[name] = fn
	}
}

// AddEmbeddedFont registers a font from in-memory bytes.
func (e *Engine) AddEmbeddedFont(font models.EmbeddedFont) {
	e.embeddedFonts = append(e.embeddedFonts, font)
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

	if err := e.renderSections(); err != nil {
		return err
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

	if err := e.renderSections(); err != nil {
		return err
	}

	if e.pdf.Err() {
		return fmt.Errorf("PDF generation error: %w", e.pdf.Error())
	}

	return e.pdf.OutputFileAndClose(filepath)
}

func (e *Engine) renderSections() error {
	for _, section := range e.report.Sections.Sections {
		if err := e.renderSection(&section); err != nil {
			return fmt.Errorf("failed to render section %s: %w", section.Name, err)
		}
	}

	return nil
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

	// Ensure unstyled text elements can render even when no explicit style has
	// selected a font yet. Styled content will override this as needed.
	e.pdf.SetFont("Arial", "", 12)

	loadedFonts := make(map[string]bool)
	for _, font := range e.embeddedFonts {
		if len(font.Data) == 0 {
			continue
		}
		e.pdf.AddUTF8FontFromBytes(font.Family, font.Style, font.Data)
		loadedFonts[fontKey(font.Family, font.Style)] = true
	}

	// Load custom fonts
	if e.report.Fonts != nil {
		for _, font := range e.report.Fonts.Fonts {
			if loadedFonts[fontKey(font.Family, font.Style)] {
				continue
			}
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

func fontKey(family, style string) string {
	return family + "|" + style
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
	rendered := false
	renderCurrentContext := func() error {
		if !e.shouldRenderCondition(section.Condition) {
			return nil
		}

		if !rendered && section.PageBreakBefore {
			e.pdf.AddPage()
		}

		rendered = true
		return e.withFlowOffset(section.PaddingLeft, func() error {
			return e.renderSectionElements(section)
		})
	}

	if section.Loop != "" {
		loopVariable := section.LoopVariable
		if loopVariable == "" {
			loopVariable = "item"
		}

		items, ok := e.resolveLoopItems(section.Loop)
		if !ok {
			return nil
		}

		for _, item := range items {
			if err := e.withScopedData(loopVariable, item, renderCurrentContext); err != nil {
				return err
			}
		}
	} else if err := renderCurrentContext(); err != nil {
		return err
	}

	if rendered && section.PageBreakAfter {
		e.pdf.AddPage()
	}

	return nil
}

func (e *Engine) renderSectionElements(section *models.Section) error {
	for _, elem := range section.Elements {
		if !e.shouldRenderCondition(e.getElementCondition(elem)) {
			continue
		}

		switch elem.Type {
		case "text":
			e.renderText(elem.Text)
		case "image":
			e.renderImage(elem.Image)
		case "table":
			e.renderTable(elem.Table)
		case "list":
			e.renderList(elem.List)
		case "keyValueList":
			e.renderKeyValueList(elem.KVList)
		case "line":
			e.renderLine(elem.Line)
		case "rectangle":
			e.renderRectangle(elem.Rectangle)
		case "row":
			if err := e.renderRow(elem.Row); err != nil {
				return err
			}
		case "spacer":
			e.renderSpacer(elem.Spacer)
		case "pageBreak":
			e.pdf.AddPage()
		}
	}

	return nil
}

func (e *Engine) getElementCondition(elem models.SectionElement) string {
	switch elem.Type {
	case "text":
		if elem.Text != nil {
			return elem.Text.Condition
		}
	case "image":
		if elem.Image != nil {
			return elem.Image.Condition
		}
	case "table":
		if elem.Table != nil {
			return elem.Table.Condition
		}
	case "list":
		if elem.List != nil {
			return elem.List.Condition
		}
	case "keyValueList":
		if elem.KVList != nil {
			return elem.KVList.Condition
		}
	case "line":
		if elem.Line != nil {
			return elem.Line.Condition
		}
	case "rectangle":
		if elem.Rectangle != nil {
			return elem.Rectangle.Condition
		}
	case "row":
		if elem.Row != nil {
			return elem.Row.Condition
		}
	case "spacer":
		if elem.Spacer != nil {
			return elem.Spacer.Condition
		}
	case "pageBreak":
		if elem.PageBreak != nil {
			return elem.PageBreak.Condition
		}
	}

	return ""
}

func (e *Engine) shouldRenderCondition(condition string) bool {
	if strings.TrimSpace(condition) == "" {
		return true
	}

	result := strings.ToLower(strings.TrimSpace(e.processTemplate(condition)))
	switch result {
	case "", "0", "false", "nil", "null", "<no value>":
		return false
	default:
		return true
	}
}

func (e *Engine) withScopedData(key string, value interface{}, fn func() error) error {
	original := e.data
	scoped := cloneDataMap(original)
	scoped[key] = value
	e.data = scoped
	defer func() {
		e.data = original
	}()

	return fn()
}

func (e *Engine) withFlowOffset(offset float64, fn func() error) error {
	original := e.flowOffsetLeft
	e.flowOffsetLeft += offset
	defer func() {
		e.flowOffsetLeft = original
	}()

	return fn()
}

func (e *Engine) flowLeftMargin() float64 {
	marginLeft, _, _, _ := e.pdf.GetMargins()
	return marginLeft + e.flowOffsetLeft
}

func cloneDataMap(source map[string]interface{}) map[string]interface{} {
	if source == nil {
		return map[string]interface{}{}
	}

	cloned := make(map[string]interface{}, len(source)+1)
	for key, value := range source {
		cloned[key] = value
	}

	return cloned
}

func (e *Engine) resolveLoopItems(loop string) ([]interface{}, bool) {
	value, ok := e.resolveTemplateValue(loop)
	if !ok {
		return nil, false
	}

	items := toInterfaceSlice(value)
	if len(items) == 0 {
		return nil, false
	}

	return items, true
}

func (e *Engine) resolveTemplateValue(tmpl string) (interface{}, bool) {
	path := extractDataPath(tmpl)
	if len(path) == 0 {
		return nil, false
	}

	var current interface{} = e.data
	for _, part := range path {
		var ok bool
		current, ok = resolveFieldValue(current, part)
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func extractDataPath(tmpl string) []string {
	key := trimSpace(tmpl)
	if strings.HasPrefix(key, "{{") && strings.HasSuffix(key, "}}") {
		key = trimSpace(key[2 : len(key)-2])
	}

	key = strings.TrimPrefix(key, ".")
	if key == "" {
		return nil
	}

	parts := strings.Split(key, ".")
	path := make([]string, 0, len(parts))
	for _, part := range parts {
		part = trimSpace(part)
		if part != "" {
			path = append(path, part)
		}
	}

	return path
}

func resolveFieldValue(current interface{}, field string) (interface{}, bool) {
	if current == nil {
		return nil, false
	}

	if data, ok := current.(map[string]interface{}); ok {
		value, found := data[field]
		return value, found
	}

	value := reflect.ValueOf(current)
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil, false
		}
		value = value.Elem()
	}

	if !value.IsValid() {
		return nil, false
	}

	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return nil, false
		}
		result := value.MapIndex(reflect.ValueOf(field))
		if !result.IsValid() {
			return nil, false
		}
		return result.Interface(), true
	case reflect.Struct:
		result := value.FieldByName(field)
		if !result.IsValid() || !result.CanInterface() {
			return nil, false
		}
		return result.Interface(), true
	default:
		return nil, false
	}
}

func toInterfaceSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}

	if items, ok := value.([]interface{}); ok {
		return items
	}

	rv := reflect.ValueOf(value)
	for rv.IsValid() && (rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer) {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil
	}

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil
	}

	items := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		items[i] = rv.Index(i).Interface()
	}

	return items
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
		"upper":          strings.ToUpper,
		"lower":          strings.ToLower,
		"title":          strings.Title,
		"trim":           strings.TrimSpace,
		"default":        defaultValue,
		"add":            add,
		"sub":            sub,
		"mul":            mul,
		"div":            div,
		"join":           strings.Join,
		"replace":        strings.ReplaceAll,
		"ifelse":         ifElse,
		"truncate":       truncate,
		"formatDate":     formatDate,
		"dateFormat":     formatDate,
		"formatNumber":   formatNumberValue,
		"formatCurrency": formatCurrencyValue,
		"formatPercent":  formatPercentValue,
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

func ifElse(condition interface{}, trueValue, falseValue interface{}) interface{} {
	if isTruthyValue(condition) {
		return trueValue
	}
	return falseValue
}

func truncate(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func formatDate(value interface{}, layout string) string {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return ""
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02 Jan 2006",
	}

	for _, candidate := range layouts {
		if parsed, err := time.Parse(candidate, text); err == nil {
			return parsed.Format(layout)
		}
	}

	return text
}

func formatNumberValue(value interface{}, decimals int) string {
	floatValue, ok := toFloat64(value)
	if !ok {
		return fmt.Sprint(value)
	}
	if decimals < 0 {
		decimals = 0
	}
	return strconv.FormatFloat(floatValue, 'f', decimals, 64)
}

func formatCurrencyValue(value interface{}, symbol ...string) string {
	prefix := "$"
	if len(symbol) > 0 && symbol[0] != "" {
		prefix = symbol[0]
	}
	return prefix + formatNumberValue(value, 2)
}

func formatPercentValue(value interface{}, decimals ...int) string {
	floatValue, ok := toFloat64(value)
	if !ok {
		return fmt.Sprint(value)
	}
	precision := 2
	if len(decimals) > 0 {
		precision = decimals[0]
	}
	return formatNumberValue(floatValue*100, precision) + "%"
}

func isTruthyValue(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		trimmed := strings.ToLower(strings.TrimSpace(v))
		return trimmed != "" && trimmed != "0" && trimmed != "false" && trimmed != "nil" && trimmed != "null" && trimmed != "<no value>"
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	default:
		return true
	}
}

func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return float64(v), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
