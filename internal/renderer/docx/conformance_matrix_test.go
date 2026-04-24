package docx_test

// conformance_matrix_test.go — AsciiDoc × OOXML conformance matrix.
//
// Each entry in the matrix represents one AsciiDoc source construct mapped to
// the OOXML elements it should produce. The matrix makes coverage gaps visible
// at a glance: an entry with Pending: true (rendered as XIt) documents a known
// gap in the renderer without requiring a failing test in CI.
//
// How to read this file:
//   - AsciiDocSpec: URL to the AsciiDoc spec section defining the construct.
//   - ECMA376Sec:   ECMA-376 Part 1 section that specifies the OOXML output.
//   - Pending:      true when the renderer does not yet implement this construct.
//
// When implementing a new renderer feature:
//  1. Write the production code.
//  2. Change Pending: true → Pending: false in the relevant entry (or entries).
//  3. Run the matrix and verify the test passes.

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type conformanceEntry struct {
	// AsciiDocSpec is the URL to the relevant AsciiDoc specification page.
	AsciiDocSpec string
	// ECMA376Sec is the ECMA-376 Part 1 section number (e.g. "§17.9.3").
	ECMA376Sec string
	// Description is a human-readable summary of what is being tested.
	Description string
	// Source is the AsciiDoc input to render.
	Source string
	// Check is the assertion closure run against the rendered document.
	Check func(doc renderedDocx)
	// Pending marks entries whose renderer feature is not yet implemented.
	// Set to true to emit an XIt (pending) instead of an It.
	Pending bool
}

