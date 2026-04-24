package docx_test

// ecma376_conformance_test.go — structural conformance tests grounded in the
// ECMA-376 Part 1 (WordprocessingML) specification and the AsciiDoc spec.
//
// Each Describe/Context block cites the relevant ECMA-376 section and, where
// applicable, the AsciiDoc spec URL so the reason for every test is clear.
//
// Why these tests are not fragile:
//   - They use the structural parsers (parsedParagraph, parsedTable, etc.) rather
//     than ContainSubstring on raw XML, so they survive whitespace and ordering
//     changes in the XML output.
//   - xml.Unmarshal well-formedness checks have no opinion about content; they
//     only fail when the XML is syntactically invalid.
//   - Each It() targets exactly one OOXML property so a failure pinpoints one bug.

import (
	"encoding/xml"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─── §17.2 Document Structure — XML Well-Formedness ───────────────────────────
//
// ECMA-376 requires every XML part to be well-formed per the XML 1.0 spec.
// The renderer builds all XML by hand (string concatenation), so a typo in any
// element name, missing namespace prefix, or unescaped attribute value would
// produce a broken file. xml.Unmarshal catches these errors immediately.

var _ = Describe("ECMA-376 §17.2 — document structure XML well-formedness", func() {

	// Use a feature-rich document so that all major XML parts are exercised.
	const richSource = `= Document Title
Author Name

== First Section

A paragraph with *bold* and _italic_ text.

NOTE: An admonition block.

|===
| Header A | Header B

| Cell 1 | Cell 2
|===

. Ordered one
. Ordered two

A footnote.footnote:[The footnote body text.]`

	It("word/document.xml is well-formed XML", func() {
		doc := renderDocx(richSource)
		Expect(xml.Unmarshal([]byte(doc.documentXML()), new(interface{}))).To(Succeed(),
			"word/document.xml must be well-formed XML — check ooxml.go for unescaped strings or missing closing tags")
	})

	It("word/styles.xml is well-formed XML", func() {
		doc := renderDocx(richSource)
		content := string(doc.files["word/styles.xml"])
		Expect(content).ToNot(BeEmpty())
		Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed(),
			"word/styles.xml must be well-formed XML")
	})

	It("word/numbering.xml is well-formed XML", func() {
		doc := renderDocx(richSource)
		content := string(doc.files["word/numbering.xml"])
		Expect(content).ToNot(BeEmpty())
		Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed(),
			"word/numbering.xml must be well-formed XML")
	})

	It("word/footnotes.xml is well-formed XML when footnotes are present", func() {
		doc := renderDocx(richSource)
		Expect(doc.files).To(HaveKey("word/footnotes.xml"),
			"a document with footnote:[] must produce word/footnotes.xml")
		content := string(doc.files["word/footnotes.xml"])
		Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed(),
			"word/footnotes.xml must be well-formed XML")
	})

	It("word/header1.xml is well-formed XML when present", func() {
		doc := renderDocx(richSource)
		if content, ok := doc.files["word/header1.xml"]; ok {
			Expect(xml.Unmarshal(content, new(interface{}))).To(Succeed(),
				"word/header1.xml must be well-formed XML")
		}
	})

	It("word/footer1.xml is well-formed XML when present", func() {
		doc := renderDocx(richSource)
		if content, ok := doc.files["word/footer1.xml"]; ok {
			Expect(xml.Unmarshal(content, new(interface{}))).To(Succeed(),
				"word/footer1.xml must be well-formed XML")
		}
	})
})

// ─── §17.6 Section Properties — Page Size and Margins ────────────────────────
//
// ECMA-376 §17.6.1 w:pgSz defines page dimensions in twips (twentieths of a point).
// ECMA-376 §17.6.11 w:pgMar defines page margins in twips.
// All four margin values must be present for the document to lay out correctly
// in Word and LibreOffice.

