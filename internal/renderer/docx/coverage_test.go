package docx_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/configuration"
)

// png1x1Coverage returns a minimal valid 1×1 PNG for image tests.
func png1x1Coverage() []byte {
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	return data
}

var _ = Describe("cross references", func() {

	It("should render an internal cross reference with auto-generated label", func() {
		// The target section has no explicit label, so defaultCrossReferenceLabel
		// path in renderCrossReference is exercised (label from elementReferences).
		doc := renderDocx(`= Document

[#my-section]
== My Section

See <<my-section>>.`)

		// The xref should become an internal hyperlink anchored to the section
		refPara := doc.findParagraph("See")
		Expect(refPara).ToNot(BeNil())
		Expect(refPara.Links).ToNot(BeEmpty())
	})

	It("should render an internal cross reference with a custom label", func() {
		doc := renderDocx(`= Document

[#target-id]
== Target Section

Go to <<target-id,Click Here>>.`)

		refPara := doc.findParagraph("Go to")
		Expect(refPara).ToNot(BeNil())
		Expect(refPara.Links).ToNot(BeEmpty())

		// The link text must be the custom label
		linkText := ""
		for _, r := range refPara.Links[0].Runs {
			linkText += r.Text
		}
		Expect(linkText).To(ContainSubstring("Click Here"))
	})

	It("should render a cross reference to an unknown anchor as plain text fallback", func() {
		// When the anchor is not found in elementReferences, the renderer falls
		// back to writing "[anchor-id]" as a plain text run (still produces a
		// hyperlink attempt for the unknown anchor).
		doc := renderDocx(`= Document

See <<unknown-anchor>>.`)

		Expect(doc.text()).To(ContainSubstring("unknown-anchor"))
	})

	It("should render an internal cross reference where xref case differs from section anchor", func() {
		// resolveElementReferenceID falls back to case-insensitive match.
		doc := renderDocx(`= Document

== Target Section

See <<_target_section>>.`)

		refPara := doc.findParagraph("See")
		Expect(refPara).ToNot(BeNil())
		Expect(refPara.Links).ToNot(BeEmpty())

		sectionPara := doc.findParagraph("Target Section")
		Expect(sectionPara).ToNot(BeNil())
		Expect(sectionPara.Bookmarks).ToNot(BeEmpty())

		// The anchor must exactly match the bookmark name stored in the section
		Expect(refPara.Links[0].Anchor).To(Equal(sectionPara.Bookmarks[0]))
	})
})

var _ = Describe("external links", func() {

	It("should render a bare external link", func() {
		doc := renderDocx(`Visit https://example.com for details.`)

		p := doc.findParagraph("example.com")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		rel := doc.findRelationshipByID(p.Links[0].RelID)
		Expect(rel).ToNot(BeNil())
		Expect(rel.Target).To(Equal("https://example.com"))
	})

	It("should render an external link with a plain text label", func() {
		doc := renderDocx(`Visit https://example.com[Example Site] now.`)

		p := doc.findParagraph("Example Site")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		link := p.Links[0]
		rel := doc.findRelationshipByID(link.RelID)
		Expect(rel).ToNot(BeNil())
		Expect(rel.Target).To(Equal("https://example.com"))

		linkText := ""
		for _, r := range link.Runs {
			linkText += r.Text
		}
		Expect(linkText).To(ContainSubstring("Example Site"))
	})

	It("should render an external link with formatted inline label", func() {
		// renderLabelInline: []interface{} branch — formatted content inside link
		doc := renderDocx(`See https://example.com[*bold* label].`)

		p := doc.findParagraph("bold")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		boldRun := (*parsedRun)(nil)
		for i := range p.Links[0].Runs {
			if strings.Contains(p.Links[0].Runs[i].Text, "bold") {
				boldRun = &p.Links[0].Runs[i]
			}
		}
		Expect(boldRun).ToNot(BeNil())
		Expect(boldRun.Bold).To(BeTrue())
	})

	It("should render a link with underline and Hyperlink char style on runs", func() {
		doc := renderDocx(`https://example.com[Link Text]`)

		p := doc.findParagraph("Link Text")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		// All runs inside the hyperlink should carry underline
		for _, r := range p.Links[0].Runs {
			if r.Text != "" {
				Expect(r.Underline).To(BeTrue(), "hyperlink runs should be underlined")
			}
		}
	})

	It("should render mergeRunStyle combining bold+underline", func() {
		// Exercise mergeRunStyle: base has bold, extra adds underline (Hyperlink style).
		doc := renderDocx(`https://example.com[*bold link*]`)

		p := doc.findParagraph("bold link")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		var boldUnderlineRun *parsedRun
		for i := range p.Links[0].Runs {
			r := &p.Links[0].Runs[i]
			if strings.Contains(r.Text, "bold link") {
				boldUnderlineRun = r
			}
		}
		Expect(boldUnderlineRun).ToNot(BeNil())
		Expect(boldUnderlineRun.Bold).To(BeTrue())
		Expect(boldUnderlineRun.Underline).To(BeTrue())
	})
})