// conformanceMatrix is the authoritative list of AsciiDoc → OOXML mappings.
// Add a new row whenever a new AsciiDoc construct is implemented in the
// DOCX renderer or when a gap is discovered during review.
var conformanceMatrix = []conformanceEntry{

	// ── Inline formatting (ECMA-376 §17.3.2 Run Properties) ──────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/text/bold-and-italic/",
		ECMA376Sec:   "§17.3.2.1",
		Description:  "constrained bold *text* emits w:b on the run",
		Source:       "plain *bold* plain",
		Check: func(doc renderedDocx) {
			r := doc.findRun("bold")
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeTrue())
			Expect(doc.findRun("plain")).ToNot(BeNil())
			Expect(doc.findRun("plain").Bold).To(BeFalse())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/text/bold-and-italic/",
		ECMA376Sec:   "§17.3.2.14",
		Description:  "constrained italic _text_ emits w:i on the run",
		Source:       "plain _italic_ plain",
		Check: func(doc renderedDocx) {
			r := doc.findRun("italic")
			Expect(r).ToNot(BeNil())
			Expect(r.Italic).To(BeTrue())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/text/monospace/",
		ECMA376Sec:   "§17.3.2.26",
		Description:  "monospace `code` emits w:rFonts with monospace font",
		Source:       "inline `code` here",
		Check: func(doc renderedDocx) {
			r := doc.findRun("code")
			Expect(r).ToNot(BeNil())
			Expect(r.Monospace).To(BeTrue())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/text/highlight/",
		ECMA376Sec:   "§17.3.2.15",
		Description:  "#marked# text emits w:highlight",
		Source:       "this is #marked# text",
		Check: func(doc renderedDocx) {
			r := doc.findRun("marked")
			Expect(r).ToNot(BeNil())
			Expect(r.Highlight).To(BeTrue())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/text/subscript-and-superscript/",
		ECMA376Sec:   "§17.3.2.42",
		Description:  "subscript ~text~ emits w:vertAlign val=subscript",
		Source:       "H~2~O",
		Check: func(doc renderedDocx) {
			r := doc.findRun("2")
			Expect(r).ToNot(BeNil())
			Expect(r.Subscript).To(BeTrue())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/text/subscript-and-superscript/",
		ECMA376Sec:   "§17.3.2.42",
		Description:  "superscript ^text^ emits w:vertAlign val=superscript",
		Source:       "x^2^",
		Check: func(doc renderedDocx) {
			r := doc.findRun("2")
			Expect(r).ToNot(BeNil())
			Expect(r.Superscript).To(BeTrue())
		},
	},

	// ── Sections / headings (ECMA-376 §17.3.1.6 pStyle) ─────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/sections/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "document title = renders with Title style",
		Source:       "= My Document\n\nBody.",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("My Document")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Title"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/sections/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "level-1 section == renders with Heading2 style",
		Source:       "= Doc\n\n== Section\n\nBody.",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("Section")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Heading2"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/sections/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "level-2 section === renders with Heading3 style",
		Source:       "= Doc\n\n== Parent\n\n=== Child\n\nBody.",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("Child")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Heading3"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/sections/",
		ECMA376Sec:   "§17.13.6.2",
		Description:  "section heading produces a w:bookmarkStart for cross-references",
		Source:       "= Doc\n\n[#my-anchor]\n== Target\n\nBody.",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("Target")
			Expect(p).ToNot(BeNil())
			Expect(p.Bookmarks).ToNot(BeEmpty())
		},
	},

	// ── Ordered lists (ECMA-376 §17.9) ───────────────────────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/",
		ECMA376Sec:   "§17.9.3",
		Description:  "ordered list . produces decimal numbering",
		Source:       ". first\n. second",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("first")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("decimal"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/#start-value",
		ECMA376Sec:   "§17.9.7",
		Description:  "[start=5] ordered list emits w:startOverride val=5",
		Source:       "[start=5]\n. Fifth\n. Sixth",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("Fifth")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].StartOverride).To(Equal("5"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/#styles",
		ECMA376Sec:   "§17.9.3",
		Description:  "[lowerroman] ordered list produces lowerRoman numFmt",
		Source:       "[lowerroman]\n. one\n. two",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("one")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("lowerRoman"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/#styles",
		ECMA376Sec:   "§17.9.3",
		Description:  "[upperroman] ordered list produces upperRoman numFmt",
		Source:       "[upperroman]\n. one\n. two",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("one")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("upperRoman"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/#styles",
		ECMA376Sec:   "§17.9.3",
		Description:  "[loweralpha] ordered list produces lowerLetter numFmt",
		Source:       "[loweralpha]\n. alpha\n. beta",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("alpha")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("lowerLetter"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/#styles",
		ECMA376Sec:   "§17.9.3",
		Description:  "[upperalpha] ordered list produces upperLetter numFmt",
		Source:       "[upperalpha]\n. Alpha\n. Beta",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("Alpha")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("upperLetter"))
		},
	},

	// ── Unordered lists (ECMA-376 §17.9) ─────────────────────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/lists/unordered/",
		ECMA376Sec:   "§17.9.3",
		Description:  "unordered list * produces bullet numbering",
		Source:       "* alpha\n* beta",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("alpha")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("bullet"))
		},
	},

	// ── Tables (ECMA-376 §17.4) ───────────────────────────────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/tables/",
		ECMA376Sec:   "§17.4.1",
		Description:  "basic table produces w:tbl with correct row and cell counts",
		Source:       "|===\n| A | B\n\n| 1 | 2\n|===",
		Check: func(doc renderedDocx) {
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			Expect(tables[0].Rows).To(HaveLen(2))
			for _, row := range tables[0].Rows {
				Expect(row.Cells).To(HaveLen(2))
			}
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/tables/",
		ECMA376Sec:   "§17.4.27",
		Description:  "table header row carries w:tblHeader; data rows do not",
		Source:       "|===\n| H1 | H2\n\n| D1 | D2\n|===",
		Check: func(doc renderedDocx) {
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			Expect(tables[0].Rows[0].IsHeader).To(BeTrue())
			Expect(tables[0].Rows[1].IsHeader).To(BeFalse())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/tables/add-columns/",
		ECMA376Sec:   "§17.4.13",
		Description:  "[cols=1,3] produces proportional gridCol widths (1:3 ratio)",
		Source:       "[cols=\"1,3\"]\n|===\n| Narrow | Wide\n|===",
		Check: func(doc renderedDocx) {
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			widths := tables[0].GridColWidths
			Expect(widths).To(HaveLen(2))
			total := widths[0] + widths[1]
			Expect(total).To(BeNumerically(">", 0))
			// Ratio 1:3 → col0 ≈ total/4, col1 ≈ 3*total/4; allow ±2 twip rounding
			unit := total / 4
			Expect(widths[0]).To(BeNumerically("~", unit, 2))
			Expect(widths[1]).To(BeNumerically("~", 3*unit, 2))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/",
		ECMA376Sec:   "§17.4.3.15",
		Description:  "column span 2+| emits w:gridSpan val=2 (PENDING: not yet implemented)",
		Source:       "|===\n2+| Wide\n| A | B\n|===",
		Pending:      true,
		Check: func(doc renderedDocx) {
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			Expect(tables[0].Rows[0].Cells[0].GridSpan).To(Equal("2"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/",
		ECMA376Sec:   "§17.7.6",
		Description:  "row span .2+| emits w:vMerge restart/continuation (PENDING: not yet implemented)",
		Source:       "|===\n.2+| Tall | B\n| C\n|===",
		Pending:      true,
		Check: func(doc renderedDocx) {
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			Expect(tables[0].Rows[0].Cells[0].VMerge).To(Equal("restart"))
			Expect(tables[0].Rows[1].Cells[0].VMerge).To(Equal("continuation"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/tables/format-cell-content/",
		ECMA376Sec:   "§17.3.1.13",
		Description:  "^| center-aligned cell emits w:jc val=center on cell paragraph (PENDING: not yet implemented)",
		Source:       "|===\n^| Centered\n|===",
		Pending:      true,
		Check: func(doc renderedDocx) {
			// ^| sets horizontal centering — rendered as w:jc val="center" on the
			// paragraph inside the cell (w:pPr/w:jc), not as w:vAlign on the cell.
			// w:vAlign (§17.4.22) is vertical alignment and is unrelated.
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			cell := tables[0].Rows[0].Cells[0]
			Expect(cell.Paragraphs).ToNot(BeEmpty(),
				"table cell must contain at least one paragraph")
			Expect(cell.Paragraphs[0].Alignment).To(Equal("center"),
				"^| must emit w:jc val=center on the paragraph inside the cell")
		},
	},

	// ── Delimited blocks (ECMA-376 §17.3.1.6 paragraph styles) ──────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/blocks/source-blocks/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "[source] block produces CodeBlock paragraph style",
		Source:       "[source,go]\n----\nfmt.Println(\"hello\")\n----",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("hello")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("CodeBlock"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/blocks/admonitions/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "NOTE admonition produces Admonition paragraph style",
		Source:       "NOTE: Pay attention.",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("NOTE:")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Admonition"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/blocks/admonitions/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "WARNING admonition block produces Admonition style with bold label",
		Source:       "[WARNING]\n====\nDanger ahead.\n====",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("WARNING:")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Admonition"))
			Expect(p.Runs).ToNot(BeEmpty())
			Expect(p.Runs[0].Bold).To(BeTrue())
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/blocks/quote-blocks/",
		ECMA376Sec:   "§17.3.1.6",
		Description:  "quote block produces Quote paragraph style",
		Source:       "[quote]\n____\nTo be or not to be.\n____",
		Check: func(doc renderedDocx) {
			p := doc.findParagraph("To be or not to be")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Quote"))
		},
	},

	// ── Hyperlinks and cross-references (ECMA-376 §17.18) ────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/macros/links/",
		ECMA376Sec:   "§17.18.1",
		Description:  "external link produces w:hyperlink with External relationship",
		Source:       "Visit https://example.com[Example].",
		Check: func(doc renderedDocx) {
			rel := doc.findRelationship("https://example.com")
			Expect(rel).ToNot(BeNil())
			Expect(rel.TargetMode).To(Equal("External"))
		},
	},
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/macros/xref/",
		ECMA376Sec:   "§17.18.1",
		Description:  "<<anchor>> cross-reference produces w:hyperlink with w:anchor matching bookmark",
		Source:       "= Doc\n\n[#target]\n== Target Section\n\nSee <<target>>.",
		Check: func(doc renderedDocx) {
			section := doc.findParagraph("Target Section")
			Expect(section).ToNot(BeNil())
			Expect(section.Bookmarks).ToNot(BeEmpty())

			ref := doc.findParagraph("See")
			Expect(ref).ToNot(BeNil())
			Expect(ref.Links).ToNot(BeEmpty())
			Expect(ref.Links[0].Anchor).To(Equal(section.Bookmarks[0]))
		},
	},

	// ── Footnotes (ECMA-376 §17.11) ──────────────────────────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/macros/footnote/",
		ECMA376Sec:   "§17.11.11",
		Description:  "footnote:[] produces w:footnoteReference in document and body in footnotes.xml",
		Source:       "A statement.footnote:[Supporting evidence.]",
		Check: func(doc renderedDocx) {
			Expect(doc.files).To(HaveKey("word/footnotes.xml"))
			Expect(doc.footnotesXML()).To(ContainSubstring("Supporting evidence."))
		},
	},

	// ── Document structure (ECMA-376 §17.6) ──────────────────────────────────
	{
		AsciiDocSpec: "https://docs.asciidoctor.org/asciidoc/latest/document/",
		ECMA376Sec:   "§17.6.1",
		Description:  "every document has w:pgSz with positive page dimensions",
		Source:       "A paragraph.",
		Check: func(doc renderedDocx) {
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.PageW).ToNot(BeEmpty())
			Expect(sp.PageH).ToNot(BeEmpty())
		},
	},
}

// The conformance matrix is driven by a Describe loop so that:
//   - Pending entries emit XIt (shown as pending in CI, no failure).
//   - Active entries emit It (must pass).
//
// Each test is labelled with its ECMA-376 section and AsciiDoc spec URL so
// the output is self-documenting.

var _ = Describe("AsciiDoc × OOXML conformance matrix", func() {
	for _, e := range conformanceMatrix {
		entry := e // capture loop variable for goroutine safety
		label := entry.Description + " [" + entry.ECMA376Sec + "]"
		run := func() {
			doc := renderDocx(entry.Source)
			entry.Check(doc)
		}
		if entry.Pending {
			XIt(label, run)
		} else {
			It(label, run)
		}
	}
})
