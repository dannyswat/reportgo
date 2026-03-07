# ReportGo — Improvement Analysis

Based on hands-on integration of `github.com/dannyswat/reportgo` into the CV export module of this project. The library provides a clean XML-template-to-PDF pipeline, but several gaps were encountered that forced workarounds or reduced output quality.

---

## 1. No Side-by-Side / Column Layout

**Severity:** High

There is no way to place two elements on the same horizontal line (e.g., job title left-aligned and date right-aligned). Every element occupies the full content width and advances the Y cursor downward.

**Workaround used:** Concatenated title and date into a single string separated by `·`, losing the clean left/right alignment the original `fpdf`-based implementation had.

**Suggested improvement:** Add a `<row>` or `<columns>` container element that lays out child elements horizontally:

```xml
<row>
    <text style="title_bold" width="130" align="L">{{.JobTitle}}</text>
    <text style="subtitle" width="50" align="R">{{.DateRange}}</text>
</row>
```

---

## 2. Element-Level Conditions Are Not Evaluated

**Severity:** High

The `BaseElement` struct defines a `Condition` attribute and every element type exposes `GetCondition()`, but `renderSection` in `engine.go` never checks element conditions before rendering. Only **section-level** `Condition` is documented/shown in examples — individual elements inside a section are always rendered.

**Impact:** Cannot conditionally show/hide a single text line within a section. Must instead create a separate `<section>` per conditional element, bloating the template.

**Suggested improvement:** In the element rendering loop inside `renderSection`, check `elem.GetCondition()` and skip rendering when the condition evaluates to empty/false:

```go
for _, elem := range section.Elements {
    if cond := getCondition(elem); cond != "" {
        result := e.processTemplate(cond)
        if result == "" || result == "false" {
            continue
        }
    }
    // ... render element
}
```

---

## 3. Section Loop/LoopVariable Not Implemented in Engine

**Severity:** High

The `Section` model parses `loop` and `loopVariable` attributes, and the DESIGN.md documents their usage:

```xml
<section name="products" loop="{{.Products}}" loopVariable="product">
    <text style="heading2">{{.product.Name}}</text>
</section>
```

However, `renderSection` in the engine does not contain any loop expansion logic — it renders each section exactly once. The `Loop` and `LoopVariable` fields are parsed but silently ignored.

**Impact:** Cannot use data-driven iteration from templates. All repetition must be handled by the caller generating the XML dynamically (which is what this project does), defeating a key benefit of a template engine.

**Suggested improvement:** Before rendering a section's elements, check if `section.Loop` is set. If so, resolve the data array, then iterate and render the section body once per item, temporarily merging the loop variable into the data context.

---

## 4. No Indentation / Padding on Elements or Sections

**Severity:** Medium

