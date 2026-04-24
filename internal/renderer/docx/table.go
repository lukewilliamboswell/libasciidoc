package docx

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
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
	outer := borderLine{sizePt: theme.BorderWidth, space: 0, color: theme.BorderColor}
	inner := borderLine{sizePt: theme.BorderWidth, space: 0, color: theme.BorderColor}
	if theme.GridColor != "" {
		inner.color = theme.GridColor
	}
	if theme.GridWidth > 0 {
		inner.sizePt = theme.GridWidth
	}

	tblW, tblWType := tableWidthAttrs(theme.Width)
	props := tableProps{
		width: tableWidth{w: tblW, wType: tblWType},
		borders: tableBorders{
			top: outer, left: outer, bottom: outer, right: outer,
			insideH: inner, insideV: inner,
		},
	}
	if theme.CellPadding > 0 {
		pad := ptToTwips(theme.CellPadding)
		props.cellMar = &tableCellMargins{top: pad, left: pad, bottom: pad, right: pad}
	}

	cols, _ := t.Columns()
	textWidth := r.doc.textWidthTwips()
	var colWidths []int
	if len(cols) > 0 {
		totalWeight := 0
		for _, c := range cols {
			totalWeight += c.Weight
		}
		if totalWeight <= 0 {
			totalWeight = len(cols)
		}
		for _, c := range cols {
			w := c.Weight * textWidth / totalWeight
			if w <= 0 {
				w = textWidth / len(cols)
			}
			colWidths = append(colWidths, w)
		}
	} else {
		colW := textWidth / colCount
		colWidths = make([]int, colCount)
		for i := range colWidths {
			colWidths[i] = colW
		}
	}

	r.writer.WriteString("<w:tbl>")
	props.xml(r.writer)
	tableGrid{colWidths: colWidths}.xml(r.writer)
	for i, row := range rows {
		bold, italic, bgColor := r.tableRowStyle(t, theme, i, len(rows))
		isHeader := t.Header != nil && i == 0
		if err := r.renderTableRow(row, colCount, bold, italic, bgColor, isHeader); err != nil {
			return err
		}
	}
	r.writer.WriteString(`</w:tbl>`)
	return nil
}

func (r *docxRenderer) tableRowStyle(t *types.Table, theme TableTheme, i int, totalRows int) (bold, italic bool, bgColor string) {
	isHeader := t.Header != nil && i == 0
	isFooter := t.Footer != nil && i == totalRows-1

	// Determine bold from theme or structure
	bold = isHeader || isFooter
	if isHeader && theme.HeadFontStyle != "" {
		hBold, _ := fontStyleBoldItalic(theme.HeadFontStyle)
		bold = hBold
	}
	if isFooter && theme.FootFontStyle != "" {
		fBold, _ := fontStyleBoldItalic(theme.FootFontStyle)
		bold = fBold
	}

	// Determine italic
	if isHeader && theme.HeadFontStyle != "" {
		_, italic = fontStyleBoldItalic(theme.HeadFontStyle)
	}
	if isFooter && theme.FootFontStyle != "" {
		_, italic = fontStyleBoldItalic(theme.FootFontStyle)
	}

	// Determine background color
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
	return
}

func (r *docxRenderer) renderTableRow(row *types.TableRow, colCount int, bold, italic bool, bgColor string, isHeader bool) error {
	r.writer.WriteString("<w:tr>")
	if isHeader {
		r.writer.WriteString(`<w:trPr><w:tblHeader/></w:trPr>`)
	}
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
	r.writer.WriteString("<w:tc>")
	tableCellProps{widthW: "0", widthWType: "auto", bgColor: bgColor}.xml(r.writer)
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
		case *types.List:
			if err := r.renderList(e); err != nil {
				r.writer = old
				return err
			}
		default:
			if err := r.renderElement(elem); err != nil {
				r.writer = old
				return err
			}
		}
	}
	r.writer = old
	content := cellWriter.String()
	if content == "" {
		r.writer.WriteString(`<w:p/>`)
	} else {
		r.writer.WriteString(content)
		// OOXML requires every <w:tc> to end with a <w:p>.
		// A nested table does not include one, so append it.
		if strings.HasSuffix(content, "</w:tbl>") {
			r.writer.WriteString(`<w:p/>`)
		}
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
