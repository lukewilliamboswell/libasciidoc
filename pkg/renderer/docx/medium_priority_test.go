package docx_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bytesparadise/libasciidoc/testsupport"
)

var _ = Describe("medium priority features", func() {

	Context("nested inline formatting", func() {

		It("should render bold+italic on the same run", func() {
			doc := renderDocx(`This is *_bold italic_* text.`)

			r := doc.findRun("bold italic")
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeTrue())
			Expect(r.Italic).To(BeTrue())
		})

		It("should render bold+monospace on the same run", func() {
			doc := renderDocx("This is *`bold code`* text.")

			r := doc.findRun("bold code")
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeTrue())
			Expect(r.Monospace).To(BeTrue())
		})

		It("should render italic+monospace on the same run", func() {
			doc := renderDocx("This is _`italic code`_ text.")

			r := doc.findRun("italic code")
			Expect(r).ToNot(BeNil())
			Expect(r.Italic).To(BeTrue())
			Expect(r.Monospace).To(BeTrue())
		})
	})

	Context("multiple authors", func() {

		It("should render multiple authors separated by semicolons", func() {
			source := `= Agreement
Alice Smith <alice@example.com>; Bob Jones <bob@example.com>

Content here.`
			metadata, result, err := testsupport.RenderDOCXWithMetadata(source)
			Expect(err).ToNot(HaveOccurred())
			doc := openDocx(result)

			Expect(doc.text()).To(ContainSubstring("Alice Smith"))
			Expect(doc.text()).To(ContainSubstring("Bob Jones"))
			Expect(metadata.Authors).To(HaveLen(2))
		})
	})

	Context("external cross references", func() {

		It("should render a cross reference to another document", func() {
			doc := renderDocx(`See xref:other-doc.adoc[Other Document].`)

			Expect(doc.text()).To(ContainSubstring("Other Document"))
		})
	})

	Context("passthrough", func() {

		It("should render inline passthrough content", func() {
			doc := renderDocx(`The value is pass:[some raw content].`)

			Expect(doc.text()).To(ContainSubstring("some raw content"))
		})

		It("should render triple-plus passthrough", func() {
			doc := renderDocx(`The value is +++raw stuff+++.`)

			Expect(doc.text()).To(ContainSubstring("raw stuff"))
		})
	})

	Context("thematic breaks", func() {

		It("should render a thematic break as a separate paragraph", func() {
			doc := renderDocx(`before

'''

after`)

			Expect(doc.text()).To(ContainSubstring("before"))
			Expect(doc.text()).To(ContainSubstring("after"))
			// The thematic break should be its own paragraph with the dash chars
			breakPara := doc.findParagraph("\u2500")
			Expect(breakPara).ToNot(BeNil())
		})
	})

	Context("links", func() {

		It("should render a bare URL as a hyperlink with relationship", func() {
			doc := renderDocx(`Visit https://example.com for details.`)

			p := doc.findParagraph("Visit")
			Expect(p).ToNot(BeNil())
			Expect(p.Links).ToNot(BeEmpty())
			rel := doc.findRelationshipByID(p.Links[0].RelID)
			Expect(rel).ToNot(BeNil())
			Expect(rel.Target).To(Equal("https://example.com"))
			Expect(rel.TargetMode).To(Equal("External"))
		})

		It("should render a link with plain text label in the hyperlink runs", func() {
			doc := renderDocx(`Check the https://example.com[project website].`)

			p := doc.findParagraph("project website")
			Expect(p).ToNot(BeNil())
			Expect(p.Links).ToNot(BeEmpty())
			// The link text should be in the hyperlink's runs
			linkText := ""
			for _, r := range p.Links[0].Runs {
				linkText += r.Text
			}
			Expect(linkText).To(ContainSubstring("project website"))
		})

		It("should render multiple links with separate relationships", func() {
			doc := renderDocx(`See https://a.com[Site A] and https://b.com[Site B].`)

			relA := doc.findRelationship("https://a.com")
			relB := doc.findRelationship("https://b.com")
			Expect(relA).ToNot(BeNil())
			Expect(relB).ToNot(BeNil())
			Expect(relA.ID).ToNot(Equal(relB.ID))
		})
	})

	Context("preamble", func() {

		It("should render preamble content before sections", func() {
			doc := renderDocx(`= Document Title

This is the preamble.

== First Section

Section content.`)

			paras := doc.parseParagraphs()
			preambleIdx := -1
			sectionIdx := -1
			for i, p := range paras {
				if strings.Contains(p.text(), "This is the preamble") {
					preambleIdx = i
				}
				if strings.Contains(p.text(), "First Section") {
					sectionIdx = i
				}
			}
			Expect(preambleIdx).To(BeNumerically(">", 0))
			Expect(sectionIdx).To(BeNumerically(">", preambleIdx))
		})
	})

	Context("document revision", func() {

		It("should include revision in metadata", func() {
			source := `= Agreement
Author Name
v2.1, 2024-03-15: Updated terms

Content.`
			metadata, _, err := testsupport.RenderDOCXWithMetadata(source)
			Expect(err).ToNot(HaveOccurred())

			Expect(metadata.Revision.Revnumber).To(Equal("2.1"))
			Expect(metadata.Revision.Revdate).To(Equal("2024-03-15"))
			Expect(metadata.Revision.Revremark).To(Equal("Updated terms"))
		})
	})
})
