package docx_test

import (
	"strconv"

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

	It("should use proportional column widths from cols attribute", func() {
		// [cols="1,3"] means col1 = 1/4 of text width, col2 = 3/4 of text width.
		// Default A4 page (11906 twips) with 20mm margins on each side (1134 twips each)
		// text width = 11906 - 1134 - 1134 = 9638 twips
		// col1 = 9638 * 1/4 = 2409, col2 = 9638 * 3/4 = 7228
		doc := renderDocx(`[cols="1,3"]
|===
| Label | Value

| Name | Alice
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		Expect(tables[0].GridColWidths).To(HaveLen(2))

		col1 := tables[0].GridColWidths[0]
		col2 := tables[0].GridColWidths[1]

		// col2 should be approximately 3× col1
		ratio := float64(col2) / float64(col1)
		Expect(ratio).To(BeNumerically("~", 3.0, 0.1), "col2 should be ~3× col1, got col1=%d col2=%d", col1, col2)

		// Total should equal text width
		textWidth := col1 + col2
		Expect(textWidth).To(BeNumerically("~", 9638, 10), "total width should equal text width (~9638 twips), got %d", textWidth)
	})

	It("should use equal column widths when no cols attribute is given", func() {
		doc := renderDocx(`|===
| A | B | C
| 1 | 2 | 3
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		Expect(tables[0].GridColWidths).To(HaveLen(3))

		w0 := tables[0].GridColWidths[0]
		w1 := tables[0].GridColWidths[1]
		w2 := tables[0].GridColWidths[2]
		// The last column absorbs the integer-division residual so the totals match.
		Expect(w0).To(Equal(w1), "first two columns should be equal width")
		Expect(w2).To(BeNumerically("~", w1, 3), "last column within rounding residual")
		Expect(w0 + w1 + w2).To(Equal(9638), "total equals available content width (A4 - 2x20mm)")
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

	Describe("column widths emitted as dxa twips", func() {
		// A4 portrait: 11906 twips wide. Default 20mm margins on each side ≈ 1134 twips.
		// Available content width = 11906 - 1134 - 1134 = 9638 twips.
		const a4Available = 9638

		It("emits two equal gridCols and matching tcW values when no cols= is given", func() {
			doc := renderDocx(`|===
| A | B
| 1 | 2
|===`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			t := tables[0]
			Expect(t.GridColWidths).To(HaveLen(2))
			Expect(t.GridColWidths[0]).To(Equal(a4Available / 2))
			Expect(t.GridColWidths[0] + t.GridColWidths[1]).To(Equal(a4Available))
			for _, row := range t.Rows {
				Expect(row.Cells).To(HaveLen(2))
				for i, cell := range row.Cells {
					Expect(cell.WidthType).To(Equal("dxa"))
					Expect(cell.Width).To(Equal(strconv.Itoa(t.GridColWidths[i])))
				}
			}
		})

		It("honours [cols=\"2,5\"] ratio across grid and cell widths", func() {
			doc := renderDocx(`[cols="2,5"]
|===
| A | B
| 1 | 2
|===`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			t := tables[0]
			Expect(t.GridColWidths).To(HaveLen(2))
			Expect(t.GridColWidths[0]).To(Equal(2 * a4Available / 7))
			Expect(t.GridColWidths[0] + t.GridColWidths[1]).To(Equal(a4Available))
			ratio := float64(t.GridColWidths[1]) / float64(t.GridColWidths[0])
			Expect(ratio).To(BeNumerically("~", 2.5, 0.01))
			for _, row := range t.Rows {
				Expect(row.Cells).To(HaveLen(2))
				for i, cell := range row.Cells {
					Expect(cell.WidthType).To(Equal("dxa"))
					Expect(cell.Width).To(Equal(strconv.Itoa(t.GridColWidths[i])))
				}
			}
		})

		It("honours [cols=\"1,1,1\"] as three equal columns", func() {
			doc := renderDocx(`[cols="1,1,1"]
|===
| A | B | C
| 1 | 2 | 3
|===`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			t := tables[0]
			Expect(t.GridColWidths).To(HaveLen(3))
			Expect(t.GridColWidths[0]).To(Equal(t.GridColWidths[1]))
			Expect(t.GridColWidths[0] + t.GridColWidths[1] + t.GridColWidths[2]).To(Equal(a4Available))
		})

		It("recomputes widths against the theme page size (Letter)", func() {
			// Letter portrait: 12240 twips. With default 20mm margins (1134 each), available = 12240 - 2268 = 9972.
			doc := renderDocxWithTheme(`[cols="1,1"]
|===
| A | B
|===`, `
page:
  size: Letter
`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			t := tables[0]
			Expect(t.GridColWidths).To(HaveLen(2))
			letterAvailable := 12240 - 2*1134
			Expect(t.GridColWidths[0] + t.GridColWidths[1]).To(Equal(letterAvailable))
			Expect(t.GridColWidths[0]).To(Equal(letterAvailable / 2))
		})

		It("recomputes widths against the theme page size (A4 explicit)", func() {
			doc := renderDocxWithTheme(`[cols="1,1"]
|===
| A | B
|===`, `
page:
  size: A4
`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			Expect(tables[0].GridColWidths[0] + tables[0].GridColWidths[1]).To(Equal(a4Available))
		})

		It("emits a dxa cell width spanning the merged columns when colspan is used", func() {
			doc := renderDocx(`[cols="1,1,1"]
|===
| A | B | C
2+| spans-two | last
|===`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			t := tables[0]
			Expect(t.GridColWidths).To(HaveLen(3))
			Expect(t.Rows).To(HaveLen(2))

			spanRow := t.Rows[1]
			Expect(spanRow.Cells).To(HaveLen(2))
			Expect(spanRow.Cells[0].GridSpan).To(Equal("2"))
			Expect(spanRow.Cells[0].WidthType).To(Equal("dxa"))
			expectedSpanned := t.GridColWidths[0] + t.GridColWidths[1]
			Expect(spanRow.Cells[0].Width).To(Equal(strconv.Itoa(expectedSpanned)))
			Expect(spanRow.Cells[1].Width).To(Equal(strconv.Itoa(t.GridColWidths[2])))
		})

		It("ensures the gridCol total exactly equals available content width", func() {
			// [cols="2,3,5"] across 9638 twips: 2/10=1927, 3/10=2891, last absorbs residual.
			doc := renderDocx(`[cols="2,3,5"]
|===
| A | B | C
|===`)
			tables := doc.parseTables()
			Expect(tables).To(HaveLen(1))
			widths := tables[0].GridColWidths
			Expect(widths).To(HaveLen(3))
			sum := 0
			for _, w := range widths {
				sum += w
			}
			Expect(sum).To(Equal(a4Available), "tblGrid widths must sum to available content width exactly")
		})
	})
})
