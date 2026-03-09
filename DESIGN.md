# ReportGo Design

## Overview

ReportGo generates PDF documents from XML templates and data supplied either as JSON files or in-memory Go maps. The implementation is built on top of [gofpdf](https://github.com/phpdave11/gofpdf).

The current codebase is centered on three stages:

1. Parse an XML template into internal models.
2. Merge runtime data and evaluate Go template expressions while rendering.
3. Render the resulting document through gofpdf.

## Current Architecture

```text
XML Template ──> parser ──> report models ──> engine ──> gofpdf ──> PDF
                              ▲                ▲
                              │                │
                        JSON data file     in-memory data map
```

Notes:

- Templates are XML only.
- Built-in file-based data loading is JSON only.
- A schema file exists in the repository, but template loading currently uses XML unmarshaling rather than XSD enforcement.

## Implemented Capabilities

### Document Model

The report model currently supports:

- `metadata`
- `document` with margins, orientation, unit, and format
- `fonts`
- `styles`
- `header`
- `footer`
- `sections`

Default values are applied during parsing:

- orientation: `portrait`
- unit: `mm`
- format: `A4`
- margins: `15mm` on all sides when omitted
- header/footer height: `15mm` when the block exists and no height is set

### Supported Elements

Sections preserve element order and can contain:

- `text`
- `image`
- `table`
- `list`
- `keyValueList`
- `line`
- `rectangle`
- `row`
- `rowgrid`
- `spacer`
- `pageBreak`

Rows are intentionally narrow in scope:

- only `text` and `image` children are supported
- `image` children must provide both `width` and `height`
- `text` children may omit `width`; the last text child expands to remaining width
- child `x` and `y` offsets are interpreted relative to the row origin

Rowgrids divide the current flow width into equal columns:

- `columns` controls the number of equal parts
- each `<col>` can contain the same flow elements that a section can contain
- the overall rowgrid height is the tallest rendered column

### Styling

Styles are named and reusable. They currently support:

- `fontFamily`
- `fontStyle`
- `fontSize`
- `textColor`
- `fillColor`
- `align`
- `lineHeight`
- `extends`

Style inheritance is resolved during template parsing. Unknown parent styles and inheritance cycles return errors.

### Data Binding and Helpers

Content fields are processed with Go's `text/template` package.

Built-in helpers currently include:

- `upper`
- `lower`
- `title`
- `trim`
- `default`
- `add`
- `sub`
- `mul`
- `div`
- `join`
- `replace`
- `ifelse`
- `truncate`
- `formatDate`
- `dateFormat`
- `formatNumber`
- `formatCurrency`
- `formatPercent`

Applications can extend the helper set via `WithFuncMap` at construction time or `AddFuncMap` afterward.

The `default` helper uses the signature `default fallback value`.

### Conditional Rendering and Loops

Conditional rendering is supported on sections and individual elements through the `condition` attribute.

After template expansion, these rendered values are treated as false:

- empty string
- `0`
- `false`
- `nil`
- `null`
- `<no value>`

Any other rendered value is treated as true.

Sections also support repeating content with:

- `loop`
- `loopVariable`

If `loopVariable` is omitted, the current item is exposed as `item`.

### Layout Behavior

The engine supports a mix of flow-based and positioned rendering:

- text and images can use explicit `x` and `y` coordinates
- wrapped text uses the current effective content width rather than raw page width
- sections can shift flow content with `paddingLeft`
- rowgrids constrain nested flow content to each column's width
- sections support `pageBreakBefore` and `pageBreakAfter`
- `spacer` advances the cursor without drawing
- `pageBreak` forces a new page immediately

Headers and footers can render `text`, `image`, and `line` elements when enabled.

### Fonts

Fonts can be provided in two ways:

- file-based fonts declared in the template `<fonts>` block
- embedded in-memory fonts registered through the public API

Embedded fonts are loaded first. File-based fonts are skipped when a font with the same family and style has already been registered from memory.

## Template Structure

The following example reflects the features that are actually implemented today:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<report xmlns="http://reportgo.io/schema/v1" version="1.0">
    <document orientation="portrait" unit="mm" format="A4">
        <margins top="15" right="15" bottom="15" left="15"/>
    </document>

    <styles>
        <style name="base_text">
            <fontFamily>Arial</fontFamily>
            <fontSize>11</fontSize>
            <textColor r="0" g="0" b="0"/>
            <align>L</align>
            <lineHeight>6</lineHeight>
        </style>

        <style name="title" extends="base_text">
            <fontStyle>B</fontStyle>
            <fontSize>20</fontSize>
        </style>

        <style name="table_header" extends="base_text">
            <fontStyle>B</fontStyle>
            <fontSize>10</fontSize>
            <textColor r="255" g="255" b="255"/>
            <fillColor r="0" g="51" b="102"/>
            <align>C</align>
        </style>
    </styles>

    <sections>
        <section name="summary">
            <text style="title" spacingAfter="6">{{.Title}}</text>
            <text style="base_text" wrap="true" spacingAfter="8">{{.Summary}}</text>
        </section>

        <section name="highlights" paddingLeft="8" condition="{{.ShowHighlights}}">
            <row spacingAfter="3">
                <text style="base_text" width="120">{{.PrimaryLabel}}</text>
                <text style="base_text" width="50" align="R">{{.PrimaryValue}}</text>
            </row>
            <spacer height="2"/>
            <list style="base_text" items="{{.Highlights}}" bullet="-" indent="6"/>
        </section>

        <section name="projects" loop="{{.Projects}}" loopVariable="project">
            <text style="title">{{title .project.Name}}</text>
            <text style="base_text" wrap="true">{{default "No description" .project.Description}}</text>
        </section>
    </sections>
</report>
```

## Element Reference

### Text

```xml
<text style="body" x="10" y="20" width="80" align="L" wrap="true" spacingAfter="6">{{.Value}}</text>
```

Supported attributes:

- `style`
- `x`
- `y`
- `width`
- `align`
- `wrap`
- `condition`
- `spacingAfter`

### Image

```xml
<image path="{{.ImagePath}}" width="50" height="30" align="C" spacingAfter="8"/>
```

Supported attributes:

- `path`
- `x`
- `y`
- `width`
- `height`
- `align`
- `condition`
- `spacingAfter`

### Table

```xml
<table dataSource="{{.Rows}}" headerStyle="table_header" cellStyle="body" border="true">
    <alternateRowColor r="245" g="245" b="245"/>
    <columns>
        <column header="Name" field="name" width="80" align="L"/>
        <column header="Value" field="value" width="30" align="R" format="number"/>
    </columns>
</table>
```

Supported attributes:

- `dataSource`
- `headerStyle`
- `cellStyle`
- `border`
- `condition`
- `spacingAfter`

Column `format` values currently handled by the renderer are `currency`, `percent`, `number`, and the empty default case.

### List

```xml
<list style="body" items="{{.Items}}" bullet="-" indent="10" spacingAfter="8"/>
```

Supported attributes:

- `items`
- `style`
- `bullet`
- `indent`
- `condition`
- `spacingAfter`

### Key-Value List

```xml
<keyValueList style="body" keyWidth="50" valueAlign="R">
    <item key="Revenue" value="{{.Revenue}}"/>
</keyValueList>
```

Supported attributes:

- `style`
- `keyWidth`
- `valueWidth`
- `valueAlign`
- `condition`
- `spacingAfter`

### Line

```xml
<line x1="15" y1="0" x2="-15" y2="0" color="#D9D1C7" width="0.3" spacingAfter="4"/>
```

Supported attributes:

- `x1`
- `y1`
- `x2`
- `y2`
- `color`
- `width`
- `condition`
- `spacingAfter`

Negative `x` coordinates are resolved relative to the page width.

### Rectangle

```xml
<rectangle x="15" y="30" width="100" height="20" radius="2">
    <fillColor r="245" g="245" b="245"/>
    <borderColor r="0" g="0" b="0"/>
    <borderWidth>0.5</borderWidth>
</rectangle>
```

Supported attributes:

- `x`
- `y`
- `width`
- `height`
- `radius`
- `condition`
- `spacingAfter`

### Row

```xml
<row spacingAfter="2">
    <text style="body">Left</text>
    <text style="body" align="R">Right</text>
</row>
```

Supported attributes:

- `condition`
- `spacingAfter`

### RowGrid

```xml
<rowgrid columns="3" spacingAfter="4">
    <col>
        <text style="body">First</text>
        <text style="body" wrap="true">More detail</text>
    </col>
    <col>
        <text style="body">Second</text>
    </col>
    <col>
        <text style="body">Third</text>
    </col>
</rowgrid>
```

Supported attributes:

- `columns`
- `condition`
- `spacingAfter`

### Spacer

```xml
<spacer height="3" spacingAfter="2"/>
```

Supported attributes:

- `height`
- `condition`
- `spacingAfter`

### Page Break

```xml
<pageBreak condition="{{.ForceBreak}}"/>
```

Supported attributes:

- `condition`

## Public API

### Typical Flow

```go
engine := reportgo.New(reportgo.WithFuncMap(template.FuncMap{
    "badge": func(input string) string { return "[" + strings.ToUpper(input) + "]" },
}))

if err := engine.LoadTemplate("templates/report.xml"); err != nil {
    return err
}

if err := engine.LoadDataFromFile("data.json"); err != nil {
    return err
}

return engine.Generate(nil, "output.pdf")
```

### Writer Output

```go
var buf bytes.Buffer
if err := engine.GenerateToWriter(&buf, map[string]interface{}{"Title": "Inline"}); err != nil {
    return err
}
```

### Embedded Fonts

```go
engine := reportgo.New(
    reportgo.WithEmbeddedFont("noto-regular", "Noto Sans", "", fontBytes),
)
```

## Current Gaps

These behaviors should be treated as not implemented yet, even though parts of the API or model exist:

- XSD validation is not executed during template loading.
- YAML data input is not supported.
- `WithFontPath`, `WithImagePath`, `WithCompression`, and `WithSchemaValidation` do not currently change renderer behavior.
- `document.customSize` is parsed but not used when creating the PDF instance.
- No implicit page-number variables are injected into template data.

## Repository Layout

```text
reportgo/
├── cmd/reportgo/               # CLI entry point
├── examples/                   # Sample programs and data
├── fonts/                      # Example font assets used by examples/tests
├── internal/engine/            # Engine logic and renderers
├── internal/models/            # XML-backed report models
├── internal/parser/            # Template and data parsing
├── pkg/reportgo/               # Public API package
├── schemas/reportgo.xsd        # Reference schema file
└── templates/examples/         # Example XML templates
```
