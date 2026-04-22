package docx

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderTable(t *types.Table) error {
	rows := tableRows(t)
	colCount := tableColumnCount(t, rows)
	if colCount == 0 || len(rows) == 0 {
		return nil
	}
	if title := t.Attributes.GetAsStringWithDefault(types.AttrTitle, ""); title != "" {
		number := r.ctx.GetAndIncrementTableCounter()
		captionPrefix := t.Attributes.GetAsStringWithDefault(types.AttrCaption, "")
		if captionPrefix == "" {
			captionPrefix = r.ctx.attributes.GetAsStringWithDefault(types.AttrTableCaption, "Table")
			if captionPrefix != "" {
				captionPrefix += " " + fmt.Sprint(number) + ". "
			}
		}
		if err := r.renderTextParagraph(captionPrefix+title, paragraphOptions{style: "Caption"}); err != nil {
			return err
		}
	}
	r.writer.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:left w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:right w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="auto"/></w:tblBorders></w:tblPr><w:tblGrid>`)
	for i := 0; i < colCount; i++ {
		r.writer.WriteString(`<w:gridCol w:w="`)
		r.writer.WriteString(fmt.Sprint(9000 / colCount))
		r.writer.WriteString(`"/>`)
	}
	r.writer.WriteString(`</w:tblGrid>`)
	for i, row := range rows {
		bold := (t.Header != nil && i == 0) || (t.Footer != nil && i == len(rows)-1)
		if err := r.renderTableRow(row, colCount, bold); err != nil {
			return err
		}
	}
	r.writer.WriteString(`</w:tbl>`)
	return nil
}

func (r *docxRenderer) renderTableRow(row *types.TableRow, colCount int, bold bool) error {
	r.writer.WriteString("<w:tr>")
	for i := 0; i < colCount; i++ {
		var cell *types.TableCell
		if i < len(row.Cells) {
			cell = row.Cells[i]
		}
		if err := r.renderTableCell(cell, bold); err != nil {
			return err
		}
	}
	r.writer.WriteString("</w:tr>")
	return nil
}

func (r *docxRenderer) renderTableCell(cell *types.TableCell, bold bool) error {
	r.writer.WriteString(`<w:tc><w:tcPr><w:tcW w:w="0" w:type="auto"/></w:tcPr>`)
	if cell == nil || len(cell.Elements) == 0 {
		r.writer.WriteString(`<w:p/>`)
		r.writer.WriteString(`</w:tc>`)
		return nil
	}
	old := r.writer
	cellWriter := &strings.Builder{}
	r.writer = cellWriter
	for _, elem := range cell.Elements {
		switch e := elem.(type) {
		case *types.Paragraph:
			para := r.startParagraph(paragraphOptions{})
			if err := r.renderInlineElements(para, e.Elements, runStyle{bold: bold}); err != nil {
				r.writer = old
				return err
			}
			r.endParagraph(para)
		default:
			if err := r.renderElement(elem); err != nil {
				r.writer = old
				return err
			}
		}
	}
	r.writer = old
	if cellWriter.Len() == 0 {
		r.writer.WriteString(`<w:p/>`)
	} else {
		r.writer.WriteString(cellWriter.String())
	}
	r.writer.WriteString(`</w:tc>`)
	return nil
}

func tableRows(t *types.Table) []*types.TableRow {
	rows := make([]*types.TableRow, 0, len(t.Rows)+2)
	if t.Header != nil {
		rows = append(rows, t.Header)
	}
	rows = append(rows, t.Rows...)
	if t.Footer != nil {
		rows = append(rows, t.Footer)
	}
	return rows
}

func tableColumnCount(t *types.Table, rows []*types.TableRow) int {
	cols, err := t.Columns()
	max := 0
	if err == nil {
		max = len(cols)
	}
	for _, row := range rows {
		if len(row.Cells) > max {
			max = len(row.Cells)
		}
	}
	return max
}
