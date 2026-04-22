package docx_test

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
	"github.com/lukewilliamboswell/libasciidoc/testsupport"
)

// ---------- document-level helpers ----------

type renderedDocx struct {
	data  []byte
	files map[string][]byte
}

func renderDocx(source string, settings ...configuration.Setting) renderedDocx {
	result, err := testsupport.RenderDOCX(source, settings...)
	Expect(err).ToNot(HaveOccurred())
	Expect(result).ToNot(BeEmpty())
	doc := openDocx(result)
	Expect(doc.files).To(HaveKey("[Content_Types].xml"))
	Expect(doc.files).To(HaveKey("_rels/.rels"))
	Expect(doc.files).To(HaveKey("word/document.xml"))
	Expect(doc.files).To(HaveKey("word/styles.xml"))
	Expect(doc.files).To(HaveKey("word/numbering.xml"))
	Expect(doc.files).To(HaveKey("word/_rels/document.xml.rels"))
	return doc
}

func openDocx(data []byte) renderedDocx {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	Expect(err).ToNot(HaveOccurred())
	files := map[string][]byte{}
	for _, file := range reader.File {
		rc, err := file.Open()
		Expect(err).ToNot(HaveOccurred())
		content, err := io.ReadAll(rc)
		Expect(err).ToNot(HaveOccurred())
		Expect(rc.Close()).To(Succeed())
		files[file.Name] = content
	}
	return renderedDocx{data: data, files: files}
}

func (d renderedDocx) documentXML() string {
	return string(d.files["word/document.xml"])
}

func (d renderedDocx) relationshipsXML() string {
	return string(d.files["word/_rels/document.xml.rels"])
}

func (d renderedDocx) numberingXML() string {
	return string(d.files["word/numbering.xml"])
}

func (d renderedDocx) footnotesXML() string {
	return string(d.files["word/footnotes.xml"])
}

func (d renderedDocx) contentTypesXML() string {
	return string(d.files["[Content_Types].xml"])
}

// text returns all character data from document.xml, normalized to single spaces.
// Useful for smoke checks, but not for whitespace-sensitive assertions.
func (d renderedDocx) text() string {
	return textFromXML(d.documentXML())
}

func textFromXML(content string) string {
	decoder := xml.NewDecoder(strings.NewReader(content))
	var result strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		Expect(err).ToNot(HaveOccurred())
		if text, ok := token.(xml.CharData); ok {
			result.WriteString(string(text))
			result.WriteString(" ")
		}
	}
	return strings.Join(strings.Fields(result.String()), " ")
}

// ---------- structural XML inspection ----------

// parsedParagraph represents a w:p element with its properties and child runs.
type parsedParagraph struct {
	Style    string // w:pStyle val
	NumID    string // w:numId val (from w:numPr)
	NumLevel string // w:ilvl val (from w:numPr)
	// Children preserves the document order of runs and hyperlinks
	Children []interface{} // parsedRun or parsedHyperlink
	// Convenience accessors (also available via Children)
	Runs  []parsedRun
	Links []parsedHyperlink
	// raw bookmark names found in this paragraph
	Bookmarks []string
}

// text returns the concatenated text of all children in document order.
func (p parsedParagraph) text() string {
	var b strings.Builder
	for _, child := range p.Children {
		switch c := child.(type) {
		case parsedRun:
			b.WriteString(c.Text)
		case parsedHyperlink:
			for _, r := range c.Runs {
				b.WriteString(r.Text)
			}
		}
	}
	return b.String()
}

type parsedRun struct {
	Text        string
	Bold        bool
	Italic      bool
	Monospace   bool // w:rFonts ascii="Courier New"
	Highlight   bool // w:highlight
	Subscript   bool // w:vertAlign val="subscript"
	Superscript bool // w:vertAlign val="superscript"
	Underline   bool // w:u
	Color       string
	CharStyle   string // w:rStyle val
	// true if this run contains a w:footnoteReference
	FootnoteRef bool
	FootnoteID  string
	// true if this run contains a w:drawing (image)
	HasDrawing bool
}

type parsedHyperlink struct {
	RelID  string // r:id
	Anchor string // w:anchor (internal link)
	Runs   []parsedRun
}

// parsedTable represents a w:tbl element.
type parsedTable struct {
	Rows []parsedTableRow
}

type parsedTableRow struct {
	Cells []parsedTableCell
}

type parsedTableCell struct {
	Paragraphs []parsedParagraph
}