var _ = Describe("ECMA-376 §17.6 — section properties", func() {

	It("should set page width and height as positive integers in twips", func() {
		doc := renderDocx("= Doc\n\nA paragraph.")
		sp := doc.parseSectionProps()
		Expect(sp).ToNot(BeNil(), "w:sectPr must be present in every document")

		w, err := strconv.Atoi(sp.PageW)
		Expect(err).ToNot(HaveOccurred(), "w:pgSz w:w must be a parseable integer")
		Expect(w).To(BeNumerically(">", 0), "page width must be positive")

		h, err := strconv.Atoi(sp.PageH)
		Expect(err).ToNot(HaveOccurred(), "w:pgSz w:h must be a parseable integer")
		Expect(h).To(BeNumerically(">", 0), "page height must be positive")
	})

	It("should set all four page margins as non-empty values", func() {
		doc := renderDocx("= Doc\n\nA paragraph.")
		sp := doc.parseSectionProps()
		Expect(sp).ToNot(BeNil())

		Expect(sp.Top).ToNot(BeEmpty(), "w:pgMar w:top must be set")
		Expect(sp.Bottom).ToNot(BeEmpty(), "w:pgMar w:bottom must be set")
		Expect(sp.Left).ToNot(BeEmpty(), "w:pgMar w:left must be set")
		Expect(sp.Right).ToNot(BeEmpty(), "w:pgMar w:right must be set")
	})

	It("should set all four page margins as positive integers in twips", func() {
		doc := renderDocx("= Doc\n\nA paragraph.")
		sp := doc.parseSectionProps()
		Expect(sp).ToNot(BeNil())

		for name, val := range map[string]string{
			"top": sp.Top, "bottom": sp.Bottom,
			"left": sp.Left, "right": sp.Right,
		} {
			v, err := strconv.Atoi(val)
			Expect(err).ToNot(HaveOccurred(), "w:pgMar w:%s must be a parseable integer", name)
			Expect(v).To(BeNumerically(">", 0), "w:pgMar w:%s must be positive", name)
		}
	})
})

// ─── §17.7 Styles — Document Defaults and Heading Chain ──────────────────────
//
// ECMA-376 §17.7.5 docDefaults provides the root of the style inheritance chain.
// Heading styles must carry w:outlineLvl and w:keepNext per §17.3.1 to support
// navigation panes and paragraph-widow prevention.
//
// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/sections/

var _ = Describe("ECMA-376 §17.7 — styles", func() {

	It("should set non-empty base font and size in docDefaults", func() {
		doc := renderDocx("A paragraph.")
		dd := doc.parseDocDefaults()
		Expect(dd.Font).ToNot(BeEmpty(), "docDefaults must declare a base font (w:rFonts)")
		Expect(dd.Size).ToNot(BeEmpty(), "docDefaults must declare a base font size (w:sz)")
	})

	It("Heading1 style should have outlineLvl 0 and keepNext", func() {
		// ECMA-376 §17.3.1.26 outlineLvl, §17.3.1.15 keepNext
		doc := renderDocx("= Doc\n\n== Heading")
		h1 := doc.findStyle("Heading1")
		Expect(h1).ToNot(BeNil(), "Heading1 style must be defined")
		Expect(h1.OutlineLevel).To(Equal("0"),
			"Heading1 must have w:outlineLvl val=0 for document navigation")
		Expect(h1.KeepNext).To(BeTrue(),
			"Heading1 must have w:keepNext to prevent orphaned headings")
	})

	It("Heading2 style should have outlineLvl 1 and keepNext", func() {
		doc := renderDocx("= Doc\n\n== H2\n\n=== H3")
		h2 := doc.findStyle("Heading2")
		Expect(h2).ToNot(BeNil())
		Expect(h2.OutlineLevel).To(Equal("1"))
		Expect(h2.KeepNext).To(BeTrue())
	})

	It("hyperlink runs should carry the Hyperlink character style and underline", func() {
		// ECMA-376 §17.7.4 character styles, §17.3.2 run properties
		doc := renderDocx("Visit https://example.com[the site].")
		p := doc.findParagraph("the site")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		for _, r := range p.Links[0].Runs {
			if r.Text != "" {
				Expect(r.CharStyle).To(Equal("Hyperlink"),
					"hyperlink runs must reference the Hyperlink character style")
				Expect(r.Underline).To(BeTrue(),
					"hyperlink runs must carry w:u (underline)")
			}
		}
	})
})

// ─── §17.9 Numbering — Multiple Lists and Start Values ───────────────────────
//
// ECMA-376 §17.9.7 w:num — each list instance must have its own w:num element
// so that two separate ordered lists in a document count independently.
//
// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/

