package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("admonitions", func() {

	for _, kind := range []string{"TIP", "IMPORTANT", "WARNING", "CAUTION"} {
		kind := kind // capture
		It("should render a "+kind+" admonition paragraph with bold label", func() {
			doc := renderDocx(kind + `: Some content.`)

			p := doc.findParagraph(kind + ":")
			Expect(p).ToNot(BeNil(), "expected paragraph containing %s:", kind)
			Expect(p.Style).To(Equal("Admonition"))
			Expect(p.Runs).ToNot(BeEmpty())
			Expect(p.Runs[0].Text).To(ContainSubstring(kind + ":"))
			Expect(p.Runs[0].Bold).To(BeTrue())
			Expect(p.text()).To(ContainSubstring("Some content."))
		})
	}

	It("should render an admonition block with label paragraph and body content", func() {
		doc := renderDocx(`[NOTE]
====
This is a note block with
multiple lines.
====`)

		// Label paragraph
		labelPara := doc.findParagraph("NOTE:")
		Expect(labelPara).ToNot(BeNil(), "expected a paragraph with the NOTE: label")
		Expect(labelPara.Style).To(Equal("Admonition"))
		Expect(labelPara.Runs[0].Bold).To(BeTrue())

		// Body content should be in a subsequent paragraph (not the label paragraph)
		bodyPara := doc.findParagraph("note block")
		Expect(bodyPara).ToNot(BeNil(), "expected a paragraph with the body content")
		// The body paragraph is separate from the label
		Expect(bodyPara.text()).ToNot(ContainSubstring("NOTE:"))
	})

	It("should not treat a plain example block as admonition", func() {
		doc := renderDocx(`====
Just an example.
====`)

		for _, p := range doc.parseParagraphs() {
			Expect(p.Style).ToNot(Equal("Admonition"), "plain example block should not use Admonition style")
		}
		Expect(doc.text()).To(ContainSubstring("Just an example."))
	})

	It("should not treat a non-admonition styled example block as admonition", func() {
		// [example] is a valid style on an example block but is NOT an admonition
		doc := renderDocx(`[example]
====
Styled but not admonition.
====`)

		for _, p := range doc.parseParagraphs() {
			Expect(p.Style).ToNot(Equal("Admonition"),
				"example block with non-admonition style should not use Admonition style")
		}
		Expect(doc.text()).To(ContainSubstring("Styled but not admonition."))
		// Should NOT produce a synthetic "EXAMPLE:" label
		Expect(doc.text()).ToNot(ContainSubstring("EXAMPLE:"))
	})

	Describe("AdmonitionLabel character style", func() {

		It("should define an AdmonitionLabel character style in styles.xml", func() {
			doc := renderDocx(`NOTE: hello`)
			style := doc.findStyle("AdmonitionLabel")
			Expect(style).ToNot(BeNil(), "expected an AdmonitionLabel character style")
		})

		It("should reference AdmonitionLabel from the inline NOTE label run", func() {
			doc := renderDocx(`NOTE: body content here.`)
			labelRun := doc.findRun("NOTE:")
			Expect(labelRun).ToNot(BeNil())
			Expect(labelRun.CharStyle).To(Equal("AdmonitionLabel"))
			Expect(labelRun.Text).To(ContainSubstring("NOTE:"))
			// Body must NOT carry the label character style.
			bodyRun := doc.findRun("body content here")
			Expect(bodyRun).ToNot(BeNil())
			Expect(bodyRun.CharStyle).ToNot(Equal("AdmonitionLabel"))
		})

		It("should reference AdmonitionLabel from the block NOTE label paragraph", func() {
			doc := renderDocx(`[NOTE]
====
This is the body.
====`)
			labelPara := doc.findParagraph("NOTE:")
			Expect(labelPara).ToNot(BeNil())
			Expect(labelPara.Runs).ToNot(BeEmpty())
			Expect(labelPara.Runs[0].CharStyle).To(Equal("AdmonitionLabel"))
			bodyPara := doc.findParagraph("This is the body")
			Expect(bodyPara).ToNot(BeNil())
			for _, run := range bodyPara.Runs {
				Expect(run.CharStyle).ToNot(Equal("AdmonitionLabel"))
			}
		})

		for _, kind := range []string{"NOTE", "TIP", "IMPORTANT", "WARNING", "CAUTION"} {
			kind := kind
			It("should reference AdmonitionLabel for "+kind+" block label", func() {
				doc := renderDocx(`[` + kind + `]
====
body
====`)
				labelPara := doc.findParagraph(kind + ":")
				Expect(labelPara).ToNot(BeNil(), "expected %s: label paragraph", kind)
				Expect(labelPara.Runs).ToNot(BeEmpty())
				Expect(labelPara.Runs[0].CharStyle).To(Equal("AdmonitionLabel"))
			})
		}

		It("should apply theme label font_size and text_transform to the AdmonitionLabel style", func() {
			doc := renderDocxWithTheme(`NOTE: hello`, `
admonition:
  label:
    font_size: 9
    text_transform: uppercase
    font_color: "1A1A1A"
`)
			style := doc.findStyle("AdmonitionLabel")
			Expect(style).ToNot(BeNil())
			Expect(style.Caps).To(BeTrue(), "text_transform: uppercase should set <w:caps/>")
			Expect(style.Size).To(Equal("18"), "font_size: 9pt should be 18 half-points")
			Expect(style.Color).To(Equal("1A1A1A"))
		})
	})
})