// parseParagraphs extracts all top-level w:p elements from document.xml.
func (d renderedDocx) parseParagraphs() []parsedParagraph {
	return parseParagraphsFromXML(d.documentXML())
}

// parseTables extracts all top-level w:tbl elements from document.xml.
func (d renderedDocx) parseTables() []parsedTable {
	return parseTablesFromXML(d.documentXML())
}

// findParagraph returns the first paragraph whose text contains the substring.
func (d renderedDocx) findParagraph(substr string) *parsedParagraph {
	for _, p := range d.parseParagraphs() {
		if strings.Contains(p.text(), substr) {
			return &p
		}
	}
	return nil
}

// findRun returns the first run across all paragraphs whose text contains the substring.
func (d renderedDocx) findRun(substr string) *parsedRun {
	for _, p := range d.parseParagraphs() {
		for _, r := range p.Runs {
			if strings.Contains(r.Text, substr) {
				return &r
			}
		}
		for _, l := range p.Links {
			for _, r := range l.Runs {
				if strings.Contains(r.Text, substr) {
					return &r
				}
			}
		}
	}
	// Also search inside tables
	for _, t := range d.parseTables() {
		for _, row := range t.Rows {
			for _, cell := range row.Cells {
				for _, p := range cell.Paragraphs {
					for _, r := range p.Runs {
						if strings.Contains(r.Text, substr) {
							return &r
						}
					}
				}
			}
		}
	}
	return nil
}

// findTableCellRun finds a run inside a table cell.
func (d renderedDocx) findTableCellRun(substr string) *parsedRun {
	for _, t := range d.parseTables() {
		for _, row := range t.Rows {
			for _, cell := range row.Cells {
				for _, p := range cell.Paragraphs {
					for _, r := range p.Runs {
						if strings.Contains(r.Text, substr) {
							return &r
						}
					}
				}
			}
		}
	}
	return nil
}

// parseRelationships extracts all Relationship elements from document.xml.rels.
type parsedRelationship struct {
	ID         string
	Type       string
	Target     string
	TargetMode string
}

func (d renderedDocx) parseRelationships() []parsedRelationship {
	type xmlRel struct {
		ID         string `xml:"Id,attr"`
		Type       string `xml:"Type,attr"`
		Target     string `xml:"Target,attr"`
		TargetMode string `xml:"TargetMode,attr"`
	}
	type xmlRels struct {
		Rels []xmlRel `xml:"Relationship"`
	}
	var rels xmlRels
	Expect(xml.Unmarshal(d.files["word/_rels/document.xml.rels"], &rels)).To(Succeed())
	result := make([]parsedRelationship, len(rels.Rels))
	for i, r := range rels.Rels {
		result[i] = parsedRelationship{ID: r.ID, Type: r.Type, Target: r.Target, TargetMode: r.TargetMode}
	}
	return result
}

func (d renderedDocx) findRelationship(target string) *parsedRelationship {
	for _, r := range d.parseRelationships() {
		if r.Target == target {
			return &r
		}
	}
	return nil
}

func (d renderedDocx) findRelationshipByID(id string) *parsedRelationship {
	for _, r := range d.parseRelationships() {
		if r.ID == id {
			return &r
		}
	}
	return nil
}

// parseContentTypes extracts Default and Override entries from [Content_Types].xml.
type parsedContentType struct {
	Extension   string // from Default
	PartName    string // from Override
	ContentType string
}

func (d renderedDocx) parseContentTypes() []parsedContentType {
	type xmlDefault struct {
		Extension   string `xml:"Extension,attr"`
		ContentType string `xml:"ContentType,attr"`
	}
	type xmlOverride struct {
		PartName    string `xml:"PartName,attr"`
		ContentType string `xml:"ContentType,attr"`
	}
	type xmlTypes struct {
		Defaults  []xmlDefault  `xml:"Default"`
		Overrides []xmlOverride `xml:"Override"`
	}
	var types xmlTypes
	Expect(xml.Unmarshal(d.files["[Content_Types].xml"], &types)).To(Succeed())
	var result []parsedContentType
	for _, d := range types.Defaults {
		result = append(result, parsedContentType{Extension: d.Extension, ContentType: d.ContentType})
	}
	for _, o := range types.Overrides {
		result = append(result, parsedContentType{PartName: o.PartName, ContentType: o.ContentType})
	}
	return result
}

func (d renderedDocx) hasContentTypeForExtension(ext, contentType string) bool {
	for _, ct := range d.parseContentTypes() {
		if ct.Extension == ext && ct.ContentType == contentType {
			return true
		}
	}
	return false
}

