package docx

import "strings"

// --- Shared primitives ---

// borderLine represents a single OOXML border line specification.
// Used in table borders (w:tblBorders) and paragraph style borders (w:pBdr).
type borderLine struct {
	sizePt float64 // width in points
	space  int     // spacing value (0 for table borders, 4 for paragraph borders)
	color  string  // hex RGB color
}

// writeAttrs writes border attributes: w:val="single" w:sz="..." w:space="..." w:color="...".
func (bl borderLine) writeAttrs(w *strings.Builder) {
	w.WriteString(`w:val="single" w:sz="`)
	w.WriteString(itoa(ptToEighths(bl.sizePt)))
	w.WriteString(`" w:space="`)
	w.WriteString(itoa(bl.space))
	w.WriteString(`" w:color="`)
	w.WriteString(xmlAttr(bl.color))
	w.WriteString(`"`)
}

// writeShading writes a <w:shd> element if fill is non-empty; no-op otherwise.
func writeShading(w *strings.Builder, fill string) {
	if fill == "" {
		return
	}
	w.WriteString(`<w:shd w:val="clear" w:color="auto" w:fill="`)
	w.WriteString(xmlAttr(fill))
	w.WriteString(`"/>`)
}

// --- Table elements ---

type tableWidth struct {
	w     string
	wType string
}

func (tw tableWidth) xml(w *strings.Builder) {
	w.WriteString(`<w:tblW w:w="`)
	w.WriteString(xmlAttr(tw.w))
	w.WriteString(`" w:type="`)
	w.WriteString(xmlAttr(tw.wType))
	w.WriteString(`"/>`)
}

type tableBorders struct {
	top, left, bottom, right borderLine
	insideH, insideV         borderLine
}

func (tb tableBorders) xml(w *strings.Builder) {
	w.WriteString(`<w:tblBorders>`)
	w.WriteString(`<w:top `)
	tb.top.writeAttrs(w)
	w.WriteString(`/>`)
	w.WriteString(`<w:left `)
	tb.left.writeAttrs(w)
	w.WriteString(`/>`)
	w.WriteString(`<w:bottom `)
	tb.bottom.writeAttrs(w)
	w.WriteString(`/>`)
	w.WriteString(`<w:right `)
	tb.right.writeAttrs(w)
	w.WriteString(`/>`)
	w.WriteString(`<w:insideH `)
	tb.insideH.writeAttrs(w)
	w.WriteString(`/>`)
	w.WriteString(`<w:insideV `)
	tb.insideV.writeAttrs(w)
	w.WriteString(`/>`)
	w.WriteString(`</w:tblBorders>`)
}

type tableCellMargins struct {
	top, left, bottom, right int // twips
}

func (m tableCellMargins) xml(w *strings.Builder) {
	w.WriteString(`<w:tblCellMar>`)
	w.WriteString(`<w:top w:w="`)
	w.WriteString(itoa(m.top))
	w.WriteString(`" w:type="dxa"/>`)
	w.WriteString(`<w:left w:w="`)
	w.WriteString(itoa(m.left))
	w.WriteString(`" w:type="dxa"/>`)
	w.WriteString(`<w:bottom w:w="`)
	w.WriteString(itoa(m.bottom))
	w.WriteString(`" w:type="dxa"/>`)
	w.WriteString(`<w:right w:w="`)
	w.WriteString(itoa(m.right))
	w.WriteString(`" w:type="dxa"/>`)
	w.WriteString(`</w:tblCellMar>`)
}

type tableProps struct {
	width   tableWidth
	borders tableBorders
	cellMar *tableCellMargins // nil = omit
}

func (tp tableProps) xml(w *strings.Builder) {
	w.WriteString(`<w:tblPr>`)
	tp.width.xml(w)
	tp.borders.xml(w)
	if tp.cellMar != nil {
		tp.cellMar.xml(w)
	}
	w.WriteString(`</w:tblPr>`)
}

type tableGrid struct {
	colWidths []int // twips, one per column
}

func (g tableGrid) xml(w *strings.Builder) {
	w.WriteString(`<w:tblGrid>`)
	for _, cw := range g.colWidths {
		w.WriteString(`<w:gridCol w:w="`)
		w.WriteString(itoa(cw))
		w.WriteString(`"/>`)
	}
	w.WriteString(`</w:tblGrid>`)
}

type tableCellProps struct {
	widthW     string
	widthWType string
	bgColor    string // shading fill; empty = no shading
}

