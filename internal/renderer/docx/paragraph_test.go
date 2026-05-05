package docx_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("paragraphs", func() {

	It("should render a simple paragraph", func() {
		doc := renderDocx("hello world")

		p := doc.findParagraph("hello world")
		Expect(p).ToNot(BeNil())
		Expect(p.Style).To(BeEmpty()) // regular paragraph, no special style
	})

	It("should render multiple paragraphs as separate w:p elements", func() {
		doc := renderDocx(`first paragraph

second paragraph`)

		paras := doc.parseParagraphs()
		texts := []string{}
		for _, p := range paras {
			t := p.text()
			if t != "" {
				texts = append(texts, t)
			}
		}
		Expect(texts).To(ContainElement(ContainSubstring("first paragraph")))
		Expect(texts).To(ContainElement(ContainSubstring("second paragraph")))
	})

	It("should render bold run with w:b on the run containing 'bold'", func() {
		doc := renderDocx("this is *bold* text")

		r := doc.findRun("bold")
		Expect(r).ToNot(BeNil())
		Expect(r.Bold).To(BeTrue())

		// "this is " run should NOT be bold
		plain := doc.findRun("this is ")
		Expect(plain).ToNot(BeNil())
		Expect(plain.Bold).To(BeFalse())
	})

	It("should render italic run with w:i on the run containing 'italic'", func() {
		doc := renderDocx("this is _italic_ text")

		r := doc.findRun("italic")
		Expect(r).ToNot(BeNil())
		Expect(r.Italic).To(BeTrue())
	})

	It("should render monospace run with Courier New font on the run containing 'code'", func() {
		doc := renderDocx("this is `code` text")

		r := doc.findRun("code")
		Expect(r).ToNot(BeNil())
		Expect(r.Monospace).To(BeTrue())
	})

	It("should render marked text with highlight on the run containing 'marked'", func() {
		doc := renderDocx("this is #marked# text")

		r := doc.findRun("marked")
		Expect(r).ToNot(BeNil())
		Expect(r.Highlight).To(BeTrue())
	})

	It("should render subscript on the run containing '2'", func() {
		doc := renderDocx("H~2~O")

		r := doc.findRun("2")
		Expect(r).ToNot(BeNil())
		Expect(r.Subscript).To(BeTrue())
	})

	It("should render superscript on the run containing '2'", func() {
		doc := renderDocx("x^2^")

		r := doc.findRun("2")
		Expect(r).ToNot(BeNil())
		Expect(r.Superscript).To(BeTrue())
	})

	It("should render admonition paragraph with Admonition style and bold label", func() {
		doc := renderDocx(`NOTE: Pay attention.`)

		p := doc.findParagraph("NOTE:")
		Expect(p).ToNot(BeNil())
		Expect(p.Style).To(Equal("Admonition"))

		// The label run should be bold
		Expect(p.Runs).ToNot(BeEmpty())
		Expect(p.Runs[0].Text).To(ContainSubstring("NOTE:"))
		Expect(p.Runs[0].Bold).To(BeTrue())

		// The body text should be in the same paragraph
		Expect(p.text()).To(ContainSubstring("Pay attention."))
	})

	It("should render hard line breaks as w:br within the paragraph", func() {
		doc := renderDocx("first +\nsecond")

		p := doc.findParagraph("first")
		Expect(p).ToNot(BeNil())
		// The paragraph should contain both "first" and "second"
		Expect(p.text()).To(ContainSubstring("first"))
		Expect(p.text()).To(ContainSubstring("second"))
		// There should be a line break character in some run
		hasBreak := false
		for _, r := range p.Runs {
			if strings.Contains(r.Text, "\n") {
				hasBreak = true
			}
		}
		Expect(hasBreak).To(BeTrue())
	})

	It("should merge adjacent plain-text runs into a single w:r", func() {
		// paragraphBuilder (Change 1) merges consecutive inline elements that share
		// the same runStyle. A paragraph like "Hello world" must produce one run,
		// not two, since both words have identical (default) styling.
		doc := renderDocx("Hello world")

		p := doc.findParagraph("Hello world")
		Expect(p).ToNot(BeNil())
		// All text must be present
		Expect(p.text()).To(Equal("Hello world"))
		// The entire text should be in a single run, not split across two
		Expect(p.Runs).To(HaveLen(1),
			"adjacent plain text should be merged into one w:r by paragraphBuilder")
		Expect(p.Runs[0].Text).To(Equal("Hello world"))
	})

	It("should not merge a bold run with adjacent plain runs", func() {
		// Bold and plain runs have different runStyles, so they must remain separate.
		doc := renderDocx("before *bold* after")

		p := doc.findParagraph("before")
		Expect(p).ToNot(BeNil())

		// Exactly three runs: plain, bold, plain
		Expect(p.Runs).To(HaveLen(3),
			"plain+bold+plain must produce exactly 3 runs — merging must not cross style boundaries")

		Expect(p.Runs[0].Bold).To(BeFalse())
		Expect(p.Runs[1].Bold).To(BeTrue())
		Expect(p.Runs[1].Text).To(ContainSubstring("bold"))
		Expect(p.Runs[2].Bold).To(BeFalse())
	})

	Describe("OOXML hygiene: empty <w:br/>-only paragraphs", func() {
		// A paragraph whose inline elements collapse to line breaks plus empty
		// text (e.g. stacked `{empty} +` whitespace lines used in the MSA
		// signature blocks, or stray hard breaks at the head/tail of an
		// otherwise empty paragraph) must render as a clean empty `<w:p></w:p>`
		// rather than `<w:p><w:r><w:br/></w:r>...</w:p>`. A hard line break
		// inside prose (`text +\nmore`) must still emit `<w:r><w:br/></w:r>`
		// between the surrounding text runs.

		// nonBreakText returns the paragraph text with `\n` chars stripped —
		// `<w:br/>` runs are surfaced by parseRunElement as "\n", but for this
		// hygiene check we want only the actual `<w:t>` content.
		nonBreakText := func(p parsedParagraph) string {
			t := p.text()
			return strings.ReplaceAll(t, "\n", "")
		}

		expectNoEmptyBreakOnlyParagraph := func(doc renderedDocx) {
			for _, p := range doc.parseParagraphs() {
				if nonBreakText(p) != "" {
					continue
				}
				if len(p.Bookmarks) > 0 {
					continue
				}
				hasBreakRun := false
				hasDrawing := false
				hasFootnoteRef := false
				for _, r := range p.Runs {
					if strings.Contains(r.Text, "\n") {
						hasBreakRun = true
					}
					if r.HasDrawing {
						hasDrawing = true
					}
					if r.FootnoteRef {
						hasFootnoteRef = true
					}
				}
				if hasDrawing || hasFootnoteRef {
					continue
				}
				Expect(hasBreakRun).To(BeFalse(),
					"found an empty paragraph containing only <w:br/> runs — collapse-to-line-breaks paragraphs must render as empty <w:p></w:p>")
			}
		}

		It("should not emit empty <w:br/>-only paragraphs for a list with continuation across two blocks", func() {
			doc := renderDocx(`. First item with prose
+
[loweralpha]
.. nested item
+
some continuation prose

. Second item`)
			expectNoEmptyBreakOnlyParagraph(doc)
		})

		It("should preserve a hard line break inside prose as a w:br within a single paragraph", func() {
			doc := renderDocx("first +\nsecond")

			p := doc.findParagraph("first")
			Expect(p).ToNot(BeNil())
			Expect(p.text()).To(ContainSubstring("first"))
			Expect(p.text()).To(ContainSubstring("second"))
			hasBreak := false
			for _, r := range p.Runs {
				if strings.Contains(r.Text, "\n") {
					hasBreak = true
				}
			}
			Expect(hasBreak).To(BeTrue(), "hard line break in prose must still emit <w:br/> between runs")
		})

		It("should not emit empty <w:br/> paragraphs when a list item has a nested-list continuation", func() {
			doc := renderDocx(`. parent item
+
.. nested item

. next parent`)
			expectNoEmptyBreakOnlyParagraph(doc)
		})

		It("should not emit empty <w:br/> paragraphs when a list item has a code-block continuation", func() {
			doc := renderDocx(". item\n+\n----\ncode line\n----\n\n. next item")
			expectNoEmptyBreakOnlyParagraph(doc)
		})

		It("should not emit empty <w:br/> paragraphs across multi-block continuations on a single item", func() {
			doc := renderDocx(`. parent item
+
first continuation paragraph
+
second continuation paragraph

. next parent`)
			expectNoEmptyBreakOnlyParagraph(doc)
		})

		It("should not emit empty <w:br/> paragraphs when a list item has a blockquote continuation", func() {
			doc := renderDocx(". item\n+\n____\nquoted prose\n____\n\n. next")
			expectNoEmptyBreakOnlyParagraph(doc)
		})

		It("should not emit an empty paragraph for stacked `{empty} +` whitespace lines", func() {
			// MSA signature blocks use `{empty} +` lines to inject vertical
			// whitespace; the resulting paragraph elements collapse to nothing
			// but a sequence of LineBreak inline nodes. The renderer must not
			// emit a `<w:p>` consisting solely of `<w:br/>` runs.
			doc := renderDocx(`Before signature.

{empty} +
{empty} +

After signature.`)
			expectNoEmptyBreakOnlyParagraph(doc)
		})
	})
})
