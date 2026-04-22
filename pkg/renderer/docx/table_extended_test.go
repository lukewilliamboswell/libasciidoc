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
})