func (tcp tableCellProps) xml(w *strings.Builder) {
	w.WriteString(`<w:tcPr><w:tcW w:w="`)
	w.WriteString(xmlAttr(tcp.widthW))
	w.WriteString(`" w:type="`)
	w.WriteString(xmlAttr(tcp.widthWType))
	w.WriteString(`"/>`)
	writeShading(w, tcp.bgColor)
	w.WriteString(`</w:tcPr>`)
}

// --- Section properties ---

type sectionProps struct {
	pageW, pageH                                     int // twips
	marginTop, marginRight, marginBottom, marginLeft int // twips
	headerRelID                                      string
	footerRelID                                      string
}

func (sp sectionProps) xml(w *strings.Builder) {
	w.WriteString(`<w:sectPr>`)
	if sp.headerRelID != "" {
		w.WriteString(`<w:headerReference w:type="default" r:id="`)
		w.WriteString(sp.headerRelID)
		w.WriteString(`"/>`)
	}
	if sp.footerRelID != "" {
		w.WriteString(`<w:footerReference w:type="default" r:id="`)
		w.WriteString(sp.footerRelID)
		w.WriteString(`"/>`)
	}
	w.WriteString(`<w:pgSz w:w="`)
	w.WriteString(itoa(sp.pageW))
	w.WriteString(`" w:h="`)
	w.WriteString(itoa(sp.pageH))
	w.WriteString(`"/>`)
	w.WriteString(`<w:pgMar w:top="`)
	w.WriteString(itoa(sp.marginTop))
	w.WriteString(`" w:right="`)
	w.WriteString(itoa(sp.marginRight))
	w.WriteString(`" w:bottom="`)
	w.WriteString(itoa(sp.marginBottom))
	w.WriteString(`" w:left="`)
	w.WriteString(itoa(sp.marginLeft))
	w.WriteString(`" w:header="709" w:footer="709" w:gutter="0"/>`)
	w.WriteString(`</w:sectPr>`)
}

// --- Document defaults ---

type docDefaults struct {
	font       string
	size       string // half-points as string
	afterTwips int
	lineVal    int
	align      string // OOXML alignment value; empty = omit w:jc
}

func (dd docDefaults) xml(w *strings.Builder) {
	w.WriteString(`<w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="`)
	w.WriteString(xmlAttr(dd.font))
	w.WriteString(`" w:hAnsi="`)
	w.WriteString(xmlAttr(dd.font))
	w.WriteString(`" w:cs="`)
	w.WriteString(xmlAttr(dd.font))
	w.WriteString(`"/><w:sz w:val="`)
	w.WriteString(dd.size)
	w.WriteString(`"/><w:szCs w:val="`)
	w.WriteString(dd.size)
	w.WriteString(`"/></w:rPr></w:rPrDefault>`)
	w.WriteString(`<w:pPrDefault><w:pPr><w:spacing w:after="`)
	w.WriteString(itoa(dd.afterTwips))
	w.WriteString(`" w:line="`)
	w.WriteString(itoa(dd.lineVal))
	w.WriteString(`" w:lineRule="auto"/>`)
	if dd.align != "" {
		w.WriteString(`<w:jc w:val="`)
		w.WriteString(xmlAttr(dd.align))
		w.WriteString(`"/>`)
	}
	w.WriteString(`</w:pPr></w:pPrDefault>`)
	w.WriteString(`</w:docDefaults>`)
}

// --- Character styles ---

type charStyle struct {
	id        string
	name      string
	color     string
	bold      bool
	italic    bool
	underline bool
	vertAlign string // "superscript", "subscript", or ""
}

func (cs charStyle) xml(w *strings.Builder) {
	w.WriteString(`<w:style w:type="character" w:styleId="`)
	w.WriteString(xmlAttr(cs.id))
	w.WriteString(`"><w:name w:val="`)
	w.WriteString(xmlAttr(cs.name))
	w.WriteString(`"/><w:rPr>`)
	if cs.color != "" {
		w.WriteString(`<w:color w:val="`)
		w.WriteString(xmlAttr(cs.color))
		w.WriteString(`"/>`)
	}
	if cs.bold {
		w.WriteString(`<w:b/>`)
	}
	if cs.italic {
		w.WriteString(`<w:i/>`)
	}
	if cs.underline {
		w.WriteString(`<w:u w:val="single"/>`)
	}
	if cs.vertAlign != "" {
		w.WriteString(`<w:vertAlign w:val="`)
		w.WriteString(xmlAttr(cs.vertAlign))
		w.WriteString(`"/>`)
	}
	w.WriteString(`</w:rPr></w:style>`)
}
