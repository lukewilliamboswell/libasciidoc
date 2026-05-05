package docx_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("running header and footer", func() {

	Context("no theme content", func() {

		It("should not emit any header or footer parts, references, or evenAndOdd setting", func() {
			doc := renderDocx(`Hello`)
			Expect(doc.files).ToNot(HaveKey("word/header1.xml"))
			Expect(doc.files).ToNot(HaveKey("word/header2.xml"))
			Expect(doc.files).ToNot(HaveKey("word/footer1.xml"))
			Expect(doc.files).ToNot(HaveKey("word/footer2.xml"))
			Expect(doc.documentXML()).ToNot(ContainSubstring(`<w:headerReference`))
			Expect(doc.documentXML()).ToNot(ContainSubstring(`<w:footerReference`))
			Expect(doc.documentXML()).ToNot(ContainSubstring(`<w:titlePg/>`))
			Expect(doc.settingsXML()).ToNot(ContainSubstring(`<w:evenAndOddHeaders`))
			ct := doc.contentTypesXML()
			Expect(ct).ToNot(ContainSubstring(`/word/header1.xml`))
			Expect(ct).ToNot(ContainSubstring(`/word/footer1.xml`))
		})
	})

	Context("recto-only footer", func() {

		It("should emit footer1.xml with only the default reference", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  recto:
    center: 'Hello'
`)
			Expect(doc.files).To(HaveKey("word/footer1.xml"))
			Expect(doc.files).ToNot(HaveKey("word/footer2.xml"))

			footer := string(doc.files["word/footer1.xml"])
			Expect(footer).To(ContainSubstring(`<w:ftr`))
			Expect(footer).To(ContainSubstring(`Hello`))

			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`<w:footerReference w:type="default"`))
			Expect(docXML).ToNot(ContainSubstring(`<w:footerReference w:type="even"`))
			Expect(docXML).ToNot(ContainSubstring(`<w:headerReference`))

			Expect(doc.settingsXML()).ToNot(ContainSubstring(`<w:evenAndOddHeaders`))
		})
	})

	Context("recto and verso footer", func() {

		It("should emit both footer parts, both references, and evenAndOddHeaders", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  recto:
    center: 'Recto centre'
  verso:
    center: 'Verso centre'
`)
			Expect(doc.files).To(HaveKey("word/footer1.xml"))
			Expect(doc.files).To(HaveKey("word/footer2.xml"))

			Expect(string(doc.files["word/footer1.xml"])).To(ContainSubstring(`Recto centre`))
			Expect(string(doc.files["word/footer2.xml"])).To(ContainSubstring(`Verso centre`))

			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`<w:footerReference w:type="default"`))
			Expect(docXML).To(ContainSubstring(`<w:footerReference w:type="even"`))

			Expect(doc.settingsXML()).To(ContainSubstring(`<w:evenAndOddHeaders/>`))
		})
	})

	Context("three-position layout", func() {

		It("should produce a paragraph with two tab stops and three text segments", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  recto:
    left: 'L'
    center: 'C'
    right: 'R'
`)
			footer := string(doc.files["word/footer1.xml"])

			// Two tab stops: one centred, one right-aligned.
			Expect(footer).To(ContainSubstring(`<w:tab w:val="center" w:pos="`))
			Expect(footer).To(ContainSubstring(`<w:tab w:val="right" w:pos="`))

			// The three labels appear in order, separated by tab runs.
			lIdx := strings.Index(footer, `>L<`)
			cIdx := strings.Index(footer, `>C<`)
			rIdx := strings.Index(footer, `>R<`)
			Expect(lIdx).To(BeNumerically(">", 0))
			Expect(cIdx).To(BeNumerically(">", lIdx))
			Expect(rIdx).To(BeNumerically(">", cIdx))

			// Two <w:tab/> separator runs between the three slots.
			Expect(strings.Count(footer, `<w:tab/>`)).To(Equal(2))
		})

		It("should place the centre tab at half the printed width and the right tab at the full width", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  recto:
    left: 'L'
    center: 'C'
    right: 'R'
`)
			footer := string(doc.files["word/footer1.xml"])

			// Default A4 with 20mm side margins -> printed width = 11906 - 2*1134 = 9638 twips.
			// Centre tab = 9638 / 2 = 4819 (integer truncation).
			Expect(footer).To(ContainSubstring(`<w:tab w:val="center" w:pos="4819"`))
			Expect(footer).To(ContainSubstring(`<w:tab w:val="right" w:pos="9638"`))
		})
	})

	Context("attribute interpolation", func() {

		It("should substitute {revnumber} from the document attribute set", func() {
			doc := renderDocxWithTheme(`= Doc
:revnumber: 1.0

Body.`, `
footer:
  recto:
    center: 'rev {revnumber}'
`)
			footer := string(doc.files["word/footer1.xml"])
			Expect(footer).To(ContainSubstring(`rev `))
			Expect(footer).To(ContainSubstring(`1.0`))
			Expect(footer).ToNot(ContainSubstring(`{revnumber}`))
		})
	})

	Context("page-number fields", func() {

		It("should emit PAGE and NUMPAGES OOXML fields, not literal text", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  recto:
    right: '{page-number} of {page-count}'
`)
			footer := string(doc.files["word/footer1.xml"])

			Expect(footer).To(ContainSubstring(`<w:instrText xml:space="preserve"> PAGE </w:instrText>`))
			Expect(footer).To(ContainSubstring(`<w:instrText xml:space="preserve"> NUMPAGES </w:instrText>`))
			Expect(footer).To(ContainSubstring(`<w:fldChar w:fldCharType="begin"/>`))
			Expect(footer).To(ContainSubstring(`<w:fldChar w:fldCharType="end"/>`))

			Expect(footer).ToNot(ContainSubstring(`{page-number}`))
			Expect(footer).ToNot(ContainSubstring(`{page-count}`))
			Expect(footer).To(ContainSubstring(` of `))
		})
	})

	Context("title page suppression", func() {

		It("should add w:titlePg to sectPr when a running region is set", func() {
			doc := renderDocxWithTheme(`Hello`, `
footer:
  recto:
    center: 'C'
`)
			Expect(doc.documentXML()).To(ContainSubstring(`<w:titlePg/>`))
		})
	})

	Context("content types and relationships", func() {

		It("should declare overrides and relationships for every emitted part", func() {
			doc := renderDocxWithTheme(`Hello`, `
header:
  recto:
    center: 'H1'
  verso:
    center: 'H2'
footer:
  recto:
    center: 'F1'
  verso:
    center: 'F2'
`)
			ct := doc.contentTypesXML()
			Expect(ct).To(ContainSubstring(`PartName="/word/header1.xml"`))
			Expect(ct).To(ContainSubstring(`PartName="/word/header2.xml"`))
			Expect(ct).To(ContainSubstring(`PartName="/word/footer1.xml"`))
			Expect(ct).To(ContainSubstring(`PartName="/word/footer2.xml"`))

			rels := doc.relationshipsXML()
			Expect(rels).To(ContainSubstring(`Target="header1.xml"`))
			Expect(rels).To(ContainSubstring(`Target="header2.xml"`))
			Expect(rels).To(ContainSubstring(`Target="footer1.xml"`))
			Expect(rels).To(ContainSubstring(`Target="footer2.xml"`))

			// Each part must be referenced from sectPr with a valid r:id that
			// resolves to one of the relationships above.
			docXML := doc.documentXML()
			Expect(docXML).To(ContainSubstring(`<w:headerReference w:type="default"`))
			Expect(docXML).To(ContainSubstring(`<w:headerReference w:type="even"`))
			Expect(docXML).To(ContainSubstring(`<w:footerReference w:type="default"`))
			Expect(docXML).To(ContainSubstring(`<w:footerReference w:type="even"`))
		})
	})

	Context("nested content mapping form", func() {

		It("should accept the Asciidoctor PDF nested form recto.left.content", func() {
			doc := renderDocxWithTheme(`= Doc
:revnumber: DRAFT

Body.`, `
footer:
  recto:
    left:
      content: '{revnumber}'
    center:
      content: 'centre'
`)
			footer := string(doc.files["word/footer1.xml"])
			Expect(footer).To(ContainSubstring(`DRAFT`))
			Expect(footer).To(ContainSubstring(`centre`))
		})
	})
})
