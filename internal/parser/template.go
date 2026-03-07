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
	if err := resolveStyleInheritance(&report); err != nil {
		return nil, err
	}

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

func resolveStyleInheritance(report *models.Report) error {
	if report.Styles == nil {
		return nil
	}

	styleIndex := make(map[string]models.Style, len(report.Styles.Styles))
	for _, style := range report.Styles.Styles {
		styleIndex[style.Name] = style
	}

	resolved := make(map[string]models.Style, len(styleIndex))
	visiting := make(map[string]bool, len(styleIndex))

	var resolve func(string) (models.Style, error)
	resolve = func(name string) (models.Style, error) {
		if style, ok := resolved[name]; ok {
			return style, nil
		}

		style, ok := styleIndex[name]
		if !ok {
			return models.Style{}, fmt.Errorf("style %q extends unknown style", name)
		}
		if visiting[name] {
			return models.Style{}, fmt.Errorf("style inheritance cycle detected at %q", name)
		}

		visiting[name] = true
		defer delete(visiting, name)

		if style.Extends == "" {
			resolved[name] = style
			return style, nil
		}

		parent, err := resolve(style.Extends)
		if err != nil {
			return models.Style{}, err
		}

		merged := parent.Merge(style)
		resolved[name] = merged
		return merged, nil
	}

	for idx := range report.Styles.Styles {
		resolvedStyle, err := resolve(report.Styles.Styles[idx].Name)
		if err != nil {
			return fmt.Errorf("failed to resolve style inheritance: %w", err)
		}
		report.Styles.Styles[idx] = resolvedStyle
	}

	return nil
}
