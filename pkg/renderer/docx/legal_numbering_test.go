package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("legal multi-level numbering", func() {

	Context("abstractNum structure", func() {

		It("should create a multilevel abstractNum when sectnums is enabled", func() {
			doc := renderDocx(`= Title
:sectnums:

== First

== Second`)

			// The legal numbering definition should exist
			first := doc.findParagraph("First")
			Expect(first).ToNot(BeNil())
			Expect(first.NumID).ToNot(BeEmpty())

			numDef := doc.findNumberingDef(first.NumID)
			Expect(numDef).ToNot(BeNil())
			Expect(numDef.MultiLevelType).To(Equal("multilevel"))
		})

		It("should define correct formats for each level", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause`)

			p := doc.findParagraph("Clause")
			Expect(p).ToNot(BeNil())
			numDef := doc.findNumberingDef(p.NumID)
			Expect(numDef).ToNot(BeNil())
			Expect(len(numDef.Levels)).To(BeNumerically(">=", 6))

			// Level 0: "1." decimal (clause headings)
			Expect(numDef.Levels[0].Format).To(Equal("decimal"))
			Expect(numDef.Levels[0].LvlText).To(Equal("%1."))
			Expect(numDef.Levels[0].Start).To(Equal("1"))

			// Level 1: "1.1" decimal (sub-clause headings)
			Expect(numDef.Levels[1].Format).To(Equal("decimal"))
			Expect(numDef.Levels[1].LvlText).To(Equal("%1.%2"))

			// Level 2: "1.1.1" decimal (sub-sub-clause headings)
			Expect(numDef.Levels[2].Format).To(Equal("decimal"))
			Expect(numDef.Levels[2].LvlText).To(Equal("%1.%2.%3"))

			// Level 3: "(a)" lowerLetter (enumerated items)
			Expect(numDef.Levels[3].Format).To(Equal("lowerLetter"))
			Expect(numDef.Levels[3].LvlText).To(Equal("(%4)"))

			// Level 4: "(i)" lowerRoman (sub-items)
			Expect(numDef.Levels[4].Format).To(Equal("lowerRoman"))
			Expect(numDef.Levels[4].LvlText).To(Equal("(%5)"))

			// Level 5: "(A)" upperLetter (sub-sub-items)
			Expect(numDef.Levels[5].Format).To(Equal("upperLetter"))
			Expect(numDef.Levels[5].LvlText).To(Equal("(%6)"))
		})

		It("should set lvlRestart on level 3 to restart after sub-clause headings", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause`)

			p := doc.findParagraph("Clause")
			numDef := doc.findNumberingDef(p.NumID)
			Expect(numDef).ToNot(BeNil())

			// Level 3 should restart when level 1 (sub-clause) changes
			Expect(numDef.Levels[3].LvlRestart).To(Equal("1"))
		})
	})

	Context("heading paragraphs", func() {

		It("should assign ilvl=0 to top-level numbered sections", func() {
			doc := renderDocx(`= Title
:sectnums:

== First

== Second`)

			first := doc.findParagraph("First")
			Expect(first).ToNot(BeNil())
			Expect(first.NumLevel).To(Equal("0"))

			second := doc.findParagraph("Second")
			Expect(second).ToNot(BeNil())
			Expect(second.NumLevel).To(Equal("0"))
		})

		It("should assign ilvl=1 to sub-sections", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause

=== SubClause`)

			sub := doc.findParagraph("SubClause")
			Expect(sub).ToNot(BeNil())
			Expect(sub.NumLevel).To(Equal("1"))
		})

		It("should share the same numId across all numbered headings", func() {
			doc := renderDocx(`= Title
:sectnums:

== First

=== Child

== Second`)

			first := doc.findParagraph("First")
			child := doc.findParagraph("Child")
			second := doc.findParagraph("Second")

			Expect(first.NumID).ToNot(BeEmpty())
			Expect(first.NumID).To(Equal(child.NumID))
			Expect(first.NumID).To(Equal(second.NumID))
		})

		It("should not assign numPr to unnumbered sections", func() {
			doc := renderDocx(`= Title

== Preamble

:sectnums:

== Numbered`)

			preamble := doc.findParagraph("Preamble")
			Expect(preamble).ToNot(BeNil())
			Expect(preamble.NumID).To(BeEmpty())

			numbered := doc.findParagraph("Numbered")
			Expect(numbered).ToNot(BeNil())
			Expect(numbered.NumID).ToNot(BeEmpty())
		})
	})

	Context("ordered lists within numbered sections", func() {

		It("should assign ilvl=3 to loweralpha lists under numbered sections", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause

[loweralpha]
. first item
. second item`)

			first := doc.findParagraph("first item")
			Expect(first).ToNot(BeNil())
			Expect(first.NumLevel).To(Equal("3"))
		})

		It("should assign ilvl=4 to nested lowerroman lists", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause

[loweralpha]
. parent item
+
[lowerroman]
.. nested item`)

			nested := doc.findParagraph("nested item")
			Expect(nested).ToNot(BeNil())
			Expect(nested.NumLevel).To(Equal("4"))
		})

		It("should assign ilvl=5 to deeply nested upperalpha lists", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause

[loweralpha]
. parent
+
[lowerroman]
.. middle
+
[upperalpha]
... deep item`)

			deep := doc.findParagraph("deep item")
			Expect(deep).ToNot(BeNil())
			Expect(deep.NumLevel).To(Equal("5"))
		})

		It("should use separate numId for lists vs headings (same abstractNum)", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause

[loweralpha]
. list item`)

			heading := doc.findParagraph("Clause")
			item := doc.findParagraph("list item")

			Expect(heading.NumID).ToNot(BeEmpty())
			Expect(item.NumID).ToNot(BeEmpty())
			// Different w:num instances but same abstractNum
			Expect(heading.NumID).ToNot(Equal(item.NumID))

			headingDef := doc.findNumberingDef(heading.NumID)
			itemDef := doc.findNumberingDef(item.NumID)
			Expect(headingDef.AbstractID).To(Equal(itemDef.AbstractID))
		})
	})

	Context("lists outside numbered sections use regular numbering", func() {

		It("should use separate numbering for lists in unnumbered sections", func() {
			doc := renderDocx(`= Title

== Preamble

[loweralpha]
. item one

:sectnums:

== Clause

[loweralpha]
. clause item`)

			preambleItem := doc.findParagraph("item one")
			Expect(preambleItem).ToNot(BeNil())

			clauseItem := doc.findParagraph("clause item")
			Expect(clauseItem).ToNot(BeNil())

			// Different numIds: preamble uses regular numbering, clause uses legal
			Expect(preambleItem.NumID).ToNot(Equal(clauseItem.NumID))
		})
	})

	Context("numbering restarts", func() {

		It("should give each list its own numId with startOverride for restart", func() {
			doc := renderDocx(`= Title
:sectnums:

== First Clause

[loweralpha]
. item a
. item b

== Second Clause

[loweralpha]
. item c
. item d`)

			itemA := doc.findParagraph("item a")
			itemC := doc.findParagraph("item c")
			Expect(itemA.NumLevel).To(Equal("3"))
			Expect(itemC.NumLevel).To(Equal("3"))

			// Different w:num instances ensure independent restart
			Expect(itemA.NumID).ToNot(Equal(itemC.NumID))

			// Both reference the same legal abstractNum
			defA := doc.findNumberingDef(itemA.NumID)
			defC := doc.findNumberingDef(itemC.NumID)
			Expect(defA.AbstractID).To(Equal(defC.AbstractID))
		})

		It("should restart list numbering after sub-clause headings", func() {
			doc := renderDocx(`= Title
:sectnums:

== Parent

=== Child One

[loweralpha]
. first list item

=== Child Two

[loweralpha]
. second list item`)

			item1 := doc.findParagraph("first list item")
			item2 := doc.findParagraph("second list item")
			Expect(item1.NumLevel).To(Equal("3"))
			Expect(item2.NumLevel).To(Equal("3"))

			// Different w:num instances for independent restart
			Expect(item1.NumID).ToNot(Equal(item2.NumID))
		})
	})

	Context("OOXML conformance", func() {

		It("should place all abstractNums before all nums in numbering.xml", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause

[loweralpha]
. item`)

			numbering := string(doc.files["word/numbering.xml"])
			// Find last abstractNum and first num positions
			lastAbstractNum := lastIndex(numbering, "</w:abstractNum>")
			firstNum := indexOf(numbering, "<w:num ")
			Expect(lastAbstractNum).To(BeNumerically("<", firstNum),
				"All abstractNum elements must precede all num elements")
		})

		It("should emit all levels starting at 1", func() {
			doc := renderDocx(`= Title
:sectnums:

== Clause`)

			p := doc.findParagraph("Clause")
			numDef := doc.findNumberingDef(p.NumID)
			for _, lvl := range numDef.Levels {
				Expect(lvl.Start).To(Equal("1"),
					"Level %s should start at 1", lvl.Level)
			}
		})
	})
})

// helper: find last occurrence of substr
func lastIndex(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// helper: find first occurrence of substr
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
