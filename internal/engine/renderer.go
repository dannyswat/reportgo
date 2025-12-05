// Package engine provides element rendering functionality.
package engine

import (
	"github.com/dannyswat/reportgo/internal/models"
)

// renderText renders a text element.
func (e *Engine) renderText(text *models.Text) {
	if text.Style != "" {
		e.applyStyle(text.Style)
	}

	content := e.processTemplate(text.Content)

	// Handle positioning
	if text.X != 0 || text.Y != 0 {
		x := text.X
		y := text.Y
		sizeX, sizeY := e.pdf.GetPageSize()
		if x < 0 {
			x = sizeX + x
		}
		if y < 0 {
			y = sizeY + y
		}
		e.pdf.SetXY(x, y)
	}

	// Get alignment
	align := "L"
	if text.Align != "" {
		align = text.Align
	}

	// Get line height from style
	lineHeight := 6.0
	if style, ok := e.styles[text.Style]; ok && style.LineHeight > 0 {
		lineHeight = style.LineHeight
	}

	// Render text
	if text.Width > 0 {
		e.pdf.MultiCell(text.Width, lineHeight, content, "", align, false)
	} else {
		e.pdf.CellFormat(0, lineHeight, content, "", 1, align, false, 0, "")
	}

	// Add spacing after
	if text.SpacingAfter > 0 {
		e.pdf.Ln(text.SpacingAfter)
	}
}

// renderImage renders an image element.
func (e *Engine) renderImage(img *models.Image) {
	path := e.processTemplate(img.Path)

	x := img.X
	y := img.Y

	// Handle alignment
	if img.Align == "C" {
		pageWidth, _ := e.pdf.GetPageSize()
		marginLeft, _, marginRight, _ := e.pdf.GetMargins()
		contentWidth := pageWidth - marginLeft - marginRight
		x = marginLeft + (contentWidth-img.Width)/2
	} else if img.Align == "R" {
		pageWidth, _ := e.pdf.GetPageSize()
		_, _, marginRight, _ := e.pdf.GetMargins()
		x = pageWidth - marginRight - img.Width
	}

	if y == 0 {
		y = e.pdf.GetY()
	}

	e.pdf.Image(path, x, y, img.Width, img.Height, false, "", 0, "")

	// Move Y position
	if img.Height > 0 {
		e.pdf.SetY(y + img.Height)
	}

	if img.SpacingAfter > 0 {
		e.pdf.Ln(img.SpacingAfter)
	}
}

// renderTable renders a table element.
func (e *Engine) renderTable(table *models.Table) {
	// Apply header style
	if table.HeaderStyle != "" {
		e.applyStyle(table.HeaderStyle)
	}

	// Render header row
	for _, col := range table.Columns.Columns {
		fill := table.HeaderStyle != ""
		e.pdf.CellFormat(col.Width, 7, col.Header, "1", 0, "C", fill, 0, "")
	}
	e.pdf.Ln(-1)

	// Get data from template
	dataKey := extractDataKey(table.DataSource)
	data, ok := e.data[dataKey]
	if !ok {
		return
	}

	rows, ok := data.([]interface{})
	if !ok {
		return
	}

	// Render data rows
	for i, row := range rows {
		if table.CellStyle != "" {
			e.applyStyle(table.CellStyle)
		}

		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}

		// Alternate row colors
		fill := false
		if table.AlternateRowColor != nil && i%2 == 0 {
			r, g, b := table.AlternateRowColor.ToRGB()
			e.pdf.SetFillColor(r, g, b)
			fill = true
		}

		for _, col := range table.Columns.Columns {
			value := formatValue(rowMap[col.Field], col.Format)
			align := col.Align
			if align == "" {
				align = "L"
			}
			border := ""
			if table.Border {
				border = "1"
			}
			e.pdf.CellFormat(col.Width, 7, value, border, 0, align, fill, 0, "")
		}
		e.pdf.Ln(-1)
	}

	if table.SpacingAfter > 0 {
		e.pdf.Ln(table.SpacingAfter)
	}
}

// renderList renders a list element.
func (e *Engine) renderList(list *models.List) {
	if list.Style != "" {
		e.applyStyle(list.Style)
	}

	dataKey := extractDataKey(list.Items)
	data, ok := e.data[dataKey]
	if !ok {
		return
	}

	items, ok := data.([]interface{})
	if !ok {
		return
	}

	bullet := list.Bullet
	if bullet == "" {
		bullet = "•"
	}

	indent := list.Indent
	if indent == 0 {
		indent = 10
	}

	marginLeft, _, _, _ := e.pdf.GetMargins()

	for _, item := range items {
		str, ok := item.(string)
		if !ok {
			continue
		}

		e.pdf.SetX(marginLeft + indent)
		e.pdf.CellFormat(10, 6, bullet, "", 0, "L", false, 0, "")
		e.pdf.CellFormat(0, 6, str, "", 1, "L", false, 0, "")
	}

	if list.SpacingAfter > 0 {
		e.pdf.Ln(list.SpacingAfter)
	}
}