var _ = Describe("ECMA-376 §17.9 — numbering", func() {

	It("two ordered lists in one document should get distinct numId values", func() {
		// ECMA-376 §17.9.7: each w:num is an independent list instance.
		// If both lists share the same numId the second list continues the
		// counter from where the first left off — wrong behaviour.
		doc := renderDocx(`. Alpha
. Beta

Paragraph in between.

. Gamma
. Delta`)

		paras := doc.parseParagraphs()
		// Find the first paragraph of each list
		var firstListID, secondListID string
		for _, p := range paras {
			if p.NumID == "" {
				continue
			}
			if firstListID == "" {
				firstListID = p.NumID
				continue
			}
			if p.NumID != firstListID && secondListID == "" {
				secondListID = p.NumID
				break
			}
		}
		Expect(firstListID).ToNot(BeEmpty(), "first list must have a numId")
		Expect(secondListID).ToNot(BeEmpty(), "second list must have a distinct numId")
		Expect(firstListID).ToNot(Equal(secondListID),
			"two separate ordered lists must use different w:num instances")
	})

	It("[start=3] ordered list should emit w:startOverride val=3 in numbering.xml", func() {
		// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/lists/ordered/#start-value
		// ECMA-376 §17.9.7 w:startOverride overrides the abstract numbering start.
		doc := renderDocx(`[start=3]
. Third item
. Fourth item`)

		p := doc.findParagraph("Third item")
		Expect(p).ToNot(BeNil())
		Expect(p.NumID).ToNot(BeEmpty())

		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].StartOverride).To(Equal("3"),
			"w:num/w:lvlOverride/w:startOverride must reflect the [start=3] AsciiDoc attribute")
	})

	It("three-level nested unordered list should use ilvl 0 for each level", func() {
		// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/lists/unordered/
		// ECMA-376 §17.9.3 w:ilvl: libasciidoc uses separate w:num per nesting depth,
		// each at ilvl 0, rather than a single multi-level list.
		doc := renderDocx(`* Level one
** Level two
*** Level three`)

		p1 := doc.findParagraph("Level one")
		Expect(p1).ToNot(BeNil())
		Expect(p1.NumLevel).To(Equal("0"))

		p2 := doc.findParagraph("Level two")
		Expect(p2).ToNot(BeNil())
		Expect(p2.NumLevel).To(Equal("0"))
		Expect(p2.NumID).ToNot(Equal(p1.NumID),
			"each nesting depth uses its own w:num instance")

		p3 := doc.findParagraph("Level three")
		Expect(p3).ToNot(BeNil())
		Expect(p3.NumLevel).To(Equal("0"))
		Expect(p3.NumID).ToNot(Equal(p2.NumID))
	})
})

// ─── §17.13.6 Bookmarks — Deduplication ──────────────────────────────────────
//
// ECMA-376 §17.13.6.2 w:bookmarkStart requires that w:name values are unique
// within a document. When two sections share the same title, the renderer must
// produce unique bookmark names (e.g. _section_name and _section_name_2).
//
// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/sections/

var _ = Describe("ECMA-376 §17.13.6 — bookmark deduplication", func() {

	It("two sections with identical titles should produce distinct bookmark names", func() {
		doc := renderDocx(`= Document

== Repeated Title

First occurrence.

== Repeated Title

Second occurrence.`)

		paras := doc.parseParagraphs()
		var bookmarkNames []string
		for _, p := range paras {
			bookmarkNames = append(bookmarkNames, p.Bookmarks...)
		}

		Expect(bookmarkNames).To(HaveLen(2),
			"each section heading must produce exactly one bookmark")
		Expect(bookmarkNames[0]).ToNot(Equal(bookmarkNames[1]),
			"duplicate section titles must receive unique w:bookmarkStart w:name values "+
				"(e.g. _repeated_title and _repeated_title_2)")
	})

	It("unique section titles should not be given suffix-qualified names", func() {
		doc := renderDocx(`= Document

== First Section

== Second Section`)

		paras := doc.parseParagraphs()
		var bookmarkNames []string
		for _, p := range paras {
			bookmarkNames = append(bookmarkNames, p.Bookmarks...)
		}

		Expect(bookmarkNames).To(HaveLen(2))
		// Neither name should carry a numeric suffix when they are genuinely distinct
		for _, name := range bookmarkNames {
			Expect(name).ToNot(MatchRegexp(`_\d+$`),
				"unique section titles must not receive a numeric suffix")
		}
	})
})

// ─── §17.4 Tables — Column Spans (pending implementation) ────────────────────
//
// ECMA-376 §17.4.3.15 w:gridSpan specifies the number of grid columns a cell
// spans. AsciiDoc uses the 2+| prefix syntax to express this.
//
// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/
//
// These tests are marked XIt (pending) because the renderer does not yet emit
// w:gridSpan. They serve as the regression target for when the feature lands.

var _ = Describe("ECMA-376 §17.4.3 — table column spans (pending)", func() {

	It("2+| syntax should emit w:gridSpan val=2 on the spanning cell", func() {
		// AsciiDoc: https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/
		// ECMA-376 §17.4.3.15 w:gridSpan
		doc := renderDocx(`|===
2+| Wide cell
| A | B
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		Expect(tables[0].Rows[0].Cells[0].GridSpan).To(Equal("2"),
			"a cell with 2+| prefix must emit w:gridSpan w:val=\"2\"")
	})

	It(".2+| syntax should emit w:vMerge restart on the first merged cell", func() {
		// AsciiDoc: https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/
		// ECMA-376 §17.7.6 w:vMerge
		doc := renderDocx(`|===
.2+| Tall | B
| C
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		Expect(tables[0].Rows[0].Cells[0].VMerge).To(Equal("restart"),
			"the first cell of a row span must emit w:vMerge w:val=\"restart\"")
		Expect(tables[0].Rows[1].Cells[0].VMerge).To(Equal("continuation"),
			"subsequent cells of a row span must emit w:vMerge with no val attribute")
	})
})
