package docx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

// cellFormat holds the parsed cell format specifier from AsciiDoc syntax.
// Format: [colspan[.rowspan]+][halign[.valign]][style]
type cellFormat struct {
	ColSpan int    // default 1; from "2+" prefix
	RowSpan int    // default 1; from ".2+" prefix
	HAlign  string // "center", "right", "left", or "" (default)
}

// parseCellFormat parses an AsciiDoc cell format string (e.g. "2+", ".2+", "^", "2+^").
func parseCellFormat(format string) cellFormat {
	cf := cellFormat{ColSpan: 1, RowSpan: 1}
	if format == "" {
		return cf
	}
	rest := format
	if i := strings.IndexByte(format, '+'); i >= 0 {
		span := format[:i]
		rest = format[i+1:]
		if dot := strings.IndexByte(span, '.'); dot >= 0 {
			if dot > 0 {
				if n, err := strconv.Atoi(span[:dot]); err == nil && n > 1 {
					cf.ColSpan = n
				}
			}
			if dot < len(span)-1 {
				if n, err := strconv.Atoi(span[dot+1:]); err == nil && n > 1 {
					cf.RowSpan = n
				}
			}
		} else if span != "" {
			if n, err := strconv.Atoi(span); err == nil && n > 1 {
				cf.ColSpan = n
			}
		}
	}
	if len(rest) > 0 {
		switch rest[0] {
		case '<':
			cf.HAlign = "left"
		case '>':
			cf.HAlign = "right"
		case '^':
			cf.HAlign = "center"
		}
	}
	return cf
}

