# DOCX Renderer — Known Issues

This file tracks known rendering defects in the DOCX output backend. Each issue includes a description of the correct behaviour per the AsciiDoc specification, what the renderer currently produces, why it matters in practice, and a minimal worked example showing the input and the expected vs actual OOXML output.

Issues are roughly ordered by impact on a real document.

---

## Issue 1 — Admonition blocks lose their style and label inside list continuations

### Spec reference

[Admonition Blocks](https://docs.asciidoctor.org/asciidoc/latest/blocks/admonitions/) — AsciiDoc defines five admonition types (`NOTE`, `TIP`, `IMPORTANT`, `CAUTION`, `WARNING`). A block admonition is written as a delimited `====` block with the admonition type as a block style attribute. The processor must render it as a visually distinct callout with the type label shown.

[List Continuation](https://docs.asciidoctor.org/asciidoc/latest/lists/continuation/) — A `+` list continuation attaches any block (including an admonition block) to the preceding list item. The attached block is logically part of the list item and should be rendered as such while still preserving its own block semantics.

### What should happen

An admonition block attached to a list item via `+` should be rendered with the `Admonition` paragraph style and a visible label ("NOTE:", "WARNING:", etc.), indented at the same level as the list item's continuation content.

### What actually happens

The admonition block is rendered as a plain `ListParagraph` paragraph with no label and no `Admonition` style. It is visually indistinguishable from the surrounding list continuation text. The `Admonition` style itself is also effectively invisible — the generated `styles.xml` definition carries no background colour, border, or indent:

```xml
<!-- Actual output — no label, wrong style -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
  </w:pPr>
  <w:r>
    <w:t xml:space="preserve">Cubit is a desktop application; ...</w:t>
  </w:r>
</w:p>
```

Two bugs compound:
1. The admonition-specific rendering path is bypassed when the block sits inside a list continuation context.
2. Even when the `Admonition` style is applied correctly outside a list, it has no visual properties (no background, no border, no label prefix), so the output still looks like plain body text.

### Worked example

**Input (`.adoc`)**

```asciidoc
. First list item
+
NOTE: This is important additional context attached to the first item.

. Second list item
```

**Expected OOXML** — the NOTE paragraph should use the `Admonition` style and open with a bold "NOTE:" run, indented to the list continuation level (`w:left="567"`):

```xml
<w:p>
  <w:pPr>
    <w:pStyle w:val="Admonition"/>
    <w:ind w:left="567"/>
  </w:pPr>
  <w:r>
    <w:rPr><w:b/></w:rPr>
    <w:t xml:space="preserve">NOTE: </w:t>
  </w:r>
  <w:r>
    <w:t xml:space="preserve">This is important additional context attached to the first item.</w:t>
  </w:r>
</w:p>
```

**Expected `styles.xml` definition** — the `Admonition` style should carry at minimum a left border and background shading so it is distinguishable at a glance:

```xml
<w:style w:type="paragraph" w:styleId="Admonition">
  <w:name w:val="Admonition"/>
  <w:pPr>
    <w:ind w:left="567" w:right="567"/>
    <w:pBdr>
      <w:left w:val="single" w:sz="8" w:space="4" w:color="4A90D9"/>
    </w:pBdr>
    <w:shd w:val="clear" w:color="auto" w:fill="EEF4FB"/>
  </w:pPr>
  <w:rPr>
    <w:sz w:val="21"/>
  </w:rPr>
</w:style>
```

---

## Issue 2 — List item title and body paragraph merged with `<w:br/>` instead of a paragraph break

### Spec reference

[Ordered List Items](https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/) — Each `.` prefix creates a discrete list item. The first paragraph of the item is its principal content. Any additional text within the item is a separate block.

[List Continuation](https://docs.asciidoctor.org/asciidoc/latest/lists/continuation/) — When a list item has attached continuation blocks, each continuation block is a logically separate block element joined to the item by the `+` marker.

### What should happen

A list item whose principal content is followed by one or more `+` continuation paragraphs should produce one `<w:p>` per logical paragraph — the numbered item paragraph, then one or more un-numbered continuation paragraphs. Each paragraph break carries its own spacing and can be styled independently.

### What actually happens

The bold title text and the first body paragraph are placed into the **same `<w:p>` element**, separated only by a `<w:br/>` (a soft line break, OOXML §17.3.3.1). A soft line break does not create a new paragraph — it is equivalent to Shift+Enter in Word. This means:

- No paragraph spacing is applied between the title and body.
- The two pieces of text cannot be independently styled (they share one `<w:pPr>`).
- The list numbering marker visually appears to cover only the first line of the combined run.
- Spell-check, grammar-check, and accessibility tooling treat the entire block as one sentence.

```xml
<!-- Actual output — title and body in one <w:p> separated by <w:br/> -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
    <w:numPr>
      <w:ilvl w:val="0"/>
      <w:numId w:val="2"/>
    </w:numPr>
  </w:pPr>
  <w:r>
    <w:rPr><w:b/></w:rPr>
    <w:t xml:space="preserve">XLSX Import and Export</w:t>
  </w:r>
  <w:r>
    <w:br/>
    <w:t xml:space="preserve">Add XLSX support as an additional import and export format...</w:t>
  </w:r>
</w:p>
```

### Worked example

**Input (`.adoc`)**

```asciidoc
. *XLSX Import and Export*
Add XLSX support as an additional import and export format alongside the existing CSV and Parquet workflows.
```

**Expected OOXML** — two separate `<w:p>` elements; the second carries the list indent but no numId so it does not show a number:

```xml
<!-- Paragraph 1: numbered, bold title -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
    <w:numPr>
      <w:ilvl w:val="0"/>
      <w:numId w:val="2"/>
    </w:numPr>
  </w:pPr>
  <w:r>
    <w:rPr><w:b/></w:rPr>
    <w:t xml:space="preserve">XLSX Import and Export</w:t>
  </w:r>
</w:p>

<!-- Paragraph 2: continuation — indented but not numbered -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
    <w:ind w:left="567"/>
  </w:pPr>
  <w:r>
    <w:t xml:space="preserve">Add XLSX support as an additional import and export format alongside the existing CSV and Parquet workflows.</w:t>
  </w:r>
</w:p>
```

---

## Issue 3 — List continuation paragraphs rendered at the left margin instead of indented

### Spec reference

[List Continuation](https://docs.asciidoctor.org/asciidoc/latest/lists/continuation/) — Continuation blocks are logically part of the list item. In rendered output they should be indented to the same level as the item's principal content (i.e., past the list marker) so they visually belong to the item rather than appearing to restart at the page margin.

### What should happen

Continuation paragraphs — those attached to a list item via `+` but without a list numbering marker — should carry the same left indent as the numbered paragraph text. For a top-level list at `w:left="567" w:hanging="360"`, the text starts at 567 − 360 = 207 twips from the number's position, but the continuation text should align with `w:left="567"` (no hanging) so the left edge matches the item text, not the margin.

### What actually happens

Continuation paragraphs receive only `<w:pStyle w:val="ListParagraph"/>` with no `<w:numPr>` and no explicit `<w:ind>`. The generated `ListParagraph` style definition in `styles.xml` carries no indentation either. As a result, Word renders them flush with the left page margin — visually detached from their owning list item.

```xml
<!-- Actual: no indent, no numPr — renders at left margin -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
  </w:pPr>
  <w:r>
    <w:rPr><w:b/></w:rPr>
    <w:t xml:space="preserve">Import:</w:t>
  </w:r>
  <w:r>
    <w:t xml:space="preserve"> The Supplier will implement a template registry...</w:t>
  </w:r>
</w:p>
```

### Worked example

**Input (`.adoc`)**

```asciidoc
. First item
+
This continuation paragraph should be indented.
+
So should this one.

. Second item
```

**Expected OOXML** — continuation paragraphs carry explicit `w:ind` matching the list level's `w:left`:

```xml
<!-- List item — numbered -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
    <w:numPr>
      <w:ilvl w:val="0"/>
      <w:numId w:val="1"/>
    </w:numPr>
  </w:pPr>
  <w:r><w:t xml:space="preserve">First item</w:t></w:r>
</w:p>

<!-- Continuation 1 — indented to match list text, no number -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
    <w:ind w:left="567"/>
  </w:pPr>
  <w:r><w:t xml:space="preserve">This continuation paragraph should be indented.</w:t></w:r>
</w:p>

<!-- Continuation 2 — same indent -->
<w:p>
  <w:pPr>
    <w:pStyle w:val="ListParagraph"/>
    <w:ind w:left="567"/>
  </w:pPr>
  <w:r><w:t xml:space="preserve">So should this one.</w:t></w:r>
</w:p>
```

The indent value (`567` at level 0, `927` at level 1, etc.) should be read from the corresponding abstractNum level's `w:left` value so continuation indents stay in sync with list indents automatically.

---

## Issue 4 — Possessive apostrophes split into separate text runs

### Spec reference

[Punctuation and Symbols](https://docs.asciidoctor.org/asciidoc/latest/text/punctuation-and-symbols/) — AsciiDoc applies typographic (curly) substitutions to certain character sequences. A straight apostrophe `'` appearing as part of a possessive or contraction (e.g., `it's`, `Customer's`) is converted to a right single quotation mark (U+2019) by the symbol substitution pass.

### What should happen

The typographic apostrophe substitution should convert `'` to U+2019 inline within the text content of the paragraph, preserving it as part of the same text run as the surrounding word. The word "Customer's" should appear as a single `<w:t>` element: `Customer\u2019s`.

### What actually happens

The parser emits the curly apostrophe as a discrete `Symbol` inline element (`types.Symbol{Name: "'"}`), causing the renderer to write it as its own `<w:r><w:t>'</w:t></w:r>`. Words containing possessives are fragmented into three runs:

```xml
<!-- Actual: "Customer's" split across three runs -->
<w:r>
  <w:rPr><w:b/></w:rPr>
  <w:t xml:space="preserve">Customer</w:t>
</w:r>
<w:r>
  <w:rPr><w:b/></w:rPr>
  <w:t xml:space="preserve">'</w:t>
</w:r>
<w:r>
  <w:rPr><w:b/></w:rPr>
  <w:t xml:space="preserve">s Manager</w:t>
</w:r>
```

This fragmentation causes:
- Double-clicking "Customer's" in Word selects only "Customer" (word boundary is at the apostrophe).
- Find & Replace for the literal string "Customer's" fails.
- Windows spell-check may flag "s Manager" as an unknown word.
- The XML is needlessly verbose — every possessive in the document contributes two extra `<w:r>` elements.

The root cause is that the renderer treats every `types.Symbol` node as a separate run, rather than checking whether the symbol can be merged with an adjacent `StringElement` run that shares identical `rPr` properties.

### Worked example

**Input (`.adoc`)**

```asciidoc
The Customer's nominated manager has authority to approve Change Orders.
```

**Expected OOXML** — the apostrophe is part of the same text run:

```xml
<w:p>
  <w:r>
    <w:t xml:space="preserve">The Customer\u2019s nominated manager has authority to approve Change Orders.</w:t>
  </w:r>
</w:p>
```

**Fix approach** — during OOXML rendering, before flushing each `<w:r>`, check whether the preceding run in the current paragraph builder has identical `rPr` properties. If so, append the symbol's text content to the existing run's `<w:t>` rather than opening a new `<w:r>`.

---

## Issue 5 — `cols` proportional width specification is ignored; all columns rendered equal width

### Spec reference

[Column Widths](https://docs.asciidoctor.org/asciidoc/latest/tables/add-columns/) — The `cols` attribute accepts a comma-separated list of positive integers representing **proportional widths**. A value of `1,3` means the first column is one unit wide and the second three units wide — the second column is three times wider than the first. The total number of units is the sum of all values; each column's share is `value / total`.

### What should happen

For a table declared with `[cols="1,3"]`, the two columns should each receive a `<w:gridCol>` whose `w:w` value is proportional to its weight:

```
total units = 1 + 3 = 4
col 1 width = (1/4) × pageTextWidth
col 2 width = (3/4) × pageTextWidth
```

For `[cols="1,3,1,3"]` (signature table), the four columns should be 1/8, 3/8, 1/8, 3/8 of the text width.

### What actually happens

The `cols` attribute values are parsed to determine the column **count** only; the proportional weights are discarded. Every column receives `tableGridWidthTwips / colCount` (9000 / colCount) twips, producing equal widths regardless of the spec:

```xml
<!-- Actual: [cols="1,3"] produces equal 4500/4500 instead of 2250/6750 -->
<w:tblGrid>
  <w:gridCol w:w="4500"/>
  <w:gridCol w:w="4500"/>
</w:tblGrid>

<!-- Actual: [cols="1,3,1,3"] produces equal 2250/2250/2250/2250 -->
<w:tblGrid>
  <w:gridCol w:w="2250"/>
  <w:gridCol w:w="2250"/>
  <w:gridCol w:w="2250"/>
  <w:gridCol w:w="2250"/>
</w:tblGrid>
```

In the SOW this affects every key-value table (Parts A, E, G, H) — the label column should be narrow (~25% for `[cols="1,3"]`) and the value column wide (~75%), but both are 50%. The signature table is most obviously wrong: the Signature/Name/Title/Date label cells should be narrow with wide write-in fields beside them.

Additionally, the constant `tableGridWidthTwips = 9000` does not match the actual text width derived from the page and margin settings in the theme (A4 with 20mm margins → 170mm ≈ 9638 twips). Column widths should be computed from the theme's resolved text width.

### Worked example

**Input (`.adoc`)**

```asciidoc
[cols="1,3"]
|===
| *Label*
| Value text that needs more space

| *Another label*
| More value text
|===
```

**Expected OOXML** — for an A4 page with 20mm left/right margins (text width ≈ 9638 twips), `[cols="1,3"]` → 1:3 ratio → 2410 / 7228 twips:

```xml
<w:tblGrid>
  <w:gridCol w:w="2410"/>
  <w:gridCol w:w="7228"/>
</w:tblGrid>
```

**Fix approach** — in `table.go`, retrieve the column weight list from `t.Columns()`, sum the weights, and for each column write `w:w = round(weight / totalWeight × textWidthTwips)` where `textWidthTwips` is derived from the theme's page width minus left and right margins.

---

## Issue 6 — `ListParagraph` style carries no base indentation

### Spec reference

This is an OOXML conformance issue rather than an AsciiDoc spec issue. OOXML §17.7.8 specifies that `List Paragraph` is a well-known paragraph style with a canonical definition that includes a non-zero left indent. Microsoft Word's built-in `List Paragraph` style includes `w:ind w:left="720"` as a style-level default, which Word users and third-party processors expect.

### What should happen

The generated `ListParagraph` style should define a base left indent matching the list level 0 indent used in the abstractNum definitions. This means that even a `ListParagraph` paragraph with no `numPr` and no overriding `w:ind` will be visually indented rather than flush with the margin.

### What actually happens

The generated `ListParagraph` style has no `w:ind`:

```xml
<w:style w:type="paragraph" w:styleId="ListParagraph">
  <w:name w:val="List Paragraph"/>
  <w:pPr>
    <w:outlineLvl w:val="0"/>
  </w:pPr>
  <w:rPr>
    <w:sz w:val="21"/>
    <w:szCs w:val="21"/>
  </w:rPr>
</w:style>
```

Without a style-level indent, any paragraph that relies on the style alone (rather than an explicit `w:ind` in the paragraph properties, or a `numPr` reference that carries the abstractNum indent) will render at the left margin. This is the root cause of Issue 3 above.

### Worked example

**Expected `ListParagraph` style definition in `styles.xml`** — mirrors Word's canonical definition:

```xml
<w:style w:type="paragraph" w:styleId="ListParagraph">
  <w:name w:val="List Paragraph"/>
  <w:pPr>
    <w:ind w:left="567"/>   <!-- matches abstractNum level 0 left indent -->
    <w:contextualSpacing/>
    <w:outlineLvl w:val="0"/>
  </w:pPr>
  <w:rPr>
    <w:sz w:val="21"/>
    <w:szCs w:val="21"/>
  </w:rPr>
</w:style>
```

The `w:contextualSpacing` element suppresses the paragraph's `w:after` spacing when the following paragraph uses the same style, which is standard Word behaviour for list paragraphs and prevents unwanted gaps between consecutive list items.

---

## Issue 7 — Table grid width constant does not match the theme-derived text width

### Description

The renderer uses a hardcoded constant `tableGridWidthTwips = 9000` when computing `<w:gridCol w:w="..."/>` values. This constant does not correspond to the actual printable text width for the document, which depends on the page size and margin settings from the theme.

For an A4 page (210mm wide) with 20mm left and 20mm right margins (as in the default `sycel-theme.yml`):

```
text width = 210mm − 20mm − 20mm = 170mm
           = 170 × (1440/25.4) twips
           ≈ 9638 twips
```

The discrepancy (9638 vs 9000, ≈ 6.6%) means the column grid declared in `<w:tblGrid>` does not match the `<w:tblW w:w="5000" w:type="pct"/>` (100% of text width) table width. Word compensates silently, but the column proportions are computed against the wrong base, and any explicit `<w:tcW>` values will be subtly incorrect.

### Fix approach

Compute `textWidthTwips` during theme loading from the page size and margin values:

```go
textWidthTwips = pageWidthTwips - marginLeftTwips - marginRightTwips
```

Use `textWidthTwips` wherever `tableGridWidthTwips` is currently used.

---

## Issue 8 — No `keepWithNext` / `keepLines` on list items

### Description

Neither the `ListParagraph` style definition nor individual list item paragraphs set `<w:keepWithNext/>` or `<w:keepLines/>`. As a result, a numbered list item can be orphaned at the bottom of a page — the number and bold title appear as the last line on one page while all the associated content appears on the next.

For long-form documents such as SOWs or contracts, this is a common and noticeable problem when items have substantial body text.

### Fix approach

Add `<w:keepNext/>` to list item paragraphs that have continuation content following them (i.e., whenever `renderListItem` emits more than one paragraph for the item, set `keepNext` on all but the last). This is the same approach used for headings in the generated styles.xml (where `<w:keepNext/>` is already present on heading styles).

---

## Issue 9 — No header row repetition (`tblHeader`) for tables that span pages

### Description

When a table is long enough to span more than one page, Word can repeat the first (header) row at the top of each continuation page if the row carries `<w:trPr><w:tblHeader/></w:trPr>`. The renderer never emits this element.

In the SOW this affects the acceptance criteria tables in Part D, which can run several pages and would benefit from repeating the column headers.

### Fix approach

The AsciiDoc table model uses a header row concept (rows above the first `|===` body separator). When a table has a header row, emit `<w:tblHeader/>` in its `<w:trPr>`:

```xml
<w:tr>
  <w:trPr>
    <w:tblHeader/>
  </w:trPr>
  ...header cells...
</w:tr>
```

---

## Issue 10 — Duplicate heading bookmark names not disambiguated

### Description

Heading bookmarks are generated by replacing spaces with underscores in the heading text. If the same heading text appears more than once in a document (e.g., multiple sections each containing a "Description" subsection), the generated `w:name` attribute will be identical for both bookmarks. OOXML requires bookmark names to be unique within a document; duplicate names produce a malformed file that Word corrects silently (by discarding the duplicate) but which causes cross-reference and navigation failures in compliant readers.

### Fix approach

Track heading bookmark names already emitted during the render pass. If a new heading produces a name that has already been used, append `_2`, `_3`, etc. to make it unique:

```go
name := toBookmarkName(heading.Title)
if count, seen := r.bookmarkNames[name]; seen {
    r.bookmarkNames[name]++
    name = fmt.Sprintf("%s_%d", name, count+1)
} else {
    r.bookmarkNames[name] = 1
}
```
