package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/internal/testsupport"
)

var _ = Describe("sections", func() {

	It("should render a document title with Title style", func() {
		doc := renderDocx(`= My Document Title

Some content here.`)

		p := doc.findParagraph("My Document Title")
		Expect(p).ToNot(BeNil())
		Expect(p.Style).To(Equal("Title"))
	})

	It("should render nested sections with correct heading styles and bookmarks", func() {
		doc := renderDocx(`= Document Title

== Section 1

Content in section 1.

=== Subsection 1.1

Content in subsection.`)

		s1 := doc.findParagraph("Section 1")
		Expect(s1).ToNot(BeNil())
		Expect(s1.Style).To(Equal("Heading2"))
		Expect(s1.Bookmarks).ToNot(BeEmpty())

		s11 := doc.findParagraph("Subsection 1.1")
		Expect(s11).ToNot(BeNil())
		Expect(s11.Style).To(Equal("Heading3"))
		Expect(s11.Bookmarks).ToNot(BeEmpty())
	})

	It("should render section numbers when enabled", func() {
		doc := renderDocx(`= Document Title
:sectnums:

== First

=== Child

== Second`)

		// With legal numbering, section numbers come from w:numPr (not plain text).
		first := doc.findParagraph("First")
		Expect(first).ToNot(BeNil())
		Expect(first.NumID).ToNot(BeEmpty())
		Expect(first.NumLevel).To(Equal("0"))

		child := doc.findParagraph("Child")
		Expect(child).ToNot(BeNil())
		Expect(child.NumID).ToNot(BeEmpty())
		Expect(child.NumLevel).To(Equal("1"))

		second := doc.findParagraph("Second")
		Expect(second).ToNot(BeNil())
		Expect(second.NumID).ToNot(BeEmpty())
		Expect(second.NumLevel).To(Equal("0"))
	})

	It("should set outlineLvl and keepNext on all heading styles", func() {
		doc := renderDocx(`= Doc

== H2

=== H3`)

		h1 := doc.findStyle("Heading1")
		Expect(h1).ToNot(BeNil())
		Expect(h1.OutlineLevel).To(Equal("0"))
		Expect(h1.KeepNext).To(BeTrue())

		h2 := doc.findStyle("Heading2")
		Expect(h2).ToNot(BeNil())
		Expect(h2.OutlineLevel).To(Equal("1"))
		Expect(h2.KeepNext).To(BeTrue())

		h3 := doc.findStyle("Heading3")
		Expect(h3).ToNot(BeNil())
		Expect(h3.OutlineLevel).To(Equal("2"))
		Expect(h3.KeepNext).To(BeTrue())
	})

	It("should render a document with author metadata", func() {
		source := `= Service Agreement
John Doe <john@example.com>

This is the agreement.`
		metadata, result, err := testsupport.RenderDOCXWithMetadata(source)
		Expect(err).ToNot(HaveOccurred())
		doc := openDocx(result)

		title := doc.findParagraph("Service Agreement")
		Expect(title).ToNot(BeNil())
		Expect(title.Style).To(Equal("Title"))

		author := doc.findParagraph("John Doe")
		Expect(author).ToNot(BeNil())
		Expect(author.Style).To(Equal("Subtitle"))

		Expect(metadata.Title).To(Equal("Service Agreement"))
		Expect(metadata.Authors).To(HaveLen(1))
	})
})
