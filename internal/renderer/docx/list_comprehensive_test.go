package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("list comprehensive", func() {

	Context("ordered list styles", func() {

		It("should render decimal style by default with correct numId on paragraphs", func() {
			doc := renderDocx(`. first
. second`)

			p := doc.findParagraph("first")
			Expect(p).ToNot(BeNil())
			Expect(p.NumID).ToNot(BeEmpty())
			Expect(p.NumLevel).To(Equal("0"))

			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("decimal"))
		})

		It("should render upperalpha style", func() {
			doc := renderDocx(`[upperalpha]
. Alpha
. Bravo`)

			p := doc.findParagraph("Alpha")
			Expect(p).ToNot(BeNil())
			Expect(p.NumID).ToNot(BeEmpty())

			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("upperLetter"))
		})

		It("should render lowerroman style", func() {
			doc := renderDocx(`[lowerroman]
. one
. two`)

			p := doc.findParagraph("one")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("lowerRoman"))
		})

		It("should render upperroman style", func() {
			doc := renderDocx(`[upperroman]
. One
. Two`)

			p := doc.findParagraph("One")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("upperRoman"))
		})
	})

	Context("nested lists", func() {

		It("should render nested unordered lists with ilvl 0 and separate numId", func() {
			doc := renderDocx(`* level 1a
** level 2a
** level 2b
* level 1b`)

			p1 := doc.findParagraph("level 1a")
			Expect(p1).ToNot(BeNil())
			Expect(p1.NumID).ToNot(BeEmpty())
			Expect(p1.NumLevel).To(Equal("0"))

			p2 := doc.findParagraph("level 2a")
			Expect(p2).ToNot(BeNil())
			Expect(p2.NumID).ToNot(BeEmpty())
			Expect(p2.NumLevel).To(Equal("0"))
			Expect(p2.NumID).ToNot(Equal(p1.NumID), "nested list should have a different numId")
		})

		It("should render nested ordered lists with ilvl 0 and separate numId", func() {
			doc := renderDocx(`. first
.. nested first
.. nested second
. second`)

			pTop := doc.findParagraph("first")
			Expect(pTop).ToNot(BeNil())
			Expect(pTop.NumLevel).To(Equal("0"))

			pNested := doc.findParagraph("nested first")
			Expect(pNested).ToNot(BeNil())
			Expect(pNested.NumLevel).To(Equal("0"))
			Expect(pNested.NumID).ToNot(Equal(pTop.NumID), "nested list should have a different numId")
		})

		It("should render mixed ordered and unordered nesting with different formats", func() {
			doc := renderDocx(`. ordered item
* unordered nested
* another nested
. second ordered`)

			pOrdered := doc.findParagraph("ordered item")
			Expect(pOrdered).ToNot(BeNil())
			Expect(pOrdered.NumID).ToNot(BeEmpty())
			orderedDef := doc.findNumberingDef(pOrdered.NumID)
			Expect(orderedDef).ToNot(BeNil())
			Expect(orderedDef.Levels[0].Format).To(Equal("decimal"))

			pUnordered := doc.findParagraph("unordered nested")
			Expect(pUnordered).ToNot(BeNil())
			Expect(pUnordered.NumID).ToNot(BeEmpty())
			unorderedDef := doc.findNumberingDef(pUnordered.NumID)
			Expect(unorderedDef).ToNot(BeNil())
			Expect(unorderedDef.Levels[0].Format).To(Equal("bullet"))
		})

		It("should use lvlText %1. for all list items regardless of nesting depth", func() {
			doc := renderDocx(`. level 1
.. level 2
... level 3`)

			for _, text := range []string{"level 1", "level 2", "level 3"} {
				p := doc.findParagraph(text)
				Expect(p).ToNot(BeNil(), "expected paragraph for %q", text)
				Expect(p.NumLevel).To(Equal("0"), "paragraph %q should use ilvl 0", text)

				def := doc.findNumberingDef(p.NumID)
				Expect(def).ToNot(BeNil(), "expected numbering def for %q", text)
				Expect(def.Levels[0].LvlText).To(Equal("%1."),
					"paragraph %q should use lvlText %%1. at level 0, not %%N. for higher N", text)
			}
		})

		It("should increase indent for deeper nesting levels", func() {
			doc := renderDocx(`. level 1
.. level 2
... level 3`)

			p1 := doc.findParagraph("level 1")
			Expect(p1).ToNot(BeNil())
			def1 := doc.findNumberingDef(p1.NumID)
			Expect(def1).ToNot(BeNil())

			p2 := doc.findParagraph("level 2")
			Expect(p2).ToNot(BeNil())
			def2 := doc.findNumberingDef(p2.NumID)
			Expect(def2).ToNot(BeNil())

			p3 := doc.findParagraph("level 3")
			Expect(p3).ToNot(BeNil())
			def3 := doc.findNumberingDef(p3.NumID)
			Expect(def3).ToNot(BeNil())

			// Indent at level 0 should increase with nesting depth
			Expect(def1.Levels[0].Indent).To(Equal("720"))  // base indent: 0*360 + 720
			Expect(def2.Levels[0].Indent).To(Equal("1080")) // nested indent: 1*360 + 720
			Expect(def3.Levels[0].Indent).To(Equal("1440")) // deeper indent: 2*360 + 720
		})

		It("should assign correct numbering format at each nesting depth", func() {
			doc := renderDocx(`. arabic
.. alpha
... roman
.... upper alpha
..... upper roman`)

			expected := []struct {
				text   string
				format string
			}{
				{"arabic", "decimal"},
				{"alpha", "lowerLetter"},
				{"roman", "lowerRoman"},
				{"upper alpha", "upperLetter"},
				{"upper roman", "upperRoman"},
			}
			for _, e := range expected {
				p := doc.findParagraph(e.text)
				Expect(p).ToNot(BeNil(), "expected paragraph for %q", e.text)
				def := doc.findNumberingDef(p.NumID)
				Expect(def).ToNot(BeNil(), "expected numbering def for %q", e.text)
				Expect(def.Levels[0].Format).To(Equal(e.format),
					"paragraph %q should use format %q", e.text, e.format)
			}
		})
	})

	Context("callout lists", func() {

		It("should render callout list items with angle-bracket prefixes", func() {
			doc := renderDocx(`[source,go]
----
fmt.Println("hello") // <1>
os.Exit(0)           // <2>
----
<1> Print a greeting
<2> Exit the program`)

			p1 := doc.findParagraph("Print a greeting")
			Expect(p1).ToNot(BeNil())
			Expect(p1.Style).To(Equal("ListParagraph"))
			// The callout prefix should appear as a run in the paragraph
			Expect(p1.text()).To(ContainSubstring("<1>"))

			p2 := doc.findParagraph("Exit the program")
			Expect(p2).ToNot(BeNil())
			Expect(p2.text()).To(ContainSubstring("<2>"))
		})
	})

	Context("list continuation", func() {

		It("should render continuation paragraphs as ListParagraph without numbering", func() {
			doc := renderDocx(`* First item
+
Continuation paragraph for first item.

* Second item`)

			// First item has numbering
			p1 := doc.findParagraph("First item")
			Expect(p1).ToNot(BeNil())
			Expect(p1.NumID).ToNot(BeEmpty())

			// Continuation paragraph should be styled but unnumbered
			pCont := doc.findParagraph("Continuation paragraph")
			Expect(pCont).ToNot(BeNil())
			Expect(pCont.Style).To(Equal("ListParagraph"))
			Expect(pCont.NumID).To(BeEmpty(), "continuation paragraph should not have numId")

			// Second item has numbering
			p2 := doc.findParagraph("Second item")
			Expect(p2).ToNot(BeNil())
			Expect(p2.NumID).ToNot(BeEmpty())
		})
	})

	Context("reversed ordered lists", func() {

		// AsciiDoc spec: [%reversed] makes the list count down from N to 1.
		// OOXML has no native reversed-list flag, so we assign a distinct
		// w:num per item, each with a decreasing w:startOverride.

		It("should render a reversed list with each item carrying its own numId", func() {
			doc := renderDocx(`[%reversed]
. Alpha
. Beta
. Gamma`)

			pA := doc.findParagraph("Alpha")
			Expect(pA).ToNot(BeNil())
			Expect(pA.NumID).ToNot(BeEmpty())

			pB := doc.findParagraph("Beta")
			Expect(pB).ToNot(BeNil())
			Expect(pB.NumID).ToNot(BeEmpty())

			pG := doc.findParagraph("Gamma")
			Expect(pG).ToNot(BeNil())
			Expect(pG.NumID).ToNot(BeEmpty())

			// Each item must have a distinct numId (one per item for per-item startOverride).
			Expect(pA.NumID).ToNot(Equal(pB.NumID))
			Expect(pB.NumID).ToNot(Equal(pG.NumID))
		})

		It("should set startOverride values 3, 2, 1 for a three-item reversed list", func() {
			doc := renderDocx(`[%reversed]
. First
. Second
. Third`)

			// numbering.xml must contain startOverride vals 3, 2, 1 in that order.
			numXML := doc.numberingXML()
			Expect(numXML).To(ContainSubstring(`w:startOverride w:val="3"`))
			Expect(numXML).To(ContainSubstring(`w:startOverride w:val="2"`))
			Expect(numXML).To(ContainSubstring(`w:startOverride w:val="1"`))
		})

		It("should render a reversed loweralpha list", func() {
			doc := renderDocx(`[%reversed,loweralpha]
. One
. Two
. Three`)

			p := doc.findParagraph("One")
			Expect(p).ToNot(BeNil())
			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("lowerLetter"))

			numXML := doc.numberingXML()
			Expect(numXML).To(ContainSubstring(`w:startOverride w:val="3"`))
		})
	})

	Context("bullet format in numbering", func() {

		It("should define bullet format for unordered lists", func() {
			doc := renderDocx(`* item one
* item two`)

			p := doc.findParagraph("item one")
			Expect(p).ToNot(BeNil())
			Expect(p.NumID).ToNot(BeEmpty())

			def := doc.findNumberingDef(p.NumID)
			Expect(def).ToNot(BeNil())
			Expect(def.Levels[0].Format).To(Equal("bullet"))
		})
	})
})