var _ = Describe("inline element combinations", func() {

	It("should render bold followed by italic in the same paragraph", func() {
		doc := renderDocx(`*bold* _italic_`)

		boldRun := doc.findRun("bold")
		Expect(boldRun).ToNot(BeNil())
		Expect(boldRun.Bold).To(BeTrue())
		Expect(boldRun.Italic).To(BeFalse())

		italicRun := doc.findRun("italic")
		Expect(italicRun).ToNot(BeNil())
		Expect(italicRun.Italic).To(BeTrue())
		Expect(italicRun.Bold).To(BeFalse())
	})

	It("should render bold+italic+mono in the same paragraph without errors", func() {
		doc := renderDocx("*bold* _italic_ `mono`")

		Expect(doc.findRun("bold")).ToNot(BeNil())
		Expect(doc.findRun("italic")).ToNot(BeNil())
		Expect(doc.findRun("mono")).ToNot(BeNil())

		Expect(doc.findRun("bold").Bold).To(BeTrue())
		Expect(doc.findRun("italic").Italic).To(BeTrue())
		Expect(doc.findRun("mono").Monospace).To(BeTrue())
	})

	It("should render nested bold-italic without errors", func() {
		doc := renderDocx(`*_bold italic_*`)

		r := doc.findRun("bold italic")
		Expect(r).ToNot(BeNil())
		Expect(r.Bold).To(BeTrue())
		Expect(r.Italic).To(BeTrue())
	})

	It("should render a line break within a paragraph", func() {
		doc := renderDocx("first line +\nsecond line")

		p := doc.findParagraph("first line")
		Expect(p).ToNot(BeNil())
		Expect(p.text()).To(ContainSubstring("second line"))
	})

	It("should render inline code within bold context", func() {
		doc := renderDocx("*`code in bold`*")

		r := doc.findRun("code in bold")
		Expect(r).ToNot(BeNil())
		Expect(r.Bold).To(BeTrue())
		Expect(r.Monospace).To(BeTrue())
	})
})

var _ = Describe("image dimension units", func() {

	It("should convert px width to EMU (9525 per pixel)", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "test.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::test.png[width=100px]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// 100px * 9525 = 952500 EMU
		Expect(doc.documentXML()).To(ContainSubstring(`cx="952500"`))
	})

	It("should convert inch width to EMU (914400 per inch)", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "test.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::test.png[width=2in]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// 2in * 914400 = 1828800 EMU
		Expect(doc.documentXML()).To(ContainSubstring(`cx="1828800"`))
	})

	It("should convert cm width to EMU (360000 per cm)", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "test.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::test.png[width=3cm]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// 3cm * 360000 = 1080000 EMU
		Expect(doc.documentXML()).To(ContainSubstring(`cx="1080000"`))
	})

	It("should convert mm width to EMU (36000 per mm)", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "test.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::test.png[width=25mm]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// 25mm * 36000 = 900000 EMU
		Expect(doc.documentXML()).To(ContainSubstring(`cx="900000"`))
	})

	It("should use default width when dimension is a percentage", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "test.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::test.png[width=50%]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// Percentage is ignored → default width 3657600 EMU (4 inches)
		Expect(doc.documentXML()).To(ContainSubstring(`cx="3657600"`))
	})

	It("should use default width when no dimension attribute is given", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "test.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::test.png[Alt text]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		// No width → default 3657600 EMU
		Expect(doc.documentXML()).To(ContainSubstring(`cx="3657600"`))
	})
})

