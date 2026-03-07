# ReportGo

A Go-based PDF report generator that creates professional PDF reports from XML templates and JSON data.

## Features

- **XML Template Format** - Define report structure using XML with XSD schema validation
- **Data Binding** - Use Go template syntax for dynamic content
- **Rich Elements** - Text, images, tables, lists, lines, rectangles
- **Styling** - Reusable styles with fonts, colors, and alignment
- **Style Inheritance** - Derive styles with `extends` to avoid duplication
- **Headers/Footers** - Automatic page headers and footers
- **Page Management** - Automatic page breaks and custom margins

## Installation

```bash
go get github.com/dannyswat/reportgo
```

## Quick Start

```go
package main

import (
    "bytes"
    "log"
    "strings"
    "text/template"

    "github.com/dannyswat/reportgo/pkg/reportgo"
)

func main() {
    engine := reportgo.New(
        reportgo.WithFuncMap(template.FuncMap{
            "shout": func(input string) string { return strings.ToUpper(input) + "!" },
        }),
    )
    
    // Load template
    if err := engine.LoadTemplate("template.xml"); err != nil {
        log.Fatal(err)
    }
    
    // Set data
    data := map[string]interface{}{
        "Title": "My Report",
        "Items": []map[string]interface{}{
            {"name": "Item 1", "value": 100},
            {"name": "Item 2", "value": 200},
        },
    }
    
    // Generate PDF
    if err := engine.Generate(data, "output.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

You can also render directly to an `io.Writer` and pass inline data in the same call:

```go
var output bytes.Buffer
if err := engine.GenerateToWriter(&output, map[string]interface{}{"Title": "Writer Output"}); err != nil {
    log.Fatal(err)
}
```

## CLI Usage

```bash
# Generate PDF from template and data
reportgo -template report.xml -data data.json -output report.pdf

# Validate template only
reportgo -template report.xml -validate
```

## Project Structure

```
reportgo/
├── cmd/reportgo/          # CLI application
├── internal/
│   ├── engine/            # PDF generation engine
│   ├── parser/            # XML/JSON parsing
│   └── models/            # Data structures
├── pkg/reportgo/          # Public API
├── schemas/               # XSD schema
├── templates/examples/    # Example templates
└── examples/              # Usage examples
```

## Template Format

Templates use XML format with XSD schema validation. See [DESIGN.md](DESIGN.md) for full specification.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4">
        <margins top="15" right="15" bottom="15" left="15"/>
    </document>
    
    <styles>
        <style name="title">
            <fontFamily>Arial</fontFamily>
            <fontSize>24</fontSize>
        </style>

        <style name="title_emphasis" extends="title">
            <fontStyle>B</fontStyle>
        </style>
    </styles>
    
    <sections>
        <section name="main">
            <text style="title">{{.Title}}</text>
        </section>
    </sections>
</report>
```

## Template Functions

Built-in helpers include `upper`, `lower`, `title`, `trim`, `default`, `add`, `sub`, `mul`, `div`, `join`, `replace`, `ifelse`, `truncate`, `formatDate`, `dateFormat`, `formatNumber`, `formatCurrency`, and `formatPercent`.

Applications can register additional helpers with `reportgo.WithFuncMap(...)` or `engine.AddFuncMap(...)`.

## Embedded Fonts

Applications can register in-memory fonts with `reportgo.WithEmbeddedFont(...)` or `engine.AddEmbeddedFont(...)`. Embedded fonts are loaded before file-based fonts from the template, so they can be used in environments where direct font file access is inconvenient.

## Dependencies

- [gofpdf](https://github.com/phpdave11/gofpdf) - PDF generation

## License

MIT License
