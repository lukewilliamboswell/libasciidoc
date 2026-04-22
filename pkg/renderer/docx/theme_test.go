package docx_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
	"github.com/lukewilliamboswell/libasciidoc/testsupport"
)

var _ = Describe("themes", func() {

	Context("default theme (no --theme flag)", func() {

		It("should produce Aptos font at 11pt in styles", func() {
			doc := renderDocx(`Hello`)
			normal := doc.findStyle("Normal")
			Expect(normal).ToNot(BeNil())
			Expect(normal.Size).To(Equal("22")) // 11pt = 22 half-points
		})

		It("should produce A4 page with 20mm margins", func() {
			doc := renderDocx(`Hello`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.PageW).To(Equal("11906"))
			Expect(sp.PageH).To(Equal("16838"))
			Expect(sp.Top).To(Equal("1134"))
			Expect(sp.Right).To(Equal("1134"))
			Expect(sp.Bottom).To(Equal("1134"))
			Expect(sp.Left).To(Equal("1134"))
		})

		It("should produce bold headings with default formula sizes", func() {
			doc := renderDocx(`Hello`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Bold).To(BeTrue())
			Expect(h1.Size).To(Equal("30")) // 32 - 1*2 = 30

			h2 := doc.findStyle("Heading2")
			Expect(h2).ToNot(BeNil())
			Expect(h2.Size).To(Equal("28")) // 32 - 2*2 = 28
		})

		It("should produce Courier New for CodeBlock style", func() {
			doc := renderDocx(`Hello`)
			code := doc.findStyle("CodeBlock")
			Expect(code).ToNot(BeNil())
			Expect(code.Font).To(Equal("Courier New"))
			Expect(code.Size).To(Equal("20")) // 10pt
		})

		It("should produce bold Title at 20pt and italic Subtitle at 12pt", func() {
			doc := renderDocx(`Hello`)
			title := doc.findStyle("Title")
			Expect(title).ToNot(BeNil())
			Expect(title.Bold).To(BeTrue())
			Expect(title.Size).To(Equal("40")) // 20pt
			Expect(title.Color).To(BeEmpty())

			sub := doc.findStyle("Subtitle")
			Expect(sub).ToNot(BeNil())
			Expect(sub.Italic).To(BeTrue())
			Expect(sub.Size).To(Equal("24")) // 12pt
		})
	})

	Context("base theme properties", func() {

		It("should apply base font family and size", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  font_family: Helvetica
  font_size: 10.5
`)
			normal := doc.findStyle("Normal")
			Expect(normal).ToNot(BeNil())
			Expect(normal.Size).To(Equal("21")) // 10.5pt = 21 half-points

			// docDefaults should also use Helvetica
			styles := string(doc.files["word/styles.xml"])
			Expect(styles).To(ContainSubstring(`w:ascii="Helvetica"`))
		})

		It("should apply base font color to Normal style", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  font_color: "333333"
`)
			normal := doc.findStyle("Normal")
			Expect(normal).ToNot(BeNil())
			Expect(normal.Color).To(Equal("333333"))
		})
	})

	Context("heading theme properties", func() {

		It("should apply heading font family", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_family: Georgia
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Font).To(Equal("Georgia"))
		})

		It("should apply all heading sizes h1-h4", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  h1:
    font_size: 16
  h2:
    font_size: 13
  h3:
    font_size: 11
  h4:
    font_size: 10.5
`)
			Expect(doc.findStyle("Heading1").Size).To(Equal("32")) // 16pt
			Expect(doc.findStyle("Heading2").Size).To(Equal("26")) // 13pt
			Expect(doc.findStyle("Heading3").Size).To(Equal("22")) // 11pt
			Expect(doc.findStyle("Heading4").Size).To(Equal("21")) // 10.5pt
			// h5 should fall back to formula: 32 - 5*2 = 22
			Expect(doc.findStyle("Heading5").Size).To(Equal("22"))
		})

		It("should apply heading font color", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_color: "336699"
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Color).To(Equal("336699"))

			h3 := doc.findStyle("Heading3")
			Expect(h3).ToNot(BeNil())
			Expect(h3.Color).To(Equal("336699"))
		})

		It("should apply heading font style italic", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_style: italic
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Italic).To(BeTrue())
			Expect(h1.Bold).To(BeFalse())
		})

		It("should apply heading font style bold_italic", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_style: bold_italic
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Bold).To(BeTrue())
			Expect(h1.Italic).To(BeTrue())
		})
	})

	Context("title page theme properties", func() {

		It("should apply title font size and color", func() {
			doc := renderDocxWithTheme(`= My Title`, `
title_page:
  title:
    font_size: 18
    font_color: "112233"
`)
			title := doc.findStyle("Title")
			Expect(title).ToNot(BeNil())
			Expect(title.Size).To(Equal("36")) // 18pt
			Expect(title.Color).To(Equal("112233"))
		})

		It("should apply subtitle font size and color", func() {
			doc := renderDocxWithTheme(`= My Title`, `
title_page:
  subtitle:
    font_size: 13
    font_color: "333333"
`)
			sub := doc.findStyle("Subtitle")
			Expect(sub).ToNot(BeNil())
			Expect(sub.Size).To(Equal("26"))    // 13pt
			Expect(sub.Color).To(Equal("333333"))
		})
	})

	Context("page layout theme properties", func() {

		It("should apply Letter page size", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  size: Letter
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.PageW).To(Equal("12240"))
			Expect(sp.PageH).To(Equal("15840"))
		})

		It("should apply landscape layout", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  size: A4
  layout: landscape
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.PageW).To(Equal("16838"))
			Expect(sp.PageH).To(Equal("11906"))
		})

		It("should apply custom margins from numeric mm values", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  margin: [25.4, 25.4, 25.4, 25.4]
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.Top).To(Equal("1440"))    // 25.4mm = 1in = 1440 twips
			Expect(sp.Right).To(Equal("1440"))
		})

		It("should parse margins with mm suffix", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  margin: [25mm, 20mm, 20mm, 20mm]
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.Top).To(Equal("1417"))  // 25mm
			Expect(sp.Right).To(Equal("1134")) // 20mm
		})

		It("should parse margins with in suffix", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  margin: [1in, 1in, 1in, 1in]
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.Top).To(Equal("1440")) // 1in = 25.4mm = 1440 twips
		})

		It("should parse margins with pt suffix", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  margin: [72pt, 36pt, 72pt, 36pt]
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.Top).To(Equal("1440"))  // 72pt = 1in = 1440 twips
			Expect(sp.Right).To(Equal("720")) // 36pt = 0.5in = 720 twips
		})
	})

	Context("table theme properties", func() {

		It("should apply table border color and width", func() {
			doc := renderDocxWithTheme(`|===
| A | B
| C | D
|===`, `
table:
  border_color: "000000"
  border_width: 1.0
`)
			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`w:sz="8"`))        // 1.0pt = 8 eighths
			Expect(docXML).To(ContainSubstring(`w:color="000000"`))
		})

		It("should apply table header background color", func() {
			doc := renderDocxWithTheme(`[%header]
|===
| Header 1 | Header 2
| Cell 1 | Cell 2
|===`, `
table:
  head:
    background_color: F0F0F0
`)
			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`w:fill="F0F0F0"`))
		})
	})

	Context("code theme properties", func() {

		It("should apply code font family to CodeBlock style", func() {
			doc := renderDocxWithTheme(`Hello`, `
code:
  font_family: JetBrains Mono
  font_size: 9
`)
			code := doc.findStyle("CodeBlock")
			Expect(code).ToNot(BeNil())
			Expect(code.Font).To(Equal("JetBrains Mono"))
			Expect(code.Size).To(Equal("18")) // 9pt
		})

		It("should apply code font to inline monospace runs", func() {
			doc := renderDocxWithTheme("some `inline code` here", `
code:
  font_family: JetBrains Mono
`)
			r := doc.findRun("inline code")
			Expect(r).ToNot(BeNil())
			Expect(r.Monospace).To(BeTrue())
			Expect(r.MonoFont).To(Equal("JetBrains Mono"))
		})
	})

	Context("list theme properties", func() {

		It("should apply custom list indent", func() {
			doc := renderDocxWithTheme(`. item one
. item two`, `
list:
  indent: 20
`)
			numbering := doc.numberingXML()
			// 20pt = 400 twips base indent (at level 0)
			Expect(numbering).To(ContainSubstring(`w:left="400"`))
		})
	})

	Context("partial theme merging", func() {

		It("should keep defaults for unspecified properties", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  font_family: Arial
`)
			normal := doc.findStyle("Normal")
			Expect(normal).ToNot(BeNil())
			// Size preserved from default (11pt = 22 half-points)
			Expect(normal.Size).To(Equal("22"))

			// Page size preserved
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.PageW).To(Equal("11906")) // A4
		})
	})

	Context("heading text_transform", func() {

		It("should apply uppercase to all headings", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  text_transform: uppercase
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Caps).To(BeTrue())

			h3 := doc.findStyle("Heading3")
			Expect(h3).ToNot(BeNil())
			Expect(h3.Caps).To(BeTrue())
		})

		It("should apply per-level text_transform overrides", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  text_transform: uppercase
  h3:
    font_size: 11
    text_transform: none
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Caps).To(BeTrue())

			h3 := doc.findStyle("Heading3")
			Expect(h3).ToNot(BeNil())
			Expect(h3.Caps).To(BeFalse())
			Expect(h3.Size).To(Equal("22")) // 11pt
		})

		It("should not apply caps when text_transform is not set", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_style: bold
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.Caps).To(BeFalse())
		})
	})

	Context("link theme properties", func() {

		It("should apply custom link color to Hyperlink style", func() {
			doc := renderDocxWithTheme(`Hello`, `
link:
  font_color: "1A1A1A"
`)
			styles := string(doc.files["word/styles.xml"])
			Expect(styles).To(ContainSubstring(`w:styleId="Hyperlink"`))
			Expect(styles).To(ContainSubstring(`w:val="1A1A1A"`))
		})

		It("should use default blue when link color is not set", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  font_family: Arial
`)
			styles := string(doc.files["word/styles.xml"])
			Expect(styles).To(ContainSubstring(`w:val="0563C1"`))
		})
	})

	Context("description list with heading font", func() {

		It("should use heading font for labeled list terms", func() {
			doc := renderDocxWithTheme(`Term:: Description here`, `
heading:
  font_family: Georgia
`)
			r := doc.findRun("Term")
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeTrue())
			Expect(r.Font).To(Equal("Georgia"))
		})
	})

	Context("theme loading errors", func() {

		It("should fail with a missing theme file", func() {
			_, err := testsupportRenderDOCXWithTheme(`Hello`, "/nonexistent/theme.yml")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file"))
		})

		It("should fail with invalid YAML", func() {
			tmpDir := GinkgoT().TempDir()
			themePath := tmpDir + "/bad.yml"
			Expect(os.WriteFile(themePath, []byte("{{{{not yaml"), 0644)).To(Succeed())

			_, err := testsupportRenderDOCXWithTheme(`Hello`, themePath)
			Expect(err).To(HaveOccurred())
		})
	})
})

// renderDocxWithTheme renders AsciiDoc source with an inline theme YAML.
func renderDocxWithTheme(source, themeYAML string) renderedDocx {
	tmpDir := GinkgoT().TempDir()
	themePath := tmpDir + "/test-theme.yml"
	err := os.WriteFile(themePath, []byte(strings.TrimSpace(themeYAML)), 0644)
	Expect(err).ToNot(HaveOccurred())

	return renderDocx(source, configuration.WithThemePath(themePath))
}

// testsupportRenderDOCXWithTheme returns the raw bytes + error (doesn't assert on error).
func testsupportRenderDOCXWithTheme(source, themePath string) ([]byte, error) {
	return testsupport.RenderDOCX(source, configuration.WithThemePath(themePath))
}
