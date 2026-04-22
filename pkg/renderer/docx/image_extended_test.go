package docx_test

import (
	"encoding/base64"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
)

var _ = Describe("images extended", func() {

	Context("inline images", func() {

		It("should render an inline image as a drawing run within the paragraph", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "icon.png"), png1x1(), 0o644)).To(Succeed())

			source := `Click the image:icon.png[Icon] button to proceed.`
			doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

			Expect(doc.files).To(HaveKey("word/media/image1.png"))

			// The drawing should be in the same paragraph as the surrounding text
			p := doc.findParagraph("Click the")
			Expect(p).ToNot(BeNil())
			hasDrawing := false
			for _, r := range p.Runs {
				if r.HasDrawing {
					hasDrawing = true
				}
			}
			Expect(hasDrawing).To(BeTrue(), "inline image should produce a drawing run in the paragraph")
		})
	})

	Context("image dimensions", func() {

		It("should embed an image with pixel dimensions in EMU", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "photo.png"), png1x1(), 0o644)).To(Succeed())

			source := `image::photo.png[Photo,200,150]`
			doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

			Expect(doc.files).To(HaveKey("word/media/image1.png"))
			xml := doc.documentXML()
			// 200px = 200*9525 = 1905000 EMU
			Expect(xml).To(ContainSubstring(`cx="1905000"`))
			// 150px = 150*9525 = 1428750 EMU
			Expect(xml).To(ContainSubstring(`cy="1428750"`))
		})
	})

	Context("image captions", func() {

		It("should number multiple images sequentially with Caption style", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "a.png"), png1x1(), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "b.png"), png1x1(), 0o644)).To(Succeed())

			source := `.First image
image::a.png[]

.Second image
image::b.png[]`
			doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

			cap1 := doc.findParagraph("Figure 1. First image")
			Expect(cap1).ToNot(BeNil())
			Expect(cap1.Style).To(Equal("Caption"))

			cap2 := doc.findParagraph("Figure 2. Second image")
			Expect(cap2).ToNot(BeNil())
			Expect(cap2.Style).To(Equal("Caption"))
		})
	})

	Context("missing images", func() {

		It("should render an italic placeholder for missing image files", func() {
			doc := renderDocx(`image::nonexistent.png[Alt text]`)

			r := doc.findRun("[image:")
			Expect(r).ToNot(BeNil())
			Expect(r.Italic).To(BeTrue())
			Expect(doc.files).ToNot(HaveKey("word/media/image1.png"))
		})
	})

	Context("JPEG images", func() {

		It("should embed a JPEG image with correct content type and relationship", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "photo.jpg"), jpgMinimal(), 0o644)).To(Succeed())

			source := `image::photo.jpg[Photo]`
			doc := renderDocx(source, configuration.WithFilename(filepath.Join(dir, "doc.adoc")))

			// Media file is stored with .jpeg extension
			Expect(doc.files).To(HaveKey("word/media/image1.jpeg"))

			// Content type registered for jpeg extension
			Expect(doc.hasContentTypeForExtension("jpeg", "image/jpeg")).To(BeTrue(),
				"[Content_Types].xml should have jpeg -> image/jpeg")

			// Relationship points to the media file
			rel := doc.findRelationship("media/image1.jpeg")
			Expect(rel).ToNot(BeNil(), "document.xml.rels should reference media/image1.jpeg")
		})
	})
})

func png1x1() []byte {
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	Expect(err).ToNot(HaveOccurred())
	return data
}

func jpgMinimal() []byte {
	data, _ := base64.StdEncoding.DecodeString("/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//2wBDAP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//AP//wAARCAABAAEDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYI4Q/SFhSRERXYnIDCQoWGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwD+f+gD/9k=")
	return data
}
