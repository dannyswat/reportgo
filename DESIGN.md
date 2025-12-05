# ReportGo - PDF Report Generator Design Document

## Overview

ReportGo is a Go-based PDF report generation system that creates professional PDF reports from file templates and data. It uses the [gofpdf](https://github.com/phpdave11/gofpdf) library for PDF generation.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Template File  │────▶│  Report Engine   │────▶│   PDF Output    │
│    (YAML/JSON)  │     │                  │     │                 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               ▲
                               │
                        ┌──────┴──────┐
                        │  Data Source │
                        │  (JSON/YAML) │
                        └─────────────┘
```

## Template Format Specification

Templates are defined in XML format with schema validation via XSD. The template defines the structure, layout, and styling of the PDF report.

### Template Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1"
        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
        xsi:schemaLocation="http://reportgo.io/schema/v1 reportgo.xsd"
        version="1.0">
    
    <metadata>
        <name>Monthly Sales Report</name>
        <description>Template for monthly sales reports</description>
        <author>ReportGo</author>
    </metadata>
    
    <document orientation="portrait" unit="mm" format="A4">
        <margins top="15" right="15" bottom="15" left="15"/>
    </document>
    
    <fonts>
        <font name="custom-font" family="CustomFont" style="" file="fonts/CustomFont-Regular.ttf"/>
    </fonts>
    
    <styles>
        <style name="title">
            <fontFamily>Arial</fontFamily>
            <fontStyle>B</fontStyle>
            <fontSize>24</fontSize>
            <textColor r="0" g="51" b="102"/>
            <align>C</align>
        </style>
        
        <style name="heading1">
            <fontFamily>Arial</fontFamily>
            <fontStyle>B</fontStyle>
            <fontSize>16</fontSize>
            <textColor r="0" g="0" b="0"/>
            <align>L</align>
        </style>
        
        <style name="heading2">
            <fontFamily>Arial</fontFamily>
            <fontStyle>B</fontStyle>
            <fontSize>14</fontSize>
            <textColor r="51" g="51" b="51"/>
            <align>L</align>
        </style>
        
        <style name="body">
            <fontFamily>Arial</fontFamily>
            <fontStyle></fontStyle>
            <fontSize>11</fontSize>
            <textColor r="0" g="0" b="0"/>
            <align>L</align>
            <lineHeight>6</lineHeight>
        </style>
        
        <style name="table_header">
            <fontFamily>Arial</fontFamily>
            <fontStyle>B</fontStyle>
            <fontSize>10</fontSize>
            <textColor r="255" g="255" b="255"/>
            <fillColor r="0" g="51" b="102"/>
            <align>C</align>
        </style>
        
        <style name="table_cell">
            <fontFamily>Arial</fontFamily>
            <fontStyle></fontStyle>
            <fontSize>10</fontSize>
            <textColor r="0" g="0" b="0"/>
            <fillColor r="245" g="245" b="245"/>
            <align>L</align>
        </style>
    </styles>
    
    <header enabled="true" height="20">
        <image path="{{.LogoPath}}" x="10" y="5" width="30" height="10"/>
        <text style="heading2" x="50" y="10">{{.CompanyName}}</text>
        <text style="body" x="-50" y="10" align="R">Generated: {{.GeneratedDate}}</text>
    </header>
    
    <footer enabled="true" height="15">
        <line x1="10" y1="0" x2="-10" y2="0" color="#C8C8C8"/>
        <text style="body" align="C" y="5">Page {{.PageNum}} of {{.TotalPages}}</text>
        <text style="body" align="R" x="-10" y="5">Confidential</text>
    </footer>
    
    <sections>
        <section name="title_section">
            <text style="title" spacingAfter="10">{{.ReportTitle}}</text>
            <text style="heading2" align="C" spacingAfter="20">{{.ReportSubtitle}}</text>
        </section>
        
        <section name="summary_section">
            <text style="heading1" spacingAfter="5">Executive Summary</text>
            <text style="body" spacingAfter="15">{{.Summary}}</text>
            <keyValueList style="body" spacingAfter="20">
                <item key="Total Revenue" value="{{.TotalRevenue}}"/>
                <item key="Total Orders" value="{{.TotalOrders}}"/>
                <item key="Average Order Value" value="{{.AvgOrderValue}}"/>
            </keyValueList>
        </section>
        
        <section name="data_table_section">
            <text style="heading1" spacingAfter="5">Sales Details</text>
            <table dataSource="{{.SalesData}}" headerStyle="table_header" 
                   cellStyle="table_cell" border="true" spacingAfter="20">
                <alternateRowColor r="255" g="255" b="255"/>
                <columns>
                    <column header="Product" field="product" width="50" align="L"/>
                    <column header="Quantity" field="quantity" width="30" align="C"/>
                    <column header="Unit Price" field="unit_price" width="35" align="R" format="currency"/>
                    <column header="Total" field="total" width="35" align="R" format="currency"/>
                </columns>
            </table>
        </section>
        
        <section name="chart_section">
            <text style="heading1" spacingAfter="5">Sales Trend</text>
            <image path="{{.ChartImagePath}}" width="150" height="80" align="C" spacingAfter="15"/>
        </section>
        
        <section name="notes_section" pageBreakBefore="true">
            <text style="heading1" spacingAfter="5">Notes &amp; Observations</text>
            <list style="body" items="{{.Notes}}" bullet="•" indent="10" spacingAfter="15"/>
        </section>
    </sections>
</report>
```

## Data Format Specification

Data is provided in JSON or YAML format:

```json
{
  "LogoPath": "assets/logo.png",
  "CompanyName": "Acme Corporation",
  "GeneratedDate": "2025-12-05",
  "ReportTitle": "Monthly Sales Report",
  "ReportSubtitle": "November 2025",
  "Summary": "This report provides a comprehensive overview of sales performance for November 2025. Overall revenue increased by 15% compared to the previous month.",
  "TotalRevenue": "$125,450.00",
  "TotalOrders": "1,234",
  "AvgOrderValue": "$101.66",
  "SalesData": [
    {
      "product": "Widget A",
      "quantity": 500,
      "unit_price": 25.00,
      "total": 12500.00
    },
    {
      "product": "Widget B",
      "quantity": 350,
      "unit_price": 45.00,
      "total": 15750.00
    }
  ],
  "ChartImagePath": "assets/chart.png",
  "Notes": [
    "Sales exceeded target by 10%",
    "New product launch contributed 25% of revenue",
    "Customer retention rate improved to 85%"
  ]
}
```

## Element Types

### Text Element

```xml
<text style="style_name" x="10" y="20" align="L" spacingAfter="10">
    {{.Variable}}
</text>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `style` | Reference to defined style | No |
| `x` | Absolute X position | No |
| `y` | Absolute Y position | No |
| `align` | L \| C \| R \| J | No |
| `spacingAfter` | Vertical spacing after element | No |

### Image Element

```xml
<image path="{{.ImagePath}}" x="10" y="20" width="50" height="30" align="C" spacingAfter="10"/>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `path` | Path to image file | Yes |
| `x` | X position | No |
| `y` | Y position | No |
| `width` | Image width | Yes |
| `height` | Image height (maintains aspect ratio if omitted) | No |
| `align` | L \| C \| R | No |
| `spacingAfter` | Vertical spacing after element | No |

### Table Element

```xml
<table dataSource="{{.DataArray}}" headerStyle="table_header" 
       cellStyle="table_cell" border="true" spacingAfter="15">
    <alternateRowColor r="240" g="240" b="240"/>
    <columns>
        <column header="Column Name" field="field_name" width="40" align="L" format="string"/>
    </columns>
</table>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `dataSource` | Template variable for data array | Yes |
| `headerStyle` | Style for header row | No |
| `cellStyle` | Style for data cells | No |
| `border` | Show borders (true/false) | No |
| `spacingAfter` | Vertical spacing after element | No |

Column `format` options: `string` | `currency` | `number` | `date` | `percent`

### List Element

```xml
<list style="body" items="{{.ListItems}}" bullet="•" indent="10" spacingAfter="10"/>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `items` | Template variable for list items | Yes |
| `style` | Reference to defined style | No |
| `bullet` | Bullet character | No |
| `indent` | Left indentation | No |
| `spacingAfter` | Vertical spacing after element | No |

### Key-Value List Element

```xml
<keyValueList style="body" keyWidth="50" spacingAfter="10">
    <item key="Label" value="{{.Value}}"/>
</keyValueList>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `style` | Reference to defined style | No |
| `keyWidth` | Width for key column | No |
| `spacingAfter` | Vertical spacing after element | No |

### Line Element

```xml
<line x1="10" y1="0" x2="-10" y2="0" color="#000000" width="0.5"/>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `x1`, `y1` | Start coordinates | Yes |
| `x2`, `y2` | End coordinates (negative = relative to margin) | Yes |
| `color` | Line color (hex or rgb) | No |
| `width` | Line width in mm | No |

### Rectangle Element

```xml
<rectangle x="10" y="20" width="100" height="50" radius="3">
    <fillColor r="240" g="240" b="240"/>
    <borderColor r="0" g="0" b="0"/>
    <borderWidth>0.5</borderWidth>
</rectangle>
```

| Attribute | Description | Required |
|-----------|-------------|----------|
| `x`, `y` | Position | Yes |
| `width`, `height` | Dimensions | Yes |
| `radius` | Corner radius for rounded rectangles | No |

### Page Break Element

```xml
<pageBreak/>
```

## Project Structure

```
reportgo/
├── cmd/
│   └── reportgo/
│       └── main.go           # CLI entry point
├── internal/
│   ├── engine/
│   │   ├── engine.go         # Core report generation engine
│   │   ├── renderer.go       # PDF rendering logic
│   │   └── elements.go       # Element rendering implementations
│   ├── parser/
│   │   ├── template.go       # XML template parsing
│   │   └── data.go           # Data parsing
│   └── models/
│       ├── template.go       # Template data structures
│       ├── document.go       # Document configuration
│       ├── elements.go       # Element definitions
│       └── styles.go         # Style definitions
├── pkg/
│   └── reportgo/
│       └── api.go            # Public API
├── schemas/
│   └── reportgo.xsd          # XML Schema Definition
├── templates/
│   └── examples/
│       ├── invoice.xml       # Invoice template example
│       └── report.xml        # Report template example
├── examples/
│   ├── simple/
│   │   └── main.go           # Simple usage example
│   └── advanced/
│       └── main.go           # Advanced usage example
├── go.mod
├── go.sum
├── README.md
└── DESIGN.md
```

## API Design

### Basic Usage

```go
package main

import (
    "log"
    "github.com/yourusername/reportgo/pkg/reportgo"
)

func main() {
    // Create a new report engine
    engine := reportgo.New()

    // Load XML template
    err := engine.LoadTemplate("templates/report.xml")
    if err != nil {
        log.Fatal(err)
    }

    // Load data
    data := map[string]interface{}{
        "ReportTitle": "Monthly Sales Report",
        "Summary":     "Executive summary content...",
        "SalesData": []map[string]interface{}{
            {"product": "Widget A", "quantity": 100, "total": 2500.00},
            {"product": "Widget B", "quantity": 50, "total": 1250.00},
        },
    }

    // Generate PDF
    err = engine.Generate(data, "output/report.pdf")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Usage

```go
package main

import (
    "log"
    "github.com/yourusername/reportgo/pkg/reportgo"
)

func main() {
    // Create engine with options
    engine := reportgo.New(
        reportgo.WithFontPath("./fonts"),
        reportgo.WithImagePath("./assets"),
        reportgo.WithCompression(true),
        reportgo.WithSchemaValidation(true),  // Enable XSD validation
    )

    // Load template from string
    templateXML := `<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" format="A4"/>
    <sections>
        <section name="main">
            <text style="title">{{.Title}}</text>
        </section>
    </sections>
</report>`

    err := engine.LoadTemplateFromString(templateXML)
    if err != nil {
        log.Fatal(err)
    }

    // Load data from JSON file
    err = engine.LoadDataFromFile("data/report-data.json")
    if err != nil {
        log.Fatal(err)
    }

    // Generate to io.Writer
    file, _ := os.Create("output/report.pdf")
    defer file.Close()
    
    err = engine.GenerateToWriter(file)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Template Functions

Built-in template functions available in content fields:

| Function | Description | Example |
|----------|-------------|---------|
| `formatDate` | Format date string | `{{formatDate .Date "2006-01-02"}}` |
| `formatNumber` | Format number with decimals | `{{formatNumber .Value 2}}` |
| `formatCurrency` | Format as currency | `{{formatCurrency .Amount "USD"}}` |
| `formatPercent` | Format as percentage | `{{formatPercent .Rate}}` |
| `upper` | Convert to uppercase | `{{upper .Name}}` |
| `lower` | Convert to lowercase | `{{lower .Name}}` |
| `title` | Convert to title case | `{{title .Name}}` |
| `trim` | Trim whitespace | `{{trim .Text}}` |
| `default` | Provide default value | `{{default .Value "N/A"}}` |
| `add` | Add numbers | `{{add .A .B}}` |
| `sub` | Subtract numbers | `{{sub .A .B}}` |
| `mul` | Multiply numbers | `{{mul .A .B}}` |
| `div` | Divide numbers | `{{div .A .B}}` |

## Conditional Sections

Sections and elements can be conditionally rendered using the `condition` attribute:

```xml
<section name="optional_section" condition="{{if .ShowSection}}true{{end}}">
    <text style="body">This section is optional</text>
</section>
```

## Loops in Templates

For repeating content based on data, use the `loop` and `loopVariable` attributes:

```xml
<section name="products" loop="{{.Products}}" loopVariable="product">
    <text style="heading2">{{.product.Name}}</text>
    <text style="body">{{.product.Description}}</text>
</section>
```

## Color Formats

Colors can be specified in multiple formats:

```xml
<!-- RGB elements -->
<textColor r="255" g="100" b="50"/>

<!-- Hex string attribute -->
<line color="#FF6432" .../>

<!-- Named color attribute -->
<line color="blue" .../>
```

## Supported Named Colors

| Color Name | RGB Value |
|------------|-----------|
| black | [0, 0, 0] |
| white | [255, 255, 255] |
| red | [255, 0, 0] |
| green | [0, 128, 0] |
| blue | [0, 0, 255] |
| gray | [128, 128, 128] |
| lightgray | [211, 211, 211] |
| darkgray | [169, 169, 169] |

## Page Formats

| Format | Size (mm) |
|--------|-----------|
| A3 | 297 x 420 |
| A4 | 210 x 297 |
| A5 | 148 x 210 |
| Letter | 215.9 x 279.4 |
| Legal | 215.9 x 355.6 |
| Custom | [width, height] |

## Error Handling

The engine provides comprehensive error handling:

```go
err := engine.Generate(data, "output.pdf")
if err != nil {
    switch e := err.(type) {
    case *reportgo.TemplateError:
        log.Printf("Template error: %v at line %d", e.Message, e.Line)
    case *reportgo.DataError:
        log.Printf("Data error: %v for field %s", e.Message, e.Field)
    case *reportgo.RenderError:
        log.Printf("Render error: %v in section %s", e.Message, e.Section)
    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

## Future Enhancements

1. **Chart Generation**: Built-in chart rendering (bar, line, pie charts)
2. **QR/Barcode Support**: Generate QR codes and barcodes
3. **Watermarks**: Add text or image watermarks
4. **Digital Signatures**: Support for PDF digital signatures
5. **Password Protection**: PDF encryption and password protection
6. **Template Inheritance**: Base templates with overridable sections
7. **Multiple Output Formats**: Export to HTML, DOCX in addition to PDF
8. **Web API**: REST API for report generation
9. **Template Validation**: CLI tool for validating templates
10. **Live Preview**: Hot-reload template preview during development

## Dependencies

- `github.com/phpdave11/gofpdf` - PDF generation
- `encoding/xml` - XML parsing (standard library)
- `encoding/json` - JSON parsing (standard library)
- `text/template` - Template processing (standard library)

## License

MIT License