// ---------- XML parsing engine ----------
//
// All parse helpers call Expect on non-EOF decoder errors so that
// malformed XML causes an immediate test failure rather than a silent
// partial parse.

func nextToken(decoder *xml.Decoder) (xml.Token, bool) {
	token, err := decoder.Token()
	if err == io.EOF {
		return nil, false
	}
	Expect(err).ToNot(HaveOccurred(), "unexpected XML decode error")
	return token, true
}

func parseParagraphsFromXML(xmlContent string) []parsedParagraph {
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	var paragraphs []parsedParagraph
	for {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		if se, ok := token.(xml.StartElement); ok && se.Name.Local == "p" && se.Name.Space == nsMl {
			p := parseParagraphElement(decoder)
			paragraphs = append(paragraphs, p)
		}
	}
	return paragraphs
}

func parseParagraphElement(decoder *xml.Decoder) parsedParagraph {
	var p parsedParagraph
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			switch {
			case t.Name.Local == "pPr" && t.Name.Space == nsMl:
				parseParagraphProperties(decoder, &p)
				depth--
			case t.Name.Local == "r" && t.Name.Space == nsMl:
				r := parseRunElement(decoder)
				p.Runs = append(p.Runs, r)
				p.Children = append(p.Children, r)
				depth--
			case t.Name.Local == "hyperlink" && t.Name.Space == nsMl:
				h := parseHyperlinkElement(decoder, t)
				p.Links = append(p.Links, h)
				p.Children = append(p.Children, h)
				depth--
			case t.Name.Local == "bookmarkStart" && t.Name.Space == nsMl:
				for _, a := range t.Attr {
					if a.Name.Local == "name" {
						p.Bookmarks = append(p.Bookmarks, a.Value)
					}
				}
			}
		case xml.EndElement:
			depth--
		}
	}
	return p
}

func parseParagraphProperties(decoder *xml.Decoder, p *parsedParagraph) {
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "pStyle":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						p.Style = a.Value
					}
				}
			case "numId":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						p.NumID = a.Value
					}
				}
			case "ilvl":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						p.NumLevel = a.Value
					}
				}
			}
		case xml.EndElement:
			depth--
		}
	}
}

func parseRunElement(decoder *xml.Decoder) parsedRun {
	var r parsedRun
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			switch {
			case t.Name.Local == "rPr" && t.Name.Space == nsMl:
				parseRunProperties(decoder, &r)
				depth--
			case t.Name.Local == "t" && t.Name.Space == nsMl:
				r.Text += collectText(decoder)
				depth--
			case t.Name.Local == "footnoteReference" && t.Name.Space == nsMl:
				r.FootnoteRef = true
				for _, a := range t.Attr {
					if a.Name.Local == "id" {
						r.FootnoteID = a.Value
					}
				}
			case t.Name.Local == "drawing":
				r.HasDrawing = true
			case t.Name.Local == "tab" && t.Name.Space == nsMl:
				r.Text += "\t"
			case t.Name.Local == "br" && t.Name.Space == nsMl:
				r.Text += "\n"
			}
		case xml.EndElement:
			depth--
		}
	}
	return r
}

func parseRunProperties(decoder *xml.Decoder, r *parsedRun) {
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "b":
				r.Bold = true
			case "i":
				r.Italic = true
			case "rFonts":
				for _, a := range t.Attr {
					if a.Name.Local == "ascii" && a.Value == "Courier New" {
						r.Monospace = true
					}
				}
			case "highlight":
				r.Highlight = true
			case "vertAlign":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						switch a.Value {
						case "subscript":
							r.Subscript = true
						case "superscript":
							r.Superscript = true
						}
					}
				}
			case "u":
				r.Underline = true
			case "color":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						r.Color = a.Value
					}
				}
			case "rStyle":
				for _, a := range t.Attr {
					if a.Name.Local == "val" {
						r.CharStyle = a.Value
					}
				}
			}
		case xml.EndElement:
			depth--
		}
	}
}

func parseHyperlinkElement(decoder *xml.Decoder, start xml.StartElement) parsedHyperlink {
	var h parsedHyperlink
	for _, a := range start.Attr {
		switch {
		case a.Name.Local == "id":
			h.RelID = a.Value
		case a.Name.Local == "anchor":
			h.Anchor = a.Value
		}
	}
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "r" && t.Name.Space == nsMl {
				r := parseRunElement(decoder)
				h.Runs = append(h.Runs, r)
				depth--
			}
		case xml.EndElement:
			depth--
		}
	}
	return h
}

