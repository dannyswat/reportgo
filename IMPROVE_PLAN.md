# ReportGo Improvement Spec and Implementation Plan

## Purpose

This document reviews the recommendations in `REPORTGO_ANALYSIS.md` against the current codebase and turns them into an implementable improvement plan.

The goals are to:

- fix features that are already modeled in XML/schema but ignored at runtime,
- add the minimum new layout primitives needed to remove common template workarounds,
- improve API extensibility without breaking existing callers,
- add regression coverage before expanding the rendering surface further.

## Review of Recommendations

| # | Recommendation | Review | Decision |
|---|---|---|---|
| 1 | Side-by-side / column layout | Valid gap. Current engine renders section elements as a vertical stream only. A generic layout container is needed. | Accept, but use a new `<row>` container rather than `<columns>` to avoid conflict with table columns. |
| 2 | Element-level conditions | Valid, but understated. Element `condition` is ignored, and section `condition` is also parsed but never evaluated. | Accept and broaden to a single conditional-rendering implementation for both sections and elements. |
| 3 | Section `loop` / `loopVariable` | Confirmed. Attributes exist in model and schema but are never executed in the engine. | Accept as a high-priority correctness fix. |
| 4 | Indentation / padding | Valid need. Absolute `x` positioning is not a substitute for flow-aware indentation. | Accept, but start with section-level left padding and row-local padding instead of broad per-element margin support. |
| 5 | Spacer element | Valid and low-risk. Current workaround is artificial text content. | Accept. |
| 6 | Style inheritance | Valid, but this is design work rather than a correctness fix. Needs resolution order and cycle handling. | Accept for phase 3. |
| 7 | Line spacing / cursor advance | Confirmed. `Line` inherits `SpacingAfter`, but renderer does not apply it. | Accept as a small correctness fix. Preserve current overlay behavior when `spacingAfter="0"`. |
| 8 | `GenerateToWriter` inconsistency | Valid issue, but Go does not support overloads. | Accept with a Go-idiomatic API shape: variadic optional data or a new helper method, while preserving compatibility. |
| 9 | Limited template functions | Partially valid. `len` already exists as a built-in `text/template` function, so it is not missing. Extensibility is the bigger gap. | Accept in modified form: add custom function registration plus a small curated helper set. |
| 10 | Embedded font support | Valid. The engine already calls `AddUTF8FontFromBytes`, so the missing piece is public registration of in-memory font data. | Accept as an API enhancement. |
| 11 | Wrap width ignores indentation / X offset | Valid, but the more robust fix is to base wrap width on the effective current X position, not only the explicit `x` attribute. | Accept in modified form. |

## Priority Summary

### P0: correctness gaps already advertised by schema or models

- Section `condition`
- Element `condition`
- Section `loop` and `loopVariable`
- `line.spacingAfter`
- Wrap width based on effective cursor X

These are the highest-priority items because the template format already implies support, but the runtime silently ignores them.

### P1: missing layout primitives needed for real documents

- `<row>` horizontal layout container
- Section-level left padding
- `<spacer>` element

These unlock cleaner templates and remove the most visible authoring workarounds.

### P2: extensibility and API cleanup

- Style inheritance via `extends`
- Custom template function registration
- Writer API consistency
- Embedded font registration

These are useful, but they should not delay the correctness and layout work.

## Improvement Spec

## 1. Conditional Rendering

### Scope

- Evaluate `section.condition` before any section rendering.
- Evaluate `condition` on every section element before dispatching to its renderer.

### Truthiness Rules

To keep templates predictable, conditional evaluation should use one shared helper with these results treated as false:

- empty string
- string value `"false"`
- string value `"0"`
- string value `"nil"`

Everything else is treated as true.

### Design Notes

- Conditions should be processed through the existing template engine so they can reference data and functions.
- The same logic must apply to `pageBreak` elements.
- No schema change is required; the attributes already exist.

### Acceptance Criteria

- A section with `condition="{{.ShowSection}}"` is skipped when the rendered value is false-like.
- A text element with `condition` is skipped independently of sibling elements.
- Existing templates without conditions behave exactly as before.

## 2. Section Loop Execution

### Scope

- Execute a section once per item when `loop` is set.
- Bind the current item under `loopVariable`.
- Restore the original data context after each iteration.

### Supported Data Shapes

- `[]map[string]interface{}`
- `[]interface{}` containing map-like or scalar values
- optionally slices of scalars, exposed through the loop variable directly

### Context Rules

