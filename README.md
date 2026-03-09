# ReportGo

ReportGo is a Go PDF report generator that renders XML templates with JSON or in-memory data.

## Features

- XML templates parsed into a report model with preserved section element order.
- JSON data loading plus direct `map[string]interface{}` data injection.
- Flow-oriented elements: text, image, table, list, key-value list, line, rectangle, row, rowgrid, spacer, and page break.
- Reusable styles with inheritance through the `extends` attribute.
- Conditional sections and conditional elements through the `condition` attribute.
- Section loops through `loop` and `loopVariable`.
- Built-in template helpers plus application-defined helpers via `template.FuncMap`.
- File-based fonts from the template and in-memory embedded fonts through the public API.
- Optional headers and footers rendered from `text`, `image`, and `line` elements.

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

	if err := engine.LoadTemplate("template.xml"); err != nil {
		log.Fatal(err)
	}

	data := map[string]interface{}{
		"Title": "My Report",
		"Items": []map[string]interface{}{
			{"name": "Item 1", "value": 100},
			{"name": "Item 2", "value": 200},
		},
	}

	if err := engine.Generate(data, "output.pdf"); err != nil {
		log.Fatal(err)
	}

	var output bytes.Buffer
	if err := engine.GenerateToWriter(&output, map[string]interface{}{"Title": "Writer Output"}); err != nil {
		log.Fatal(err)
	}

	_ = output
}
```

## CLI Usage

```bash
# Generate PDF from template and data
reportgo -template report.xml -data data.json -output report.pdf

# Parse the template only
reportgo -template report.xml -validate
```

`-validate` currently checks that the template can be loaded and parsed. It does not run XSD validation.

## Project Structure

```text
reportgo/
├── cmd/reportgo/          # CLI application
├── internal/
│   ├── engine/            # PDF generation engine and renderers
│   ├── parser/            # XML template and JSON data parsing
│   └── models/            # Template and style models
├── pkg/reportgo/          # Public API
├── schemas/               # XML schema file
├── templates/examples/    # Example templates
└── examples/              # Usage examples
```

## Template Format

Templates are defined in XML. A schema file is included in [schemas/reportgo.xsd](/Users/dannys/repos/reportgo/schemas/reportgo.xsd), but the current parser relies on Go XML unmarshaling and does not enforce XSD validation.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4">
        <margins top="15" right="15" bottom="15" left="15"/>
    </document>

    <styles>
        <style name="base_text">
            <fontFamily>Arial</fontFamily>
            <fontSize>11</fontSize>
        </style>

        <style name="title" extends="base_text">
            <fontStyle>B</fontStyle>
            <fontSize>24</fontSize>
        </style>
    </styles>

    <sections>
        <section name="main">
            <text style="title">{{.Title}}</text>
        </section>
    </sections>
</report>
```

See [DESIGN.md](/Users/dannys/repos/reportgo/DESIGN.md) for the current design and element reference.

## Template Capabilities

Sections support `condition`, `loop`, `loopVariable`, `paddingLeft`, `pageBreakBefore`, and `pageBreakAfter`.

Rows provide horizontal flow layout for `text` and `image` children only. `rowgrid` splits the available width into equal columns and lets each column stack multiple child elements. Spacer elements add vertical space without drawing content.

## Template Functions

Built-in helpers include `upper`, `lower`, `title`, `trim`, `default`, `add`, `sub`, `mul`, `div`, `join`, `replace`, `ifelse`, `truncate`, `formatDate`, `dateFormat`, `formatNumber`, `formatCurrency`, and `formatPercent`.

Applications can register additional helpers with `reportgo.WithFuncMap(...)` or `engine.AddFuncMap(...)`.

`default` takes the fallback value first and the candidate value second:

```gotemplate
{{default "N/A" .Value}}
```

## Embedded Fonts

Applications can register in-memory fonts with `reportgo.WithEmbeddedFont(...)` or `engine.AddEmbeddedFont(...)`. These fonts are loaded before file-based fonts declared in the template.

## Current Limitations

- The bundled XSD file is not enforced during template loading.
- Built-in file loading supports JSON data only.
- `WithFontPath`, `WithImagePath`, `WithCompression`, and `WithSchemaValidation` are present in the public API but are not applied by the renderer yet.
- `document.customSize` is parsed into the model but is not used when initializing the PDF document.
- Header and footer templates render against the supplied data map only; implicit fields such as page number totals are not injected automatically.

## Dependencies

- [gofpdf](https://github.com/phpdave11/gofpdf) - PDF generation

## License

MIT License