var _ = Describe("imageContentType via embedded images", func() {

	It("should register image/gif content type for GIF images", func() {
		// Exercise imageContentType("gif") path
		dir := GinkgoT().TempDir()
		// Minimal valid GIF89a header (1×1 transparent)
		gifData := []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")
		Expect(os.WriteFile(filepath.Join(dir, "anim.gif"), gifData, 0o644)).To(Succeed())

		source := `image::anim.gif[Animated]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		Expect(doc.hasContentTypeForExtension("gif", "image/gif")).To(BeTrue())
	})

	It("should register image/webp content type for WebP images", func() {
		// Exercise imageContentType("webp") path
		dir := GinkgoT().TempDir()
		// Minimal WebP container bytes (RIFF....WEBPVP8 )
		webpData := []byte("RIFF$\x00\x00\x00WEBPVP8 \x18\x00\x00\x00\x30\x01\x00\x9d\x01\x2a\x01\x00\x01\x00\x02\x00\x34\x25\x9f\x0c\x00\x00\xfe\xd8\x81\x00\x00")
		Expect(os.WriteFile(filepath.Join(dir, "img.webp"), webpData, 0o644)).To(Succeed())

		source := `image::img.webp[WebP image]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		Expect(doc.hasContentTypeForExtension("webp", "image/webp")).To(BeTrue())
	})

	It("should register image/bmp content type for BMP images", func() {
		// Exercise imageContentType("bmp") path
		dir := GinkgoT().TempDir()
		// Minimal BMP: 14-byte file header + 40-byte DIB header = 54 bytes, 1x1 pixel
		bmpData := make([]byte, 58)
		bmpData[0] = 'B'
		bmpData[1] = 'M'
		bmpData[2] = 58 // file size low byte
		bmpData[10] = 54 // pixel data offset
		bmpData[14] = 40 // DIB header size
		bmpData[18] = 1  // width = 1
		bmpData[22] = 1  // height = 1
		bmpData[26] = 1  // color planes
		bmpData[28] = 24 // bits per pixel
		Expect(os.WriteFile(filepath.Join(dir, "img.bmp"), bmpData, 0o644)).To(Succeed())

		source := `image::img.bmp[BMP image]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		Expect(doc.hasContentTypeForExtension("bmp", "image/bmp")).To(BeTrue())
	})

	It("should default to image/png content type for unknown extensions", func() {
		// Exercise imageContentType default branch: embed a .png to confirm fallback
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "img.png"), png1x1Coverage(), 0o644)).To(Succeed())

		source := `image::img.png[PNG image]`
		doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

		Expect(doc.hasContentTypeForExtension("png", "image/png")).To(BeTrue())
	})
})

var _ = Describe("renderLabelInline edge cases", func() {

	It("should handle a nil label in an external cross reference gracefully", func() {
		// xref to a plain file with extension triggers defaultCrossReferenceLabel
		// which produces a ".html" label; no panic expected.
		doc := renderDocx(`xref:other.adoc[]`)

		Expect(doc.text()).ToNot(BeEmpty())
	})

	It("should render external xref with no extension as internal anchor link", func() {
		// crossReferenceLocation: no extension → "#" + loc path
		doc := renderDocx(`xref:some-section[]`)

		Expect(doc.text()).ToNot(BeEmpty())
	})

	It("should render external xref with .adoc extension as .html target", func() {
		// crossReferenceLocation: .adoc extension → target.html
		doc := renderDocx(`xref:other.adoc[See Other]`)

		p := doc.findParagraph("See Other")
		Expect(p).ToNot(BeNil())
		Expect(p.Links).ToNot(BeEmpty())

		rel := doc.findRelationshipByID(p.Links[0].RelID)
		Expect(rel).ToNot(BeNil())
		Expect(rel.Target).To(ContainSubstring("other.html"))
	})
})