- The original root data remains accessible.
- The loop variable is injected into the current render context for that iteration.
- Nested loops are not required for the first implementation, but the design should not block them.

### Design Notes

- Implement loop execution in `renderSection`, not in parsing.
- Introduce a small context helper to clone/extend `e.data` safely for each iteration.
- If `loopVariable` is empty, use a conservative default such as `item`.

### Acceptance Criteria

- A section with `loop="{{.Products}}" loopVariable="product"` renders once per product.
- `{{.product.Name}}` resolves correctly inside the loop body.
- Non-loop sections render once as they do today.

## 3. Flow-Aware Horizontal Layout

### Scope

Add a new `<row>` container that renders child elements on one horizontal line while keeping normal vertical flow.

Example:

```xml
<row spacingAfter="2">
    <text style="title_bold" width="130" align="L">{{.JobTitle}}</text>
    <text style="subtitle" width="50" align="R">{{.DateRange}}</text>
</row>
```

### Initial Supported Children

- `text`
- `image`
- `spacer` should not be allowed inside a row in phase 1

### Layout Rules

- Children render left-to-right.
- Each child must either define width explicitly or have a deterministic width rule.
- Row height is the max rendered height of its children.
- After the row completes, Y advances by row height plus `spacingAfter`.

### Design Notes

- Do not reuse the name `columns`; that term is already used by tables in the schema and documentation.
- Row rendering will likely require a dedicated renderer instead of reusing the vertical stream directly.
- Phase 1 should reject unsupported child types with a clear error instead of guessing behavior.

### Acceptance Criteria

- Two text nodes can be rendered on the same line with independent alignment.
- The next block starts below the tallest row child.
- Rows participate in normal page flow.

## 4. Indentation / Padding

### Scope

Start with section-level left padding:

```xml
<section name="project" paddingLeft="4">
    <text style="project_title">{{.ProjectName}}</text>
</section>
```

### Rationale

Section-level padding covers the most common use case and avoids a wide surface-area change across every element type.

### Layout Rules

- `paddingLeft` shifts the starting X position for flow-rendered content in that section.
- Wrapped text width is reduced accordingly.
- Nested rows should inherit the padded origin.

### Design Notes

- This requires schema and model updates.
- Implement as a temporary X offset in the section render context, not via absolute coordinates in each template.
- Defer full `marginRight`, per-element margin, and nested container padding until the row implementation settles.

### Acceptance Criteria

- A padded section indents flow content without requiring absolute `x` values.
- Wrapped text inside the section does not overflow the right margin.

## 5. Spacer Element

### Scope

Add:

```xml
<spacer height="3"/>
```

### Behavior

- Advances Y by `height` with no drawn content.
- May also support `condition` for consistency.

### Acceptance Criteria

- Templates no longer need dummy text nodes to create vertical spacing.

## 6. Style Inheritance

### Scope

Add `extends` to `<style>`:

```xml
<style name="title_bold" extends="base_text">
    <fontStyle>B</fontStyle>
</style>
```

### Resolution Rules

- Child values override parent values.
- Missing child values inherit from the parent.
- Multiple inheritance is out of scope.
- Cycles must produce a parse-time error.

### Design Notes

- Resolve style inheritance once after parsing and before rendering.
- Keep the runtime style lookup simple by storing resolved styles in the engine map.

### Acceptance Criteria

- A derived style inherits unspecified font, color, alignment, and line-height values.
- Cyclic style references fail clearly.

## 7. Line Spacing Behavior

### Scope

- Apply `spacingAfter` for `line` elements.
- Keep current overlay capability when spacing is omitted or zero.

### Acceptance Criteria

- A horizontal rule followed by text does not overlap when `spacingAfter` is set.

## 8. Writer API Consistency

### Scope

Keep the current method working, but make one-call generation with inline data possible.

Recommended shape:

```go
func (e *Engine) GenerateToWriter(w io.Writer, data ...map[string]interface{}) error
```

### Design Notes

- This is backward compatible for existing callers.
- Merge any provided map into engine data before generation.
- Update README examples to show both file and writer flows.

### Acceptance Criteria

- Callers can generate to a writer without a separate `SetData()` call.

## 9. Template Function Extensibility

### Scope

- Add public API for registering custom template functions.
- Add only a small curated default helper set.

Recommended additions:

- `join`
- `replace`
- `ifelse`
- `truncate`
- `dateFormat`

### Design Notes

- Do not add `len`; it already exists in Go templates.
- Prefer a merge API such as `WithFuncMap` or `AddFuncMap` so applications can provide domain-specific helpers.
- Avoid exposing unsafe or side-effecting functions.

