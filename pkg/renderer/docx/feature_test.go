package docx_test

import (
	"encoding/base64"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
)

var _ = Describe("docx feature coverage", func() {

	It("should render rich link labels with formatting inside the hyperlink", func() {
		doc := renderDocx(`Go to https://example.com[*Example* Site].`)

		p := doc.findParagraph("Example")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		link := p.Links[0]
		rel := doc.findRelationshipByID(link.RelID)
		Expect(rel).ToNot(BeNil())
		Expect(rel.Target).To(Equal("https://example.com"))

		// The bold run should be inside the hyperlink, not just anywhere
		foundBold := false
		for _, r := range link.Runs {
			if r.Text == "Example" {
				Expect(r.Bold).To(BeTrue())
				foundBold = true
			}
		}
		Expect(foundBold).To(BeTrue(), "the 'Example' run inside the hyperlink should be bold")
	})

	It("should render internal cross references with bookmark and anchor link", func() {
		doc := renderDocx(`= Document

== Target Section

See <<_target_section>>.`)

		// The section heading should have a bookmark
		sectionPara := doc.findParagraph("Target Section")
		Expect(sectionPara).ToNot(BeNil())
		Expect(sectionPara.Bookmarks).ToNot(BeEmpty())
		bookmarkName := sectionPara.Bookmarks[0]

		// The cross-reference should be an internal hyperlink whose anchor
		// exactly matches the section bookmark name
		refPara := doc.findParagraph("See")
		Expect(refPara).ToNot(BeNil())
		Expect(refPara.Links).ToNot(BeEmpty())
		Expect(refPara.Links[0].Anchor).To(Equal(bookmarkName),
			"xref anchor should exactly match section bookmark")
	})

	It("should render footnote references and footnote bodies", func() {
		doc := renderDocx(`A statement.footnote:[The supporting note.]`)

		// Find the footnote reference run in the document
		paras := doc.parseParagraphs()
		foundFootnoteRef := false
		for _, p := range paras {
			for _, r := range p.Runs {
				if r.FootnoteRef && r.FootnoteID == "1" {
					foundFootnoteRef = true
				}
			}
		}
		Expect(foundFootnoteRef).To(BeTrue(), "document should contain footnoteReference with id=1")

		// Footnotes XML should exist with the body text
		Expect(doc.files).To(HaveKey("word/footnotes.xml"))
		Expect(doc.footnotesXML()).To(ContainSubstring("The supporting note."))

		// Relationship to footnotes.xml should exist
		rel := doc.findRelationship("footnotes.xml")
		Expect(rel).ToNot(BeNil())
	})

	It("should render table of contents entries", func() {
		doc := renderDocx(`= Document
:toc:
:sectnums:

== First

== Second`)

		tocHeading := doc.findParagraph("Table of Contents")
		Expect(tocHeading).ToNot(BeNil())
		Expect(tocHeading.Style).To(Equal("Heading1"))

		Expect(doc.text()).To(ContainSubstring("1 First"))
		Expect(doc.text()).To(ContainSubstring("2 Second"))
	})

	It("should embed local images and keep later imagesdir changes scoped", func() {
		dir := GinkgoT().TempDir()
		one := filepath.Join(dir, "one")
		two := filepath.Join(dir, "two")
		Expect(os.Mkdir(one, 0o755)).To(Succeed())
		Expect(os.Mkdir(two, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(one, "pixel.png"), png1x1Feat(), 0o644)).To(Succeed())

		source := `:imagesdir: one

image::pixel.png[Pixel]

:imagesdir: two

image::missing.png[Missing]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// First image should be embedded
		Expect(doc.files).To(HaveKey("word/media/image1.png"))

		// Missing image should render as italic fallback text, not an embedded image
		missingRun := doc.findRun("[image:")
		Expect(missingRun).ToNot(BeNil())
		Expect(missingRun.Italic).To(BeTrue())
	})
})

func png1x1Feat() []byte {
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	return data
}
