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
})