func (r *docxRenderer) renderTable(t *types.Table) error {
	rows := tableRows(t)
	colCount := tableColumnCount(t, rows)
	if colCount == 0 || len(rows) == 0 {
		return nil
	}

	// Redistribute cells into rows when spans are present.
	// The AsciiDoc parser may group all cells from multiple source lines into
	// a single raw row; redistributeCells corrects this using span-aware placement.
	rows, colCount = redistributeCells(rows, colCount)

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
	weights := make([]int, colCount)
	useWeights := len(cols) > 0 && len(cols) == colCount
	if useWeights {
		for i, c := range cols {
			weights[i] = c.Weight
		}
	} else {
		for i := range weights {
			weights[i] = 1
		}
	}
	colWidths := distributeWidths(textWidth, weights)

	r.writer.WriteString("<w:tbl>")
	props.xml(r.writer)
	tableGrid{colWidths: colWidths}.xml(r.writer)
	vmerge := make([]int, colCount) // remaining rows for vertical merge at each column
	for i, row := range rows {
		bold, italic, bgColor := r.tableRowStyle(t, theme, i, len(rows))
		isHeader := t.Header != nil && i == 0
		if err := r.renderTableRow(row, colCount, bold, italic, bgColor, isHeader, vmerge, colWidths); err != nil {
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

func (r *docxRenderer) renderTableRow(row *types.TableRow, colCount int, bold, italic bool, bgColor string, isHeader bool, vmerge []int, colWidths []int) error {
	r.writer.WriteString("<w:tr>")
	if isHeader {
		r.writer.WriteString(`<w:trPr><w:tblHeader/></w:trPr>`)
	}
	cellIdx := 0
	for col := 0; col < colCount; col++ {
		if vmerge[col] > 0 {
			// Emit vertical merge continuation cell.
			r.writer.WriteString(`<w:tc>`)
			tableCellProps{widthW: strconv.Itoa(colWidths[col]), widthWType: "dxa", bgColor: bgColor, vMerge: "continue"}.xml(r.writer)
			r.writer.WriteString(`<w:p/></w:tc>`)
			vmerge[col]--
			continue
		}
		var cell *types.TableCell
		if cellIdx < len(row.Cells) {
			cell = row.Cells[cellIdx]
		}
		cellIdx++
		var cf cellFormat
		if cell != nil {
			cf = parseCellFormat(cell.Format)
		}
		if cf.RowSpan > 1 {
			vmerge[col] = cf.RowSpan - 1
		}
		cellW := 0
		spanEnd := col + cf.ColSpan
		if spanEnd > colCount {
			spanEnd = colCount
		}
		for c := col; c < spanEnd; c++ {
			cellW += colWidths[c]
		}
		if err := r.renderTableCell(cell, bold, italic, bgColor, cf, cellW); err != nil {
			return err
		}
		// Skip columns consumed by a column span.
		if cf.ColSpan > 1 {
			col += cf.ColSpan - 1
		}
	}
	r.writer.WriteString("</w:tr>")
	return nil
}

func (r *docxRenderer) renderTableCell(cell *types.TableCell, bold, italic bool, bgColor string, cf cellFormat, cellWidth int) error {
	r.writer.WriteString("<w:tc>")
	tcp := tableCellProps{widthW: strconv.Itoa(cellWidth), widthWType: "dxa", bgColor: bgColor}
	if cf.ColSpan > 1 {
		tcp.gridSpan = cf.ColSpan
	}
	if cf.RowSpan > 1 {
		tcp.vMerge = "restart"
	}
	tcp.xml(r.writer)
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
			para := r.startParagraph(paragraphOptions{alignment: cf.HAlign})
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

// distributeWidths splits available twips across the given column weights.
// The last column absorbs rounding so the totals match available exactly.
func distributeWidths(available int, weights []int) []int {
	n := len(weights)
	widths := make([]int, n)
	if n == 0 || available <= 0 {
		return widths
	}
	total := 0
	for _, w := range weights {
		if w > 0 {
			total += w
		}
	}
	if total <= 0 {
		total = n
		for i := range weights {
			weights[i] = 1
		}
	}
	assigned := 0
	for i := 0; i < n-1; i++ {
		w := weights[i]
		if w <= 0 {
			w = 1
		}
		col := w * available / total
		widths[i] = col
		assigned += col
	}
	widths[n-1] = available - assigned
	return widths
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

// redistributeCells corrects the row layout when cells have row or column spans.
// The AsciiDoc parser may group cells from multiple source lines into a single
// raw TableRow. This function detects span format specifiers, computes the
// effective column count, and places cells into the correct rows.
// When no spans are present, the original rows and colCount are returned unchanged.
func redistributeCells(rows []*types.TableRow, rawColCount int) ([]*types.TableRow, int) {
	// Flatten all cells.
	var cells []*types.TableCell
	for _, row := range rows {
		cells = append(cells, row.Cells...)
	}
	// Check for spans.
	hasSpans := false
	maxCS := 1
	for _, c := range cells {
		if c != nil {
			cf := parseCellFormat(c.Format)
			if cf.RowSpan > 1 || cf.ColSpan > 1 {
				hasSpans = true
			}
			if cf.ColSpan > maxCS {
				maxCS = cf.ColSpan
			}
		}
	}
	if !hasSpans {
		return rows, rawColCount
	}
	// Find the correct column count by trying values from rawColCount down.
	colCount := rawColCount
	for cc := rawColCount; cc >= maxCS; cc-- {
		if fitsGrid(cells, cc) {
			colCount = cc
			break
		}
	}
	// Place cells into rows using span-aware grid placement.
	type pos struct{ r, c int }
	grid := map[pos]bool{}
	newRows := []*types.TableRow{}
	row, col := 0, 0
	ensureRow := func(r int) {
		for len(newRows) <= r {
			newRows = append(newRows, &types.TableRow{})
		}
	}
	for _, cell := range cells {
		for grid[pos{row, col}] {
			col++
			if col >= colCount {
				col = 0
				row++
			}
		}
		ensureRow(row)
		newRows[row].Cells = append(newRows[row].Cells, cell)
		cf := cellFormat{ColSpan: 1, RowSpan: 1}
		if cell != nil {
			cf = parseCellFormat(cell.Format)
		}
		for r := 0; r < cf.RowSpan; r++ {
			for c := 0; c < cf.ColSpan; c++ {
				grid[pos{row + r, col + c}] = true
			}
		}
		col += cf.ColSpan
		if col >= colCount {
			col = 0
			row++
		}
	}
	if len(newRows) == 0 {
		return rows, rawColCount
	}
	return newRows, colCount
}

// fitsGrid returns true if placing cells into a grid of colCount columns
// fills every position with no gaps.
func fitsGrid(cells []*types.TableCell, colCount int) bool {
	if colCount <= 0 {
		return false
	}
	type pos struct{ r, c int }
	grid := map[pos]bool{}
	row, col := 0, 0
	for _, cell := range cells {
		for grid[pos{row, col}] {
			col++
			if col >= colCount {
				col = 0
				row++
			}
		}
		cf := cellFormat{ColSpan: 1, RowSpan: 1}
		if cell != nil {
			cf = parseCellFormat(cell.Format)
		}
		if col+cf.ColSpan > colCount {
			return false
		}
		for r := 0; r < cf.RowSpan; r++ {
			for c := 0; c < cf.ColSpan; c++ {
				grid[pos{row + r, col + c}] = true
			}
		}
		col += cf.ColSpan
		if col >= colCount {
			col = 0
			row++
		}
	}
	// Determine total rows used and check every position is filled.
	maxRow := 0
	for p := range grid {
		if p.r > maxRow {
			maxRow = p.r
		}
	}
	for r := 0; r <= maxRow; r++ {
		for c := 0; c < colCount; c++ {
			if !grid[pos{r, c}] {
				return false
			}
		}
	}
	return true
}
