package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tables extended", func() {

	It("should render a table footer with bold text", func() {
		doc := renderDocx(`[%footer]
|===
| Item | Price

| Widget | $10
| Gadget | $20

| Total | $30
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		table := tables[0]
		Expect(len(table.Rows)).To(BeNumerically(">=", 3))

		// Last row is the footer — its cells should be bold
		lastRow := table.Rows[len(table.Rows)-1]
		Expect(lastRow.Cells).ToNot(BeEmpty())
		for _, cell := range lastRow.Cells {
			for _, p := range cell.Paragraphs {
				for _, r := range p.Runs {
					if r.Text != "" {
						Expect(r.Bold).To(BeTrue(), "footer cell run %q should be bold", r.Text)
					}
				}
			}
		}

		// Verify footer content
		totalRun := doc.findTableCellRun("Total")
		Expect(totalRun).ToNot(BeNil())
		Expect(totalRun.Bold).To(BeTrue())

		priceRun := doc.findTableCellRun("$30")
		Expect(priceRun).ToNot(BeNil())
		Expect(priceRun.Bold).To(BeTrue())
	})

	It("should render inline formatting inside table cells on the correct runs", func() {
		doc := renderDocx(`|===
| *Bold cell* | _Italic cell_

| Normal | ` + "`code`" + `
|===`)

		boldRun := doc.findTableCellRun("Bold cell")
		Expect(boldRun).ToNot(BeNil())
		Expect(boldRun.Bold).To(BeTrue())

		italicRun := doc.findTableCellRun("Italic cell")
		Expect(italicRun).ToNot(BeNil())
		Expect(italicRun.Italic).To(BeTrue())

		codeRun := doc.findTableCellRun("code")
		Expect(codeRun).ToNot(BeNil())
		Expect(codeRun.Monospace).To(BeTrue())

		normalRun := doc.findTableCellRun("Normal")
		Expect(normalRun).ToNot(BeNil())
		Expect(normalRun.Bold).To(BeFalse())
		Expect(normalRun.Italic).To(BeFalse())
	})

	It("should not emit a w:tbl for an empty table", func() {
		doc := renderDocx(`|===
|===`)

		tables := doc.parseTables()
		Expect(tables).To(BeEmpty(), "empty table should not produce a w:tbl element")
	})

	It("should render a multi-column table with correct cell count", func() {
		doc := renderDocx(`|===
| A | B | C | D

| 1 | 2 | 3 | 4
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		for _, row := range tables[0].Rows {
			Expect(row.Cells).To(HaveLen(4))
		}
	})

	It("should default to full-width table (auto-fit to window)", func() {
		doc := renderDocx(`|===
| A | B
| C | D
|===`)
		docXML := doc.documentXML()
		Expect(docXML).To(ContainSubstring(`w:tblW w:w="5000" w:type="pct"`))
	})

	It("should apply auto width from theme", func() {
		doc := renderDocxWithTheme(`|===
| A | B
|===`, `
table:
  width: auto
`)
		docXML := doc.documentXML()
		Expect(docXML).To(ContainSubstring(`w:tblW w:w="0" w:type="auto"`))
	})

	It("should render a nested table inside an asciidoc cell and emit a trailing paragraph", func() {
		// OOXML requires every <w:tc> to end with a <w:p>.
		// When the only cell content is a nested <w:tbl>, the renderer must
		// append a <w:p/> so that Word accepts the document.
		doc := renderDocx(`[cols="1,2"]
|===
| Label
a|
[cols="1,1"]
!===
! A ! B
! 1 ! 2
!===
|===`)

		// The nested table's cell content must be present.
		Expect(doc.findTableCellRun("A")).ToNot(BeNil())
		Expect(doc.findTableCellRun("B")).ToNot(BeNil())

		// Every </w:tbl> that appears inside a <w:tc> must be followed
		// immediately by a <w:p> (or <w:p/>) before </w:tc>.
		docXML := doc.documentXML()
		Expect(docXML).ToNot(ContainSubstring("</w:tbl></w:tc>"),
			"a <w:tc> must not close immediately after </w:tbl> — Word requires a trailing <w:p>")
	})

	It("should render an unordered list inside an asciidoc cell", func() {
		doc := renderDocx(`[cols="1,3"]
|===
| Label
a|
* Alpha
* Beta
* Gamma
|===`)

		// All three items must appear in table cell runs
		Expect(doc.findTableCellRun("Alpha")).ToNot(BeNil())
		Expect(doc.findTableCellRun("Beta")).ToNot(BeNil())
		Expect(doc.findTableCellRun("Gamma")).ToNot(BeNil())

		// The list paragraphs must carry a numbering ID (i.e. be bulleted)
		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		var listParas []parsedParagraph
		for _, row := range tables[0].Rows {
			for _, cell := range row.Cells {
				for _, p := range cell.Paragraphs {
					if p.NumID != "" {
						listParas = append(listParas, p)
					}
				}
			}
		}
		Expect(listParas).To(HaveLen(3), "expected 3 bulleted paragraphs in cell")
	})

	It("should render an ordered list inside an asciidoc cell", func() {
		doc := renderDocx(`[cols="1,3"]
|===
| Steps
a|
. Download
. Install
. Restart
|===`)

		Expect(doc.findTableCellRun("Download")).ToNot(BeNil())
		Expect(doc.findTableCellRun("Install")).ToNot(BeNil())
		Expect(doc.findTableCellRun("Restart")).ToNot(BeNil())

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		var listParas []parsedParagraph
		for _, row := range tables[0].Rows {
			for _, cell := range row.Cells {
				for _, p := range cell.Paragraphs {
					if p.NumID != "" {
						listParas = append(listParas, p)
					}
				}
			}
		}
		Expect(listParas).To(HaveLen(3), "expected 3 numbered paragraphs in cell")
	})

	It("should render a styled ordered list inside an asciidoc cell", func() {
		doc := renderDocx(`[cols="1,3"]
|===
| Options
a|
[loweralpha]
. Enable logging
. Configure endpoint
. Verify connectivity
|===`)

		Expect(doc.findTableCellRun("Enable logging")).ToNot(BeNil())
		Expect(doc.findTableCellRun("Configure endpoint")).ToNot(BeNil())
		Expect(doc.findTableCellRun("Verify connectivity")).ToNot(BeNil())

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		var listParas []parsedParagraph
		for _, row := range tables[0].Rows {
			for _, cell := range row.Cells {
				for _, p := range cell.Paragraphs {
					if p.NumID != "" {
						listParas = append(listParas, p)
					}
				}
			}
		}
		Expect(listParas).To(HaveLen(3), "expected 3 lettered paragraphs in cell")
	})

	It("should apply percentage width from theme", func() {
		doc := renderDocxWithTheme(`|===
| A | B
|===`, `
table:
  width: "80%"
`)
		docXML := doc.documentXML()
		// 80% = 80 * 50 = 4000 fiftieths-of-a-percent
		Expect(docXML).To(ContainSubstring(`w:tblW w:w="4000" w:type="pct"`))
	})
})
