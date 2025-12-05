// Package reportgo provides the public API for PDF report generation.
package reportgo

import (
	"io"

	"github.com/dannyswat/reportgo/internal/engine"
	"github.com/dannyswat/reportgo/internal/parser"
)

// Engine is the main report generation engine.
type Engine struct {
	engine *engine.Engine
	data   map[string]interface{}
}

// Option is a function that configures the engine.
type Option func(*Engine)

// New creates a new report engine with optional configuration.
func New(opts ...Option) *Engine {
	e := &Engine{
		engine: engine.New(),
		data:   make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithFontPath sets the base path for font files.
func WithFontPath(path string) Option {
	return func(e *Engine) {
		// Store font path for later use
		e.data["_fontPath"] = path
	}
}

// WithImagePath sets the base path for image files.
func WithImagePath(path string) Option {
	return func(e *Engine) {
		// Store image path for later use
		e.data["_imagePath"] = path
	}
}

// WithCompression enables PDF compression.
func WithCompression(enabled bool) Option {
	return func(e *Engine) {
		e.data["_compression"] = enabled
	}
}

// WithSchemaValidation enables XSD schema validation.
func WithSchemaValidation(enabled bool) Option {
	return func(e *Engine) {
		e.data["_schemaValidation"] = enabled
	}
}

// LoadTemplate loads an XML template from a file.
func (e *Engine) LoadTemplate(filepath string) error {
	report, err := parser.ParseTemplate(filepath)
	if err != nil {
		return err
	}

	e.engine.SetReport(report)
	return nil
}

// LoadTemplateFromString loads an XML template from a string.
func (e *Engine) LoadTemplateFromString(xmlStr string) error {
	report, err := parser.ParseTemplateFromString(xmlStr)
	if err != nil {
		return err
	}

	e.engine.SetReport(report)
	return nil
}

// LoadDataFromFile loads JSON data from a file.
func (e *Engine) LoadDataFromFile(filepath string) error {
	data, err := parser.ParseDataFromFile(filepath)
	if err != nil {
		return err
	}

	// Merge with existing data
	for k, v := range data {
		e.data[k] = v
	}
	e.engine.SetData(e.data)
	return nil
}

// SetData sets the data for template rendering.
func (e *Engine) SetData(data map[string]interface{}) {
	for k, v := range data {
		e.data[k] = v
	}
	e.engine.SetData(e.data)
}

// Generate generates a PDF report and writes it to a file.
// If data is provided, it will be merged with any previously set data.
func (e *Engine) Generate(data map[string]interface{}, filepath string) error {
	if data != nil {
		for k, v := range data {
			e.data[k] = v
		}
		e.engine.SetData(e.data)
	}

	return e.engine.GenerateToFile(filepath)
}

// GenerateToWriter generates a PDF report and writes it to an io.Writer.
func (e *Engine) GenerateToWriter(w io.Writer) error {
	e.engine.SetData(e.data)
	return e.engine.Generate(w)
}
