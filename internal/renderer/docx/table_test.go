package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tables", func() {

	It("should render a simple table with header and data rows", func() {
		doc := renderDocx(`|===
| Header 1 | Header 2

| Cell 1 | Cell 2
| Cell 3 | Cell 4
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))

		table := tables[0]
		Expect(table.Rows).To(HaveLen(3)) // 1 header + 2 data rows

		// Header cells should be bold
		headerRun := doc.findTableCellRun("Header 1")
		Expect(headerRun).ToNot(BeNil())
		Expect(headerRun.Bold).To(BeTrue())

		// Data cells should NOT be bold
		dataRun := doc.findTableCellRun("Cell 1")
		Expect(dataRun).ToNot(BeNil())
		Expect(dataRun.Bold).To(BeFalse())
	})

	It("should render table without header with no bold cells", func() {
		doc := renderDocx(`|===
| A | B
| C | D
|===`)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))

		// No cells should be bold since there is no header
		for _, text := range []string{"A", "B", "C", "D"} {
			r := doc.findTableCellRun(text)
			Expect(r).ToNot(BeNil())
			Expect(r.Bold).To(BeFalse(), "cell %q should not be bold in headerless table", text)
		}
	})

	It("should render table captions with Caption style and numbering", func() {
		doc := renderDocx(`.Pricing
|===
| A | B
|===`)

		caption := doc.findParagraph("Table 1. Pricing")
		Expect(caption).ToNot(BeNil())
		Expect(caption.Style).To(Equal("Caption"))
	})
})
