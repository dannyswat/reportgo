// Package parser provides XML template and data parsing functionality.
package parser

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/dannyswat/reportgo/internal/models"
)

// ParseTemplate parses an XML template file.
func ParseTemplate(filepath string) (*models.Report, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	return ParseTemplateFromBytes(data)
}

// ParseTemplateFromBytes parses an XML template from bytes.
func ParseTemplateFromBytes(data []byte) (*models.Report, error) {
	var report models.Report
	if err := xml.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse template XML: %w", err)
	}

	applyDefaults(&report)

	return &report, nil
}

// ParseTemplateFromString parses an XML template from a string.
func ParseTemplateFromString(xmlStr string) (*models.Report, error) {
	return ParseTemplateFromBytes([]byte(xmlStr))
}

// applyDefaults applies default values to the report configuration.
func applyDefaults(report *models.Report) {
	if report.Document.Orientation == "" {
		report.Document.Orientation = "portrait"
	}
	if report.Document.Unit == "" {
		report.Document.Unit = "mm"
	}
	if report.Document.Format == "" {
		report.Document.Format = "A4"
	}

	if report.Document.Margins == nil {
		report.Document.Margins = &models.Margins{
			Top:    15,
			Right:  15,
			Bottom: 15,
			Left:   15,
		}
	}

	if report.Header != nil && report.Header.Height == 0 {
		report.Header.Height = 15
	}

	if report.Footer != nil && report.Footer.Height == 0 {
		report.Footer.Height = 15
	}
}
