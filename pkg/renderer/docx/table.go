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

	theme := r.ctx.theme.Table
	borderSz := itoa(ptToEighths(theme.BorderWidth))
	borderColor := theme.BorderColor
	outerBorder := `w:val="single" w:sz="` + borderSz + `" w:space="0" w:color="` + xmlAttr(borderColor) + `"`

	// Grid lines may differ from outer borders
	gridSz := borderSz
	gridColor := borderColor
	if theme.GridColor != "" {
		gridColor = theme.GridColor
	}
	if theme.GridWidth > 0 {
		gridSz = itoa(ptToEighths(theme.GridWidth))
	}
	innerBorder := `w:val="single" w:sz="` + gridSz + `" w:space="0" w:color="` + xmlAttr(gridColor) + `"`

	r.writer.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders>` +
		`<w:top ` + outerBorder + `/><w:left ` + outerBorder + `/><w:bottom ` + outerBorder + `/>` +
		`<w:right ` + outerBorder + `/><w:insideH ` + innerBorder + `/><w:insideV ` + innerBorder + `/>` +
		`</w:tblBorders>`)

	// Cell padding (table-level cell margins)
	if theme.CellPadding > 0 {
		pad := itoa(ptToTwips(theme.CellPadding))
		r.writer.WriteString(`<w:tblCellMar>`)
		r.writer.WriteString(`<w:top w:w="` + pad + `" w:type="dxa"/>`)
		r.writer.WriteString(`<w:left w:w="` + pad + `" w:type="dxa"/>`)
		r.writer.WriteString(`<w:bottom w:w="` + pad + `" w:type="dxa"/>`)
		r.writer.WriteString(`<w:right w:w="` + pad + `" w:type="dxa"/>`)
		r.writer.WriteString(`</w:tblCellMar>`)
	}

	r.writer.WriteString(`</w:tblPr><w:tblGrid>`)
	for i := 0; i < colCount; i++ {
		r.writer.WriteString(`<w:gridCol w:w="`)
		r.writer.WriteString(fmt.Sprint(tableGridWidthTwips / colCount))
		r.writer.WriteString(`"/>`)
	}
	r.writer.WriteString(`</w:tblGrid>`)
	for i, row := range rows {
		isHeader := t.Header != nil && i == 0
		isFooter := t.Footer != nil && i == len(rows)-1

		// Determine bold from theme or structure
		bold := isHeader || isFooter
		if isHeader && theme.HeadFontStyle != "" {
			hBold, _ := fontStyleBoldItalic(theme.HeadFontStyle)
			bold = hBold
		}
		if isFooter && theme.FootFontStyle != "" {
			fBold, _ := fontStyleBoldItalic(theme.FootFontStyle)
			bold = fBold
		}

		// Determine italic
		italic := false
		if isHeader && theme.HeadFontStyle != "" {
			_, italic = fontStyleBoldItalic(theme.HeadFontStyle)
		}
		if isFooter && theme.FootFontStyle != "" {
			_, italic = fontStyleBoldItalic(theme.FootFontStyle)
		}

		// Determine background color
		bgColor := ""
		if isHeader && theme.HeadBgColor != "" {
			bgColor = theme.HeadBgColor
		} else if isFooter && theme.FootBgColor != "" {
			bgColor = theme.FootBgColor
		} else if !isHeader && !isFooter && theme.StripeBgColor != "" {
			// Stripe: apply to even data rows (0-indexed after header)
			dataIdx := i
			if t.Header != nil {
				dataIdx = i - 1
			}
			if dataIdx%2 == 1 {
				bgColor = theme.StripeBgColor
			}
		}

		if err := r.renderTableRow(row, colCount, bold, italic, bgColor); err != nil {
			return err
		}
	}
	r.writer.WriteString(`</w:tbl>`)
	return nil
}

func (r *docxRenderer) renderTableRow(row *types.TableRow, colCount int, bold, italic bool, bgColor string) error {
	r.writer.WriteString("<w:tr>")
	for i := 0; i < colCount; i++ {
		var cell *types.TableCell
		if i < len(row.Cells) {
			cell = row.Cells[i]
		}
		if err := r.renderTableCell(cell, bold, italic, bgColor); err != nil {
			return err
		}
	}
	r.writer.WriteString("</w:tr>")
	return nil
}

func (r *docxRenderer) renderTableCell(cell *types.TableCell, bold, italic bool, bgColor string) error {
	r.writer.WriteString(`<w:tc><w:tcPr><w:tcW w:w="0" w:type="auto"/>`)
	if bgColor != "" {
		r.writer.WriteString(`<w:shd w:val="clear" w:color="auto" w:fill="`)
		r.writer.WriteString(xmlAttr(bgColor))
		r.writer.WriteString(`"/>`)
	}
	r.writer.WriteString(`</w:tcPr>`)
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
			if err := r.renderInlineElements(para, e.Elements, runStyle{bold: bold, italic: italic}); err != nil {
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