There is no `paddingLeft`, `marginLeft`, or `indent` attribute on `<text>` or `<section>` elements. The only way to indent content is to use absolute `x` positioning, which breaks the normal document flow (since the element then doesn't participate in auto page-break calculations the same way).

**Workaround used:** Removed indentation from the project listings entirely, losing the visual hierarchy from the original implementation.

**Suggested improvement:** Add `paddingLeft` / `marginLeft` to `BaseElement` or to `Section`, so content inside can be offset from the left margin:

```xml
<section name="project" paddingLeft="4">
    <text style="project_title">{{.ProjectName}}</text>
</section>
```

---

## 5. No Dedicated Spacer / Vertical Spacing Element

**Severity:** Medium

The only way to insert vertical space is:
- `spacingAfter` on an element (requires rendering actual content first)
- A dummy `<text>` element with tiny font and a space character

There is no empty spacing element like `<spacer height="3"/>`.

**Workaround used:** Created a `spacer` style with 1pt font and rendered `<text style="spacer" spacingAfter="3"> </text>` throughout the template.

**Suggested improvement:** Add a `<spacer>` element type:

```xml
<spacer height="3"/>
```

This would simply call `pdf.Ln(height)` with no content.

---

## 6. No Style Inheritance or Composition

**Severity:** Medium

Every style must fully specify all its properties. There is no `extends` or `basedOn` mechanism. This leads to heavy duplication when multiple styles share most properties but differ in one aspect (e.g., same font/size but different color, or same everything but bold).

**Example:** In this project, `title_bold`, `project_title`, and `cert_title` are nearly identical styles differing only in font size.

**Suggested improvement:** Add an `extends` attribute to `<style>`:

```xml
<style name="base_text">
    <fontFamily>Helvetica</fontFamily>
    <fontSize>9</fontSize>
    <textColor r="33" g="33" b="33"/>
</style>
<style name="title_bold" extends="base_text">
    <fontStyle>B</fontStyle>
    <fontSize>10</fontSize>
</style>
```

---

## 7. Line Element Does Not Advance Y Cursor

**Severity:** Low–Medium

After rendering a `<line>`, the Y cursor stays at the same position. This means any element immediately following the line will overlap it. You must insert a spacer after every line.

The current behavior may be intentional for decorative overlays, but for horizontal rules (the most common use case), it forces an extra spacer on every occurrence.

**Suggested improvement:** Add a `spacingAfter` attribute on `<line>` (it inherits `BaseElement` which already has it, but the renderer doesn't apply it). Apply `pdf.Ln(spacingAfter)` after drawing the line, consistent with how other elements handle it.

---

## 8. `GenerateToWriter` API Inconsistency

**Severity:** Low

`Generate(data, filepath)` accepts inline data and merges it, but `GenerateToWriter(w)` requires a separate `SetData()` call beforehand. This is an inconsistent API surface.

**Suggested improvement:** Add a `GenerateToWriter(data, w)` overload or make the existing method accept an optional data parameter:

```go
func (e *Engine) GenerateToWriter(w io.Writer, data ...map[string]interface{}) error
```

---

## 9. Limited Built-in Template Functions

**Severity:** Low

The template function map includes basic arithmetic and string operations but lacks common helpers useful for report generation:

| Missing Function | Use Case |
|---|---|
| `join` | Join a string slice with a separator |
| `len` | Get length of a slice/map for conditional rendering |
| `ifelse` / ternary | Inline conditional values |
| `dateFormat` | Parse and reformat date strings |
| `truncate` | Limit string length with ellipsis |
| `replace` | String replacement |

**Suggested improvement:** Expand the default `FuncMap`, or provide a `WithFuncMap(fmap)` option so callers can register custom functions.

---

## 10. No Support for Custom Fonts via Byte Slice

**Severity:** Low

`initPDF` loads fonts from file paths only (`os.ReadFile(font.File)`). There is no API to register fonts from embedded `[]byte` data (e.g., via Go's `embed` package), which is the preferred approach for containerized/serverless deployments where the filesystem may be read-only or ephemeral.

**Suggested improvement:** Add a `WithEmbeddedFont(name, family, style string, data []byte)` option, or allow fonts to be registered on the engine before template loading.

---

## 11. No Text Wrapping Width Control with Indentation

**Severity:** Low

When `wrap="true"` is set on a `<text>` element, `renderText` calculates available width as `pageWidth - marginLeft - marginRight`. If the element has an `x` offset, the wrapped text still uses the full margin-to-margin width, potentially overflowing the right margin.

**Suggested improvement:** When `wrap="true"` and `x > 0`, calculate available width as `pageWidth - x - marginRight` so wrapped text respects its starting position.

---

## Summary Table

| # | Issue | Severity | Status |
|---|---|---|---|
| 1 | No side-by-side / column layout | High | Missing feature |
| 2 | Element-level conditions not evaluated | High | Bug (parsed but ignored) |
| 3 | Section loop/loopVariable not implemented | High | Bug (parsed but ignored) |
| 4 | No indentation / padding support | Medium | Missing feature |
| 5 | No spacer element | Medium | Missing feature |
| 6 | No style inheritance | Medium | Missing feature |
| 7 | Line doesn't advance Y / apply spacingAfter | Low–Medium | Bug |
| 8 | GenerateToWriter API inconsistency | Low | Design issue |
| 9 | Limited template functions | Low | Missing feature |
| 10 | No embedded font support | Low | Missing feature |
| 11 | Text wrap ignores x offset | Low | Bug |