### Acceptance Criteria

- Callers can register custom functions before rendering.
- Built-in helpers cover common string and formatting needs without overloading the default surface.

## 10. Embedded Font Registration

### Scope

Add a public API to register font bytes directly on the engine.

Possible shape:

```go
func WithEmbeddedFont(name, family, style string, data []byte) Option
```

### Design Notes

- The renderer already uses `AddUTF8FontFromBytes`; this is primarily an engine configuration change.
- File-based font configuration in XML must keep working.
- Decide whether embedded registrations augment or override template-defined fonts when family/style collide.

### Acceptance Criteria

- An application can render using `embed`-provided font bytes without reading font files from disk.

## 11. Wrap Width Calculation

### Scope

- When wrapping text, compute available width from the effective current X position after all positioning and padding adjustments.

### Rationale

This is more accurate than checking only `text.X > 0`, because the effective X may come from section padding, row layout, or prior positioning.

### Acceptance Criteria

- Wrapped text started from an indented X position stays within the right content boundary.

## Implementation Plan

## Phase 0: Regression Harness

### Deliverables

- Add the first `*_test.go` files.
- Create unit tests around engine behavior that can inspect page number, cursor Y, and context-sensitive rendering.
- Add small parser tests for style inheritance resolution and schema/model parsing where relevant.

### Suggested Tests

- Section condition skip
- Element condition skip
- Loop renders N times
- Line `spacingAfter` advances Y
- Wrapped text respects current X
- Spacer advances Y
- Style inheritance resolution and cycle detection

### Exit Criteria

- New work lands with regression coverage for each behavior change.

## Phase 1: Correctness Fixes

### Scope

- Conditional rendering for sections and elements
- Section loop execution
- Line spacing behavior
- Wrap width calculation

### Files Likely Affected

- `internal/engine/engine.go`
- `internal/engine/renderer.go`
- `internal/models/document.go`
- `internal/models/elements.go`

### Notes

- This phase should not add major new XML constructs.
- It should end with updated docs because some currently documented behavior is misleading.

### Exit Criteria

- All P0 items behave as documented.
- No public API breakage.

## Phase 2: Layout Primitives

### Scope

- `<row>` container
- `paddingLeft` on section
- `<spacer>` element

### Files Likely Affected

- `internal/models/document.go`
- `internal/models/elements.go`
- `internal/engine/engine.go`
- `internal/engine/renderer.go`
- `schemas/reportgo.xsd`
- `DESIGN.md`
- example templates under `templates/examples/`

### Notes

- Schema and examples must ship with the renderer changes.
- Keep the first row implementation narrow and deterministic.

### Exit Criteria

- Real templates can express left/right aligned content and explicit vertical gaps without dummy text nodes.

## Phase 3: Extensibility and API Cleanup

### Scope

- Style inheritance
- Custom template function registration
- Improved writer API
- Embedded font registration

### Files Likely Affected

- `internal/models/styles.go`
- `internal/parser/template.go`
- `internal/engine/engine.go`
- `pkg/reportgo/api.go`
- `README.md`
- `DESIGN.md`

### Notes

- Treat this as a compatibility-sensitive phase because it touches public API and template authoring conventions.

### Exit Criteria

- New extensibility features are documented and covered by examples.

## Recommended Delivery Order

1. Add tests and harness support.
2. Fix conditional rendering and loop execution.
3. Fix line spacing and wrap width.
4. Add spacer.
5. Add section padding.
6. Add row layout.
7. Add style inheritance.
8. Add custom function registration.
9. Add writer API enhancement.
10. Add embedded font registration.
11. Refresh documentation and examples at the end of each phase.

## Risks and Constraints

- PDF output is harder to snapshot-test than plain text output. Prefer behavior-focused tests over raw PDF binary comparison.
- Row layout and padding introduce coordinate state that can easily regress wrapping and page-flow behavior if implemented ad hoc.
- Style inheritance needs deterministic resolution and cycle handling before it is safe to document.
- Public API additions should be backward compatible; avoid renaming existing methods unless there is a major-version plan.

## Recommended First Milestone

The first implementation milestone should include only:

- test harness setup,
- section and element condition evaluation,
- section loop execution,
- line `spacingAfter`,
- wrap-width correction,
- documentation fixes for those behaviors.

This milestone delivers the highest value with the lowest schema churn and removes the most misleading gaps between the documented template model and actual runtime behavior.
