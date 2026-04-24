package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lists", func() {

	It("should render an unordered list with bullet numbering on each item", func() {
		doc := renderDocx(`* item one
* item two
* item three`)

		for _, text := range []string{"item one", "item two", "item three"} {
			p := doc.findParagraph(text)
			Expect(p).ToNot(BeNil(), "expected paragraph for %q", text)
			Expect(p.NumID).ToNot(BeEmpty(), "paragraph %q should have numId", text)
		}

		p := doc.findParagraph("item one")
		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("bullet"))
	})

	It("should render an ordered list with decimal numbering on each item", func() {
		doc := renderDocx(`. first
. second
. third`)

		for _, text := range []string{"first", "second", "third"} {
			p := doc.findParagraph(text)
			Expect(p).ToNot(BeNil(), "expected paragraph for %q", text)
			Expect(p.NumID).ToNot(BeEmpty(), "paragraph %q should have numId", text)
		}

		p := doc.findParagraph("first")
		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("decimal"))
	})

	It("should render ordered list with start and loweralpha style", func() {
		doc := renderDocx(`[loweralpha,start=3]
. third
. fourth`)

		p := doc.findParagraph("third")
		Expect(p).ToNot(BeNil())

		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("lowerLetter"))
		Expect(def.Levels[0].Start).To(Equal("3"))
	})

	It("should render checklists with checkbox characters", func() {
		doc := renderDocx(`* [x] done
* [ ] todo`)

		Expect(doc.text()).To(ContainSubstring("\u2611")) // ☑
		Expect(doc.text()).To(ContainSubstring("\u2610")) // ☐
	})

	It("should render a labeled list with bold term and plain definition", func() {
		doc := renderDocx(`Term 1:: Definition 1
Term 2:: Definition 2`)

		termRun := doc.findRun("Term 1")
		Expect(termRun).ToNot(BeNil())
		Expect(termRun.Bold).To(BeTrue())

		defPara := doc.findParagraph("Definition 1")
		Expect(defPara).ToNot(BeNil())
	})

	It("should split a list item with a hard line break into two paragraphs", func() {
		// A hard line break (+ at end of line) inside a list item should produce
		// two separate paragraphs, not a <w:br/> inside a single paragraph.
		doc := renderDocx(`. Line one +
line two`)

		p1 := doc.findParagraph("Line one")
		Expect(p1).ToNot(BeNil())
		Expect(p1.NumID).ToNot(BeEmpty(), "first segment should carry the numbering")

		p2 := doc.findParagraph("line two")
		Expect(p2).ToNot(BeNil())
		Expect(p2.NumID).To(BeEmpty(), "second segment should not carry numbering")

		// They must be distinct paragraphs — not the same paragraph
		Expect(p1.text()).ToNot(ContainSubstring("line two"), "first paragraph should not contain the second line's text")
	})
})
