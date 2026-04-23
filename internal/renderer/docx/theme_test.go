package docx_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/configuration"
	"github.com/lukewilliamboswell/libasciidoc/internal/testsupport"
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

		It("should produce default doc defaults spacing", func() {
			doc := renderDocx(`Hello`)
			dd := doc.parseDocDefaults()
			Expect(dd.AfterTwips).To(Equal("160")) // ~8pt
			Expect(dd.LineVal).To(Equal("259"))    // ~1.08
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

		It("should apply base text alignment to docDefaults", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  text_align: justify
`)
			dd := doc.parseDocDefaults()
			Expect(dd.Alignment).To(Equal("both")) // OOXML "both" = justify
		})

		It("should apply base line height to docDefaults", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  line_height: 1.5
`)
			dd := doc.parseDocDefaults()
			Expect(dd.LineVal).To(Equal("360")) // 1.5 * 240 = 360
		})

		It("should apply base font style", func() {
			doc := renderDocxWithTheme(`Hello`, `
base:
  font_style: bold
`)
			normal := doc.findStyle("Normal")
			Expect(normal).ToNot(BeNil())
			Expect(normal.Bold).To(BeTrue())
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

		It("should apply all heading sizes h1-h6", func() {
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
  h5:
    font_size: 9
  h6:
    font_size: 8
`)
			Expect(doc.findStyle("Heading1").Size).To(Equal("32")) // 16pt
			Expect(doc.findStyle("Heading2").Size).To(Equal("26")) // 13pt
			Expect(doc.findStyle("Heading3").Size).To(Equal("22")) // 11pt
			Expect(doc.findStyle("Heading4").Size).To(Equal("21")) // 10.5pt
			Expect(doc.findStyle("Heading5").Size).To(Equal("18")) // 9pt
			Expect(doc.findStyle("Heading6").Size).To(Equal("16")) // 8pt
			// h7 should fall back to formula: 32 - 7*2 = 18, capped at min 20
			Expect(doc.findStyle("Heading7").Size).To(Equal("20"))
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

		It("should apply heading margin top and bottom", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  margin_top: 12
  margin_bottom: 6
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1).ToNot(BeNil())
			Expect(h1.SpaceBefore).To(Equal("240")) // 12pt * 20 = 240 twips
			Expect(h1.SpaceAfter).To(Equal("120"))  // 6pt * 20 = 120 twips
		})

		It("should apply per-level heading font color", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_color: "333333"
  h1:
    font_color: "FF0000"
  h3:
    font_color: "00FF00"
`)
			Expect(doc.findStyle("Heading1").Color).To(Equal("FF0000"))
			Expect(doc.findStyle("Heading2").Color).To(Equal("333333")) // inherits
			Expect(doc.findStyle("Heading3").Color).To(Equal("00FF00"))
		})

		It("should apply per-level heading font style", func() {
			doc := renderDocxWithTheme(`Hello`, `
heading:
  font_style: bold
  h2:
    font_style: italic
  h3:
    font_style: bold_italic
`)
			h1 := doc.findStyle("Heading1")
			Expect(h1.Bold).To(BeTrue())
			Expect(h1.Italic).To(BeFalse())

			h2 := doc.findStyle("Heading2")
			Expect(h2.Bold).To(BeFalse())
			Expect(h2.Italic).To(BeTrue())

			h3 := doc.findStyle("Heading3")
			Expect(h3.Bold).To(BeTrue())
			Expect(h3.Italic).To(BeTrue())
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
			Expect(sub.Size).To(Equal("26")) // 13pt
			Expect(sub.Color).To(Equal("333333"))
		})

		It("should apply title font family", func() {
			doc := renderDocxWithTheme(`Hello`, `
title_page:
  title:
    font_family: Georgia
`)
			title := doc.findStyle("Title")
			Expect(title).ToNot(BeNil())
			Expect(title.Font).To(Equal("Georgia"))
		})

		It("should apply subtitle font family and style", func() {
			doc := renderDocxWithTheme(`Hello`, `
title_page:
  subtitle:
    font_family: Georgia
    font_style: bold
`)
			sub := doc.findStyle("Subtitle")
			Expect(sub).ToNot(BeNil())
			Expect(sub.Font).To(Equal("Georgia"))
			Expect(sub.Bold).To(BeTrue())
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
			Expect(sp.Top).To(Equal("1440")) // 25.4mm = 1in = 1440 twips
			Expect(sp.Right).To(Equal("1440"))
		})

		It("should parse margins with mm suffix", func() {
			doc := renderDocxWithTheme(`Hello`, `
page:
  margin: [25mm, 20mm, 20mm, 20mm]
`)
			sp := doc.parseSectionProps()
			Expect(sp).ToNot(BeNil())
			Expect(sp.Top).To(Equal("1417"))   // 25mm
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

	Context("prose theme properties", func() {

		It("should apply prose margin_bottom to docDefaults", func() {
			doc := renderDocxWithTheme(`Hello`, `
prose:
  margin_bottom: 12
`)
			dd := doc.parseDocDefaults()
			Expect(dd.AfterTwips).To(Equal("240")) // 12pt * 20 = 240 twips
		})

		It("should apply prose text_align to docDefaults", func() {
			doc := renderDocxWithTheme(`Hello`, `
prose:
  text_align: center
`)
			dd := doc.parseDocDefaults()
			Expect(dd.Alignment).To(Equal("center"))
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
			Expect(docXML).To(ContainSubstring(`w:sz="8"`)) // 1.0pt = 8 eighths
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

		It("should apply table cell padding", func() {
			doc := renderDocxWithTheme(`|===
| A | B
|===`, `
table:
  cell_padding: 4
`)
			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`<w:tblCellMar>`))
			Expect(docXML).To(ContainSubstring(`w:w="80"`)) // 4pt * 20 = 80 twips
		})

		It("should apply separate grid color and width", func() {
			doc := renderDocxWithTheme(`|===
| A | B
| C | D
|===`, `
table:
  border_color: "000000"
  border_width: 1.0
  grid_color: "CCCCCC"
  grid_width: 0.5
`)
			docXML := doc.documentXML()
			// Outer borders: 000000 at sz=8
			Expect(docXML).To(ContainSubstring(`<w:top w:val="single" w:sz="8" w:space="0" w:color="000000"/>`))
			// Inside borders: CCCCCC at sz=4
			Expect(docXML).To(ContainSubstring(`<w:insideH w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/>`))
		})

		It("should apply stripe background color", func() {
			doc := renderDocxWithTheme(`[%header]
|===
| H1 | H2
| A1 | A2
| B1 | B2
| C1 | C2
|===`, `
table:
  stripe_background_color: "F5F5F5"
`)
			docXML := doc.documentXML()
			// Stripe is applied to alternating data rows
			Expect(docXML).To(ContainSubstring(`w:fill="F5F5F5"`))
		})

		It("should apply header font style", func() {
			doc := renderDocxWithTheme(`[%header]
|===
| Header 1 | Header 2
| Cell 1 | Cell 2
|===`, `
table:
  head:
    font_style: bold_italic
`)
			r := doc.findRun("Header 1")
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeTrue())
			Expect(r.Italic).To(BeTrue())
		})

		It("should apply footer background color and style", func() {
			doc := renderDocxWithTheme(`[%footer]
|===
| Cell 1 | Cell 2
| Footer 1 | Footer 2
|===`, `
table:
  foot:
    background_color: "E0E0E0"
    font_style: italic
`)
			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`w:fill="E0E0E0"`))
			r := doc.findRun("Footer 1")
			Expect(r).ToNot(BeNil())
			Expect(r.Italic).To(BeTrue())
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

		It("should apply code background color to CodeBlock style", func() {
			doc := renderDocxWithTheme(`Hello`, `
code:
  background_color: "F5F5F5"
`)
			code := doc.findStyle("CodeBlock")
			Expect(code).ToNot(BeNil())
			Expect(code.Shading).To(Equal("F5F5F5"))
		})

		It("should apply code font color to CodeBlock style", func() {
			doc := renderDocxWithTheme(`Hello`, `
code:
  font_color: "333333"
`)
			code := doc.findStyle("CodeBlock")
			Expect(code).ToNot(BeNil())
			Expect(code.Color).To(Equal("333333"))
		})

		It("should apply code border to CodeBlock style", func() {
			doc := renderDocxWithTheme(`Hello`, `
code:
  border_color: "CCCCCC"
  border_width: 1
`)
			code := doc.findStyle("CodeBlock")
			Expect(code).ToNot(BeNil())
			Expect(code.BorderAll).To(Equal("CCCCCC"))
		})

		It("should apply code line height to CodeBlock style", func() {
			doc := renderDocxWithTheme(`Hello`, `
code:
  line_height: 1.2
`)
			code := doc.findStyle("CodeBlock")
			Expect(code).ToNot(BeNil())
			Expect(code.LineSpacing).To(Equal("288")) // 1.2 * 240
		})
	})

	Context("codespan theme properties", func() {

		It("should apply codespan font family to inline monospace", func() {
			doc := renderDocxWithTheme("some `inline` here", `
codespan:
  font_family: Fira Code
`)
			r := doc.findRun("inline")
			Expect(r).ToNot(BeNil())
			Expect(r.Monospace).To(BeTrue())
			Expect(r.Font).To(Equal("Fira Code"))
		})

		It("should apply codespan font color", func() {
			doc := renderDocxWithTheme("some `code` here", `
codespan:
  font_color: "B22222"
`)
			r := doc.findRun("code")
			Expect(r).ToNot(BeNil())
			Expect(r.Color).To(Equal("B22222"))
		})

		It("should apply codespan background color as inline shading", func() {
			doc := renderDocxWithTheme("some `code` here", `
codespan:
  background_color: "F0F0F0"
`)
			r := doc.findRun("code")
			Expect(r).ToNot(BeNil())
			Expect(r.Shading).To(Equal("F0F0F0"))
		})

		It("should prefer codespan font over code font for inline", func() {
			doc := renderDocxWithTheme("some `inline` here", `
code:
  font_family: Courier New
codespan:
  font_family: Fira Code
`)
			r := doc.findRun("inline")
			Expect(r).ToNot(BeNil())
			Expect(r.Font).To(Equal("Fira Code"))
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

		It("should apply marker font color to numbering", func() {
			doc := renderDocxWithTheme(`. item one`, `
list:
  marker_font_color: "FF0000"
`)
			numbering := doc.numberingXML()
			Expect(numbering).To(ContainSubstring(`w:val="FF0000"`))
		})

		It("should apply marker font color to bullet lists", func() {
			doc := renderDocxWithTheme(`* bullet item`, `
list:
  marker_font_color: "0000FF"
`)
			numbering := doc.numberingXML()
			Expect(numbering).To(ContainSubstring(`w:val="0000FF"`))
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

		It("should apply link font style", func() {
			doc := renderDocxWithTheme(`Hello`, `
link:
  font_style: bold
`)
			styles := string(doc.files["word/styles.xml"])
			// Hyperlink style should contain bold
			Expect(styles).To(ContainSubstring(`w:styleId="Hyperlink"`))
			// Check that <w:b/> appears within the Hyperlink style
			hyperlinkIdx := strings.Index(styles, `w:styleId="Hyperlink"`)
			Expect(hyperlinkIdx).To(BeNumerically(">", 0))
			remaining := styles[hyperlinkIdx:]
			Expect(remaining).To(ContainSubstring(`<w:b/>`))
		})

		It("should disable underline when text_decoration is none", func() {
			doc := renderDocxWithTheme(`Hello`, `
link:
  text_decoration: none
`)
			styles := string(doc.files["word/styles.xml"])
			hyperlinkIdx := strings.Index(styles, `w:styleId="Hyperlink"`)
			remaining := styles[hyperlinkIdx:]
			styleEnd := strings.Index(remaining, `</w:style>`)
			hyperlinkStyle := remaining[:styleEnd]
			Expect(hyperlinkStyle).ToNot(ContainSubstring(`<w:u`))
		})
	})

	Context("quote theme properties", func() {

		It("should apply quote border color to Quote style", func() {
			doc := renderDocxWithTheme(`Hello`, `
quote:
  border_color: "AAAAAA"
  border_width: 2
`)
			style := doc.findStyle("Quote")
			Expect(style).ToNot(BeNil())
			Expect(style.BorderLeft).To(Equal("AAAAAA"))
		})

		It("should apply quote font color", func() {
			doc := renderDocxWithTheme(`Hello`, `
quote:
  font_color: "666666"
`)
			style := doc.findStyle("Quote")
			Expect(style).ToNot(BeNil())
			Expect(style.Color).To(Equal("666666"))
		})

		It("should apply quote font style", func() {
			doc := renderDocxWithTheme(`Hello`, `
quote:
  font_style: bold
`)
			style := doc.findStyle("Quote")
			Expect(style).ToNot(BeNil())
			Expect(style.Bold).To(BeTrue())
			Expect(style.Italic).To(BeFalse())
		})

		It("should apply quote font family", func() {
			doc := renderDocxWithTheme(`Hello`, `
quote:
  font_family: Georgia
`)
			style := doc.findStyle("Quote")
			Expect(style).ToNot(BeNil())
			Expect(style.Font).To(Equal("Georgia"))
		})

		It("should apply quote font size", func() {
			doc := renderDocxWithTheme(`Hello`, `
quote:
  font_size: 10
`)
			style := doc.findStyle("Quote")
			Expect(style).ToNot(BeNil())
			Expect(style.Size).To(Equal("20")) // 10pt
		})
	})

	Context("admonition theme properties", func() {

		It("should apply admonition background color to Admonition style", func() {
			doc := renderDocxWithTheme(`Hello`, `
admonition:
  background_color: "FFF3CD"
`)
			style := doc.findStyle("Admonition")
			Expect(style).ToNot(BeNil())
			Expect(style.Shading).To(Equal("FFF3CD"))
		})

		It("should apply admonition border to Admonition style", func() {
			doc := renderDocxWithTheme(`Hello`, `
admonition:
  border_color: "FFD700"
  border_width: 1
`)
			style := doc.findStyle("Admonition")
			Expect(style).ToNot(BeNil())
			Expect(style.BorderAll).To(Equal("FFD700"))
		})

		It("should apply admonition font color", func() {
			doc := renderDocxWithTheme(`Hello`, `
admonition:
  font_color: "856404"
`)
			style := doc.findStyle("Admonition")
			Expect(style).ToNot(BeNil())
			Expect(style.Color).To(Equal("856404"))
		})

		It("should apply admonition label font style and color", func() {
			doc := renderDocxWithTheme(`NOTE: This is a note.`, `
admonition:
  label:
    font_style: bold_italic
    font_color: "FF0000"
`)
			r := doc.findRun("NOTE:")
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeTrue())
			Expect(r.Italic).To(BeTrue())
			Expect(r.Color).To(Equal("FF0000"))
		})
	})

	Context("sidebar theme properties", func() {

		It("should apply Sidebar style to sidebar blocks", func() {
			doc := renderDocxWithTheme(`****
Sidebar content.
****`, `
sidebar:
  background_color: "F0F0F0"
  border_color: "CCCCCC"
`)
			style := doc.findStyle("Sidebar")
			Expect(style).ToNot(BeNil())
			Expect(style.Shading).To(Equal("F0F0F0"))
			Expect(style.BorderAll).To(Equal("CCCCCC"))
		})

		It("should render sidebar block content with Sidebar style", func() {
			doc := renderDocxWithTheme(`****
Sidebar text here.
****`, `
sidebar:
  background_color: "EEEEEE"
`)
			p := doc.findParagraph("Sidebar text here.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Sidebar"))
		})
	})

	Context("example theme properties", func() {

		It("should apply Example style to example blocks", func() {
			doc := renderDocxWithTheme(`====
Example content.
====`, `
example:
  background_color: "F8F9FA"
  border_color: "DEE2E6"
`)
			style := doc.findStyle("Example")
			Expect(style).ToNot(BeNil())
			Expect(style.Shading).To(Equal("F8F9FA"))
			Expect(style.BorderAll).To(Equal("DEE2E6"))
		})

		It("should render example block content with Example style", func() {
			doc := renderDocxWithTheme(`====
Example text here.
====`, `
example:
  background_color: "F0F0F0"
`)
			p := doc.findParagraph("Example text here.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Example"))
		})
	})

	Context("caption theme properties", func() {

		It("should apply caption font size", func() {
			doc := renderDocxWithTheme(`Hello`, `
caption:
  font_size: 9
`)
			style := doc.findStyle("Caption")
			Expect(style).ToNot(BeNil())
			Expect(style.Size).To(Equal("18")) // 9pt
		})

		It("should apply caption font style bold", func() {
			doc := renderDocxWithTheme(`Hello`, `
caption:
  font_style: bold
`)
			style := doc.findStyle("Caption")
			Expect(style).ToNot(BeNil())
			Expect(style.Bold).To(BeTrue())
			Expect(style.Italic).To(BeFalse())
		})

		It("should apply caption font color", func() {
			doc := renderDocxWithTheme(`Hello`, `
caption:
  font_color: "666666"
`)
			style := doc.findStyle("Caption")
			Expect(style).ToNot(BeNil())
			Expect(style.Color).To(Equal("666666"))
		})

		It("should apply caption font family", func() {
			doc := renderDocxWithTheme(`Hello`, `
caption:
  font_family: Georgia
`)
			style := doc.findStyle("Caption")
			Expect(style).ToNot(BeNil())
			Expect(style.Font).To(Equal("Georgia"))
		})

		It("should apply caption text alignment", func() {
			doc := renderDocxWithTheme(`Hello`, `
caption:
  text_align: center
`)
			style := doc.findStyle("Caption")
			Expect(style).ToNot(BeNil())
			Expect(style.Alignment).To(Equal("center"))
		})
	})

	Context("description list theme properties", func() {

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

		It("should apply description list term font style", func() {
			doc := renderDocxWithTheme(`Term:: Description here`, `
description_list:
  term_font_style: italic
`)
			r := doc.findRun("Term")
			Expect(r).ToNot(BeNil())
			Expect(r.Italic).To(BeTrue())
			Expect(r.Bold).To(BeFalse())
		})

		It("should apply description list term font color", func() {
			doc := renderDocxWithTheme(`Term:: Description here`, `
description_list:
  term_font_color: "336699"
`)
			r := doc.findRun("Term")
			Expect(r).ToNot(BeNil())
			Expect(r.Color).To(Equal("336699"))
		})

		It("should apply description list term font family", func() {
			doc := renderDocxWithTheme(`Term:: Description here`, `
description_list:
  term_font_family: Arial
`)
			r := doc.findRun("Term")
			Expect(r).ToNot(BeNil())
			Expect(r.Font).To(Equal("Arial"))
		})

		It("should prefer description_list font over heading font", func() {
			doc := renderDocxWithTheme(`Term:: Description here`, `
heading:
  font_family: Georgia
description_list:
  term_font_family: Arial
`)
			r := doc.findRun("Term")
			Expect(r).ToNot(BeNil())
			Expect(r.Font).To(Equal("Arial"))
		})
	})

	Context("running header/footer", func() {

		It("should create header with page number", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  content: "Page {page-number}"
`)
			Expect(doc.files).To(HaveKey("word/footer1.xml"))
			footer := string(doc.files["word/footer1.xml"])
			Expect(footer).To(ContainSubstring(`Page `))
			Expect(footer).To(ContainSubstring(`PAGE`))
			Expect(footer).To(ContainSubstring(`w:fldCharType="begin"`))
		})

		It("should create header with plain text", func() {
			doc := renderDocxWithTheme(`Hello`, `
header:
  content: "My Document"
`)
			Expect(doc.files).To(HaveKey("word/header1.xml"))
			header := string(doc.files["word/header1.xml"])
			Expect(header).To(ContainSubstring(`My Document`))
		})

		It("should apply header font properties", func() {
			doc := renderDocxWithTheme(`Hello`, `
header:
  content: "Header"
  font_family: Georgia
  font_size: 8
  font_color: "999999"
  font_style: italic
`)
			header := string(doc.files["word/header1.xml"])
			Expect(header).To(ContainSubstring(`Georgia`))
			Expect(header).To(ContainSubstring(`w:val="16"`)) // 8pt = 16 half-pts
			Expect(header).To(ContainSubstring(`w:val="999999"`))
			Expect(header).To(ContainSubstring(`<w:i/>`))
		})

		It("should reference header and footer in sectPr", func() {
			doc := renderDocxWithTheme(`Hello`, `
header:
  content: "Header"
footer:
  content: "Footer"
`)
			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`<w:headerReference`))
			Expect(docXML).To(ContainSubstring(`<w:footerReference`))
		})

		It("should add content types for header and footer", func() {
			doc := renderDocxWithTheme(`Hello`, `
header:
  content: "Header"
footer:
  content: "Footer"
`)
			ct := doc.contentTypesXML()
			Expect(ct).To(ContainSubstring(`header1.xml`))
			Expect(ct).To(ContainSubstring(`footer1.xml`))
		})

		It("should not create header/footer when content is empty", func() {
			doc := renderDocx(`Hello`)
			Expect(doc.files).ToNot(HaveKey("word/header1.xml"))
			Expect(doc.files).ToNot(HaveKey("word/footer1.xml"))
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

	Context("theme loading errors", func() {

		It("should fail with a missing theme file", func() {
			_, err := testsupportRenderDOCXWithTheme(`Hello`, "/nonexistent/theme.yml")
			Expect(err).To(HaveOccurred())
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