// renderKeyValueList renders a key-value list element.
func (e *Engine) renderKeyValueList(kvList *models.KeyValueList) {
	if kvList.Style != "" {
		e.applyStyle(kvList.Style)
	}

	keyWidth := kvList.KeyWidth
	if keyWidth == 0 {
		keyWidth = 50
	}

	for _, item := range kvList.Items {
		value := e.processTemplate(item.Value)
		e.pdf.CellFormat(keyWidth, 6, item.Key+":", "", 0, "L", false, 0, "")
		e.pdf.CellFormat(0, 6, value, "", 1, "L", false, 0, "")
	}

	if kvList.SpacingAfter > 0 {
		e.pdf.Ln(kvList.SpacingAfter)
	}
}

// renderLine renders a line element.
func (e *Engine) renderLine(line *models.Line) {
	pageWidth, pageHeight := e.pdf.GetPageSize()

	x1, y1 := line.X1, line.Y1
	x2, y2 := line.X2, line.Y2

	// Handle negative coordinates
	if x1 < 0 {
		x1 = pageWidth + x1
	}
	if x2 < 0 {
		x2 = pageWidth + x2
	}
	if y1 < 0 {
		y1 = pageHeight + y1
	}
	if y2 < 0 {
		y2 = pageHeight + y2
	}

	// Set line color
	if line.Color != "" {
		r, g, b := models.ParseColor(line.Color)
		e.pdf.SetDrawColor(r, g, b)
	}

	// Set line width
	if line.Width > 0 {
		e.pdf.SetLineWidth(line.Width)
	}

	e.pdf.Line(x1, y1, x2, y2)
}

// renderRectangle renders a rectangle element.
func (e *Engine) renderRectangle(rect *models.Rectangle) {
	style := ""

	if rect.FillColor != nil {
		r, g, b := rect.FillColor.ToRGB()
		e.pdf.SetFillColor(r, g, b)
		style = "F"
	}

	if rect.BorderColor != nil {
		r, g, b := rect.BorderColor.ToRGB()
		e.pdf.SetDrawColor(r, g, b)
		if style == "F" {
			style = "FD"
		} else {
			style = "D"
		}
	}

	if rect.BorderWidth > 0 {
		e.pdf.SetLineWidth(rect.BorderWidth)
	}

	if rect.Radius > 0 {
		e.pdf.RoundedRect(rect.X, rect.Y, rect.Width, rect.Height, rect.Radius, "1234", style)
	} else {
		e.pdf.Rect(rect.X, rect.Y, rect.Width, rect.Height, style)
	}

	if rect.SpacingAfter > 0 {
		e.pdf.SetY(rect.Y + rect.Height + rect.SpacingAfter)
	}
}

// extractDataKey extracts the data key from a template string like "{{.SalesData}}"
func extractDataKey(tmpl string) string {
	if len(tmpl) < 5 {
		return tmpl
	}
	// Remove {{ and }}
	key := tmpl
	if len(key) > 2 && key[:2] == "{{" {
		key = key[2:]
	}
	if len(key) > 2 && key[len(key)-2:] == "}}" {
		key = key[:len(key)-2]
	}
	// Remove leading . and spaces
	key = trimPrefix(key, ".")
	key = trimSpace(key)
	return key
}

func trimPrefix(s, prefix string) string {
	for len(s) > 0 && len(prefix) > 0 && s[0] == prefix[0] {
		s = s[1:]
		break
	}
	return s
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// formatValue formats a value based on the column format.
func formatValue(val interface{}, format string) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case float64:
		switch format {
		case "currency":
			return formatCurrency(v)
		case "percent":
			return formatPercent(v)
		default:
			return formatNumber(v)
		}
	case int:
		return formatNumber(float64(v))
	default:
		return ""
	}
}

func formatCurrency(v float64) string {
	return "$" + formatNumber(v)
}

func formatPercent(v float64) string {
	return formatNumber(v*100) + "%"
}

func formatNumber(v float64) string {
	// Simple formatting
	if v == float64(int(v)) {
		return intToString(int(v))
	}
	return floatToString(v, 2)
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

func floatToString(f float64, decimals int) string {
	// Multiply to shift decimal places
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}

	rounded := int(f*multiplier + 0.5)
	intPart := rounded / int(multiplier)
	decPart := rounded % int(multiplier)

	result := intToString(intPart) + "."

	// Pad decimal part with leading zeros if needed
	decStr := intToString(decPart)
	for len(decStr) < decimals {
		decStr = "0" + decStr
	}

	return result + decStr
}