func collectText(decoder *xml.Decoder) string {
	var text string
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.CharData:
			text += string(t)
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return text
}

func parseTablesFromXML(xmlContent string) []parsedTable {
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	var tables []parsedTable
	for {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		if se, ok := token.(xml.StartElement); ok && se.Name.Local == "tbl" && se.Name.Space == nsMl {
			t := parseTableElement(decoder)
			tables = append(tables, t)
		}
	}
	return tables
}

func parseTableElement(decoder *xml.Decoder) parsedTable {
	var t parsedTable
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch tok := token.(type) {
		case xml.StartElement:
			depth++
			if tok.Name.Local == "tr" && tok.Name.Space == nsMl {
				row := parseTableRowElement(decoder)
				t.Rows = append(t.Rows, row)
				depth--
			}
		case xml.EndElement:
			depth--
		}
	}
	return t
}

func parseTableRowElement(decoder *xml.Decoder) parsedTableRow {
	var row parsedTableRow
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "tc" && t.Name.Space == nsMl {
				cell := parseTableCellElement(decoder)
				row.Cells = append(row.Cells, cell)
				depth--
			}
		case xml.EndElement:
			depth--
		}
	}
	return row
}

func parseTableCellElement(decoder *xml.Decoder) parsedTableCell {
	var cell parsedTableCell
	depth := 1
	for depth > 0 {
		token, ok := nextToken(decoder)
		if !ok {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "p" && t.Name.Space == nsMl {
				p := parseParagraphElement(decoder)
				cell.Paragraphs = append(cell.Paragraphs, p)
				depth--
			}
		case xml.EndElement:
			depth--
		}
	}
	return cell
}

// parseNumberingDefs parses numbering.xml to extract abstractNum definitions.
type parsedNumberingDef struct {
	AbstractID string
	NumID      string
	Levels     []parsedNumberingLevel
}

type parsedNumberingLevel struct {
	Level  string // w:ilvl
	Format string // w:numFmt val
	Start  string // w:start val
}

func (d renderedDocx) parseNumberingDefs() []parsedNumberingDef {
	type xmlStart struct {
		Val string `xml:"val,attr"`
	}
	type xmlNumFmt struct {
		Val string `xml:"val,attr"`
	}
	type xmlLvl struct {
		Ilvl   string     `xml:"ilvl,attr"`
		Start  *xmlStart  `xml:"start"`
		NumFmt *xmlNumFmt `xml:"numFmt"`
	}
	type xmlAbstractNum struct {
		AbstractNumID string   `xml:"abstractNumId,attr"`
		Levels        []xmlLvl `xml:"lvl"`
	}
	type xmlAbstractNumIDRef struct {
		Val string `xml:"val,attr"`
	}
	type xmlNum struct {
		NumID        string              `xml:"numId,attr"`
		AbstractRef  xmlAbstractNumIDRef `xml:"abstractNumId"`
	}
	type xmlNumbering struct {
		AbstractNums []xmlAbstractNum `xml:"abstractNum"`
		Nums         []xmlNum         `xml:"num"`
	}
	var numbering xmlNumbering
	Expect(xml.Unmarshal(d.files["word/numbering.xml"], &numbering)).To(Succeed())

	// Build map of abstractNumId -> levels
	absMap := map[string][]parsedNumberingLevel{}
	for _, abs := range numbering.AbstractNums {
		var levels []parsedNumberingLevel
		for _, lvl := range abs.Levels {
			pl := parsedNumberingLevel{Level: lvl.Ilvl}
			if lvl.NumFmt != nil {
				pl.Format = lvl.NumFmt.Val
			}
			if lvl.Start != nil {
				pl.Start = lvl.Start.Val
			}
			levels = append(levels, pl)
		}
		absMap[abs.AbstractNumID] = levels
	}

	var result []parsedNumberingDef
	for _, num := range numbering.Nums {
		def := parsedNumberingDef{
			NumID:      num.NumID,
			AbstractID: num.AbstractRef.Val,
			Levels:     absMap[num.AbstractRef.Val],
		}
		result = append(result, def)
	}
	return result
}

func (d renderedDocx) findNumberingDef(numID string) *parsedNumberingDef {
	for _, def := range d.parseNumberingDefs() {
		if def.NumID == numID {
			return &def
		}
	}
	return nil
}

const nsMl = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
