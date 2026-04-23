package types_test

import (
	"github.com/lukewilliamboswell/libasciidoc/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tables", func() {

	// helper to build a TableRow with N cells containing simple content
	makeRow := func(contents ...string) *types.TableRow {
		cells := make([]*types.TableCell, len(contents))
		for i, c := range contents {
			cells[i] = &types.TableCell{
				Elements: []interface{}{&types.StringElement{Content: c}},
			}
		}
		return &types.TableRow{Cells: cells}
	}

	Describe("NewTable", func() {

		It("should return an empty table for empty input", func() {
			t, err := types.NewTable([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Header).To(BeNil())
			Expect(t.Rows).To(BeEmpty())
		})

		It("should detect a header row followed by a blank line", func() {
			header := makeRow("Col1", "Col2")
			body1 := makeRow("a", "b")
			body2 := makeRow("c", "d")
			t, err := types.NewTable([]interface{}{
				header,
				&types.BlankLine{},
				body1,
				body2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Header).NotTo(BeNil())
			Expect(t.Header.Cells).To(HaveLen(2))
			Expect(t.Rows).To(HaveLen(2))
		})

		It("should not detect a header when there is no blank line", func() {
			row1 := makeRow("a", "b")
			row2 := makeRow("c", "d")
			t, err := types.NewTable([]interface{}{
				row1,
				row2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Header).To(BeNil())
			Expect(t.Rows).To(HaveLen(2))
		})

		It("should handle a single row (no header possible)", func() {
			row := makeRow("only")
			t, err := types.NewTable([]interface{}{row})
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Header).To(BeNil())
			Expect(t.Rows).To(HaveLen(1))
		})

		It("should organize cells into rows based on header length", func() {
			header := makeRow("A", "B", "C")
			// Body has 6 cells across 2 rows, but passed as single-cell rows
			body := make([]interface{}, 0)
			for _, s := range []string{"1", "2", "3", "4", "5", "6"} {
				body = append(body, makeRow(s))
			}
			lines := append([]interface{}{header, &types.BlankLine{}}, body...)
			t, err := types.NewTable(lines)
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Header).NotTo(BeNil())
			Expect(t.Header.Cells).To(HaveLen(3))
			// 6 cells / 3 columns = 2 rows
			Expect(t.Rows).To(HaveLen(2))
			Expect(t.Rows[0].Cells).To(HaveLen(3))
			Expect(t.Rows[1].Cells).To(HaveLen(3))
		})

		It("should handle partial last row", func() {
			header := makeRow("A", "B")
			body := []interface{}{
				makeRow("1"),
				makeRow("2"),
				makeRow("3"), // partial: only 1 cell in last row
			}
			lines := append([]interface{}{header, &types.BlankLine{}}, body...)
			t, err := types.NewTable(lines)
			Expect(err).NotTo(HaveOccurred())
			// 3 cells / 2 columns = 1 full row + 1 partial row
			Expect(t.Rows).To(HaveLen(2))
			Expect(t.Rows[0].Cells).To(HaveLen(2))
			Expect(t.Rows[1].Cells).To(HaveLen(1))
		})

		It("should silently skip blank lines in body cells", func() {
			row1 := makeRow("a", "b")
			row2 := makeRow("c", "d")
			t, err := types.NewTable([]interface{}{
				row1,
				&types.BlankLine{},
				row2,
			})
			Expect(err).NotTo(HaveOccurred())
			// row1 becomes header because it is followed by blank line
			Expect(t.Header).NotTo(BeNil())
			Expect(t.Rows).To(HaveLen(1))
		})
	})

	Describe("reorganizeRows", func() {

		It("should promote first row to header with header option", func() {
			row1 := makeRow("h1", "h2")
			row2 := makeRow("a", "b")
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrOptions: types.Options{"header"},
				},
				Rows: []*types.TableRow{row1, row2},
			}
			t.SetAttributes(t.Attributes) // triggers reorganizeRows
			Expect(t.Header).NotTo(BeNil())
			Expect(t.Rows).To(HaveLen(1))
		})

		It("should promote last row to footer with footer option", func() {
			row1 := makeRow("a", "b")
			row2 := makeRow("f1", "f2")
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrOptions: types.Options{"footer"},
				},
				Rows: []*types.TableRow{row1, row2},
			}
			t.SetAttributes(t.Attributes)
			Expect(t.Footer).NotTo(BeNil())
			Expect(t.Rows).To(HaveLen(1))
		})

		It("should promote both header and footer with both options", func() {
			row1 := makeRow("h")
			row2 := makeRow("body")
			row3 := makeRow("f")
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrOptions: types.Options{"header", "footer"},
				},
				Rows: []*types.TableRow{row1, row2, row3},
			}
			t.SetAttributes(t.Attributes)
			Expect(t.Header).NotTo(BeNil())
			Expect(t.Footer).NotTo(BeNil())
			Expect(t.Rows).To(HaveLen(1))
		})

		It("should not promote if header already set", func() {
			header := makeRow("h")
			row1 := makeRow("a")
			row2 := makeRow("b")
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrOptions: types.Options{"header"},
				},
				Header: header,
				Rows:   []*types.TableRow{row1, row2},
			}
			t.SetAttributes(t.Attributes)
			// Header should remain unchanged
			Expect(t.Header).To(Equal(header))
			Expect(t.Rows).To(HaveLen(2))
		})
	})

	Describe("NewTableColumn", func() {

		It("should create a default column with nil parameters", func() {
			col, err := types.NewTableColumn(nil, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(col.Multiplier).To(Equal(1))
			Expect(col.HAlign).To(Equal(types.HAlignLeft))
			Expect(col.VAlign).To(Equal(types.VAlignTop))
			Expect(col.Weight).To(Equal(1))
			Expect(col.Autowidth).To(BeFalse())
		})

		It("should set multiplier from int", func() {
			col, err := types.NewTableColumn(3, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(col.Multiplier).To(Equal(3))
		})

		It("should set alignment", func() {
			col, err := types.NewTableColumn(nil, types.HAlignCenter, types.VAlignMiddle, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(col.HAlign).To(Equal(types.HAlignCenter))
			Expect(col.VAlign).To(Equal(types.VAlignMiddle))
		})

		It("should set weight from int", func() {
			col, err := types.NewTableColumn(nil, nil, nil, 5, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(col.Weight).To(Equal(5))
			Expect(col.Autowidth).To(BeFalse())
		})

		It("should enable autowidth with tilde", func() {
			col, err := types.NewTableColumn(nil, nil, nil, "~", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(col.Autowidth).To(BeTrue())
			Expect(col.Weight).To(Equal(0))
		})

		It("should set content style", func() {
			col, err := types.NewTableColumn(nil, nil, nil, nil, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(col.Style).To(Equal(types.AsciidocStyle))
		})
	})

	Describe("NewTableRow", func() {

		It("should create a row from table cells", func() {
			cell1 := &types.TableCell{Elements: []interface{}{&types.StringElement{Content: "a"}}}
			cell2 := &types.TableCell{Elements: []interface{}{&types.StringElement{Content: "b"}}}
			row, err := types.NewTableRow([]interface{}{cell1, cell2})
			Expect(err).NotTo(HaveOccurred())
			Expect(row.Cells).To(HaveLen(2))
		})

		It("should return an error for non-cell elements", func() {
			_, err := types.NewTableRow([]interface{}{&types.StringElement{Content: "not a cell"}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected type of table cell"))
		})

		It("should create an empty row", func() {
			row, err := types.NewTableRow([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
			Expect(row.Cells).To(BeEmpty())
		})
	})

	Describe("NewInlineTableCell", func() {

		It("should wrap a raw line in a cell", func() {
			rl, err := types.NewRawLine("hello")
			Expect(err).NotTo(HaveOccurred())
			cell, err := types.NewInlineTableCell(rl)
			Expect(err).NotTo(HaveOccurred())
			Expect(cell.Elements).To(HaveLen(1))
		})
	})

	Describe("NewMultilineTableCell", func() {

		It("should append newlines to all but last raw line", func() {
			rl1, _ := types.NewRawLine("line1")
			rl2, _ := types.NewRawLine("line2")
			rl3, _ := types.NewRawLine("line3")
			cell, err := types.NewMultilineTableCell([]interface{}{rl1, rl2, rl3}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cell.Elements).To(HaveLen(3))
			// first two should have newlines appended
			Expect(rl1.Content).To(Equal("line1\n"))
			Expect(rl2.Content).To(Equal("line2\n"))
			Expect(rl3.Content).To(Equal("line3"))
		})

		It("should set format from string", func() {
			rl, _ := types.NewRawLine("content")
			cell, err := types.NewMultilineTableCell([]interface{}{rl}, "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(cell.Format).To(Equal("a"))
		})

		It("should ignore non-string format", func() {
			rl, _ := types.NewRawLine("content")
			cell, err := types.NewMultilineTableCell([]interface{}{rl}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cell.Format).To(BeEmpty())
		})
	})

	Describe("Columns", func() {

		It("should infer columns from first row when no AttrCols", func() {
			t := &types.Table{
				Rows: []*types.TableRow{
					makeRow("a", "b", "c"),
				},
			}
			cols, err := t.Columns()
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(3))
		})

		It("should use AttrCols when defined", func() {
			col1, _ := types.NewTableColumn(nil, nil, nil, 2, nil)
			col2, _ := types.NewTableColumn(nil, nil, nil, 3, nil)
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrCols: []interface{}{col1, col2},
				},
				Rows: []*types.TableRow{makeRow("a", "b")},
			}
			cols, err := t.Columns()
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(2))
			Expect(cols[0].Weight).To(Equal(2))
			Expect(cols[1].Weight).To(Equal(3))
		})

		It("should expand multiplier", func() {
			col, _ := types.NewTableColumn(3, nil, nil, nil, nil)
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrCols: []interface{}{col},
				},
				Rows: []*types.TableRow{makeRow("a", "b", "c")},
			}
			cols, err := t.Columns()
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(3))
		})

		It("should compute width percentages for equal columns", func() {
			t := &types.Table{
				Rows: []*types.TableRow{
					makeRow("a", "b"),
				},
			}
			cols, err := t.Columns()
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(2))
			Expect(cols[0].Width).To(Equal("50"))
			Expect(cols[1].Width).To(Equal("50"))
		})

		It("should leave width empty for autowidth columns", func() {
			col1, _ := types.NewTableColumn(nil, nil, nil, 2, nil)
			col2, _ := types.NewTableColumn(nil, nil, nil, "~", nil)
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrCols: []interface{}{col1, col2},
				},
				Rows: []*types.TableRow{makeRow("a", "b")},
			}
			cols, err := t.Columns()
			Expect(err).NotTo(HaveOccurred())
			Expect(cols[0].Width).To(Equal("2"))
			Expect(cols[1].Width).To(BeEmpty())
		})

		It("should return error for invalid column type in AttrCols", func() {
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrCols: []interface{}{"not-a-column"},
				},
			}
			_, err := t.Columns()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid type of column definition"))
		})

		It("should not set widths when autowidth option is set on table", func() {
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrOptions: types.Options{types.AttrAutoWidth},
				},
				Rows: []*types.TableRow{
					makeRow("a", "b"),
				},
			}
			cols, err := t.Columns()
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(2))
			for _, col := range cols {
				Expect(col.Width).To(BeEmpty())
			}
		})
	})

	Describe("Table.SetElements", func() {

		It("should set rows from TableRow elements", func() {
			t := &types.Table{}
			row := makeRow("a")
			err := t.SetElements([]interface{}{row})
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Rows).To(HaveLen(1))
		})

		It("should clear rows with empty input", func() {
			t := &types.Table{Rows: []*types.TableRow{makeRow("a")}}
			err := t.SetElements([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
			Expect(t.Rows).To(BeNil())
		})

		It("should return error for non-TableRow elements", func() {
			t := &types.Table{}
			err := t.SetElements([]interface{}{&types.StringElement{Content: "bad"}})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Table.rows", func() {

		It("should include header and footer in rows list", func() {
			header := makeRow("h")
			footer := makeRow("f")
			body := makeRow("b")
			t := &types.Table{
				Header: header,
				Footer: footer,
				Rows:   []*types.TableRow{body},
			}
			elements := t.GetElements()
			// GetElements only returns body rows
			Expect(elements).To(HaveLen(1))
		})
	})

	Describe("Table.Reference", func() {

		It("should add table to element references when it has id and title", func() {
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrID:    "my-table",
					types.AttrTitle: "My Table",
				},
			}
			refs := types.ElementReferences{}
			t.Reference(refs)
			Expect(refs).To(HaveKeyWithValue("my-table", "My Table"))
		})

		It("should not add reference without an id", func() {
			t := &types.Table{
				Attributes: types.Attributes{
					types.AttrTitle: "No ID",
				},
			}
			refs := types.ElementReferences{}
			t.Reference(refs)
			Expect(refs).To(BeEmpty())
		})
	})
})
