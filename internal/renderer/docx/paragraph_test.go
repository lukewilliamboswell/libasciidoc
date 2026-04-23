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
})
