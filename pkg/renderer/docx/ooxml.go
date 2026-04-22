package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	relTypeOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	relTypeStyles         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relTypeNumbering      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	relTypeFootnotes      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes"
	relTypeHyperlink      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	relTypeImage          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
)

type docxDocument struct {
	body         strings.Builder
	footnotes    strings.Builder
	rels         []relationship
	media        []mediaItem
	numbering    []numberingDefinition
	nextRelID    int
	nextMediaID  int
	nextDrawing  int
	nextBookmark int
	nextNumID    int
	nextAbsNumID int
	hasFootnotes bool
}

type relationship struct {
	ID         string
	Type       string
	Target     string
	TargetMode string
}

type mediaItem struct {
	Name        string
	Data        []byte
	ContentType string
}

type numberingDefinition struct {
	AbstractID int
	NumID      int
	Format     string
	Start      int
}

func newDocxDocument() *docxDocument {
	return &docxDocument{
		nextRelID:    10,
		nextDrawing:  1,
		nextNumID:    1,
		nextAbsNumID: 1,
	}
}

func (d *docxDocument) addExternalRelationship(relType, target string) string {
	id := "rId" + strconv.Itoa(d.nextRelID)
	d.nextRelID++
	d.rels = append(d.rels, relationship{
		ID:         id,
		Type:       relType,
		Target:     target,
		TargetMode: "External",
	})
	return id
}

func (d *docxDocument) addImage(data []byte, source string) (string, string) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(source), "."))
	if ext == "jpg" {
		ext = "jpeg"
	}
	if ext == "" {
		ext = "png"
	}
	d.nextMediaID++
	name := fmt.Sprintf("image%d.%s", d.nextMediaID, ext)
	d.media = append(d.media, mediaItem{
		Name:        name,
		Data:        data,
		ContentType: imageContentType(ext),
	})
	id := "rId" + strconv.Itoa(d.nextRelID)
	d.nextRelID++
	d.rels = append(d.rels, relationship{
		ID:     id,
		Type:   relTypeImage,
		Target: "media/" + name,
	})
	return id, name
}

func (d *docxDocument) addNumbering(format string, start int) (numID int) {
	if start <= 0 {
		start = 1
	}
	def := numberingDefinition{
		AbstractID: d.nextAbsNumID,
		NumID:      d.nextNumID,
		Format:     format,
		Start:      start,
	}
	d.nextAbsNumID++
	d.nextNumID++
	d.numbering = append(d.numbering, def)
	return def.NumID
}

func (d *docxDocument) nextBookmarkID() int {
	id := d.nextBookmark
	d.nextBookmark++
	return id
}

func (d *docxDocument) nextDrawingID() int {
	id := d.nextDrawing
	d.nextDrawing++
	return id
}

func (d *docxDocument) WriteTo(output io.Writer) error {
	buf := bytes.NewBuffer(nil)
	zw := zip.NewWriter(buf)
	files := map[string]string{
		"[Content_Types].xml":          d.contentTypesXML(),
		"_rels/.rels":                  packageRelsXML(),
		"word/document.xml":            d.documentXML(),
		"word/styles.xml":              stylesXML(),
		"word/numbering.xml":           d.numberingXML(),
		"word/_rels/document.xml.rels": d.documentRelsXML(),
	}
	if d.hasFootnotes {
		files["word/footnotes.xml"] = d.footnotesXML()
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := writeZipFile(zw, name, []byte(files[name])); err != nil {
			_ = zw.Close()
			return err
		}
	}
	for _, media := range d.media {
		if err := writeZipFile(zw, "word/media/"+media.Name, media.Data); err != nil {
			_ = zw.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	_, err := output.Write(buf.Bytes())
	return err
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func packageRelsXML() string {
	return xmlHeader() +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="` + relTypeOfficeDocument + `" Target="word/document.xml"/>` +
		`</Relationships>`
}

func (d *docxDocument) documentRelsXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	b.WriteString(`<Relationship Id="rId1" Type="` + relTypeStyles + `" Target="styles.xml"/>`)
	b.WriteString(`<Relationship Id="rId2" Type="` + relTypeNumbering + `" Target="numbering.xml"/>`)
	if d.hasFootnotes {
		b.WriteString(`<Relationship Id="rId3" Type="` + relTypeFootnotes + `" Target="footnotes.xml"/>`)
	}
	for _, rel := range d.rels {
		b.WriteString(`<Relationship Id="`)
		b.WriteString(xmlAttr(rel.ID))
		b.WriteString(`" Type="`)
		b.WriteString(xmlAttr(rel.Type))
		b.WriteString(`" Target="`)
		b.WriteString(xmlAttr(rel.Target))
		b.WriteString(`"`)
		if rel.TargetMode != "" {
			b.WriteString(` TargetMode="`)
			b.WriteString(xmlAttr(rel.TargetMode))
			b.WriteString(`"`)
		}
		b.WriteString(`/>`)
	}
	b.WriteString(`</Relationships>`)
	return b.String()
}

func (d *docxDocument) documentXML() string {
	return xmlHeader() +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"` +
		` xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"` +
		` xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"` +
		` xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"` +
		` xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">` +
		`<w:body>` + d.body.String() + sectionPropertiesXML() + `</w:body></w:document>`
}

func sectionPropertiesXML() string {
	return `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134" w:header="709" w:footer="709" w:gutter="0"/></w:sectPr>`
}

func (d *docxDocument) contentTypesXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	seen := map[string]string{}
	for _, media := range d.media {
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(media.Name), "."))
		if ext == "jpg" {
			ext = "jpeg"
		}
		seen[ext] = media.ContentType
	}
	exts := make([]string, 0, len(seen))
	for ext := range seen {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	for _, ext := range exts {
		b.WriteString(`<Default Extension="`)
		b.WriteString(xmlAttr(ext))
		b.WriteString(`" ContentType="`)
		b.WriteString(xmlAttr(seen[ext]))
		b.WriteString(`"/>`)
	}
	b.WriteString(`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>`)
	b.WriteString(`<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>`)
	b.WriteString(`<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>`)
	if d.hasFootnotes {
		b.WriteString(`<Override PartName="/word/footnotes.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml"/>`)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func (d *docxDocument) footnotesXML() string {
	return xmlHeader() +
		`<w:footnotes xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<w:footnote w:type="separator" w:id="-1"><w:p><w:r><w:separator/></w:r></w:p></w:footnote>` +
		`<w:footnote w:type="continuationSeparator" w:id="0"><w:p><w:r><w:continuationSeparator/></w:r></w:p></w:footnote>` +
		d.footnotes.String() +
		`</w:footnotes>`
}

func (d *docxDocument) numberingXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	for _, def := range d.numbering {
		b.WriteString(`<w:abstractNum w:abstractNumId="`)
		b.WriteString(strconv.Itoa(def.AbstractID))
		b.WriteString(`">`)
		for level := 0; level < 9; level++ {
			d.writeNumberingLevel(b, def, level)
		}
		b.WriteString(`</w:abstractNum>`)
		b.WriteString(`<w:num w:numId="`)
		b.WriteString(strconv.Itoa(def.NumID))
		b.WriteString(`"><w:abstractNumId w:val="`)
		b.WriteString(strconv.Itoa(def.AbstractID))
		b.WriteString(`"/></w:num>`)
	}
	b.WriteString(`</w:numbering>`)
	return b.String()
}

func (d *docxDocument) writeNumberingLevel(b *strings.Builder, def numberingDefinition, level int) {
	indent := 720 + level*360
	hanging := 360
	b.WriteString(`<w:lvl w:ilvl="`)
	b.WriteString(strconv.Itoa(level))
	b.WriteString(`">`)
	if level == 0 {
		b.WriteString(`<w:start w:val="`)
		b.WriteString(strconv.Itoa(def.Start))
		b.WriteString(`"/>`)
	} else {
		b.WriteString(`<w:start w:val="1"/>`)
	}
	if def.Format == "bullet" {
		bullet := []string{"•", "◦", "▪"}[level%3]
		b.WriteString(`<w:numFmt w:val="bullet"/><w:lvlText w:val="`)
		b.WriteString(xmlAttr(bullet))
		b.WriteString(`"/><w:rPr><w:rFonts w:ascii="Symbol" w:hAnsi="Symbol"/></w:rPr>`)
	} else {
		b.WriteString(`<w:numFmt w:val="`)
		b.WriteString(xmlAttr(def.Format))
		b.WriteString(`"/><w:lvlText w:val="%`)
		b.WriteString(strconv.Itoa(level + 1))
		b.WriteString(`."/>`)
	}
	b.WriteString(`<w:pPr><w:ind w:left="`)
	b.WriteString(strconv.Itoa(indent))
	b.WriteString(`" w:hanging="`)
	b.WriteString(strconv.Itoa(hanging))
	b.WriteString(`"/></w:pPr></w:lvl>`)
}

func stylesXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	b.WriteString(`<w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Aptos" w:hAnsi="Aptos" w:cs="Aptos"/><w:sz w:val="22"/><w:szCs w:val="22"/></w:rPr></w:rPrDefault></w:docDefaults>`)
	b.WriteString(styleParagraph("Normal", "Normal", "", 22, false, false))
	b.WriteString(styleParagraph("Title", "Title", "", 40, true, false))
	b.WriteString(styleParagraph("Subtitle", "Subtitle", "", 24, false, true))
	for i := 1; i <= 9; i++ {
		size := 32 - i*2
		if size < 20 {
			size = 20
		}
		b.WriteString(styleParagraph("Heading"+strconv.Itoa(i), "heading "+strconv.Itoa(i), "", size, true, false))
	}
	b.WriteString(styleParagraph("Quote", "Quote", "", 22, false, true))
	b.WriteString(styleParagraph("Admonition", "Admonition", "", 22, false, false))
	b.WriteString(styleParagraph("Caption", "Caption", "", 20, false, true))
	b.WriteString(styleParagraph("CodeBlock", "Code Block", "Courier New", 20, false, false))
	b.WriteString(styleParagraph("ListParagraph", "List Paragraph", "", 22, false, false))
	b.WriteString(styleParagraph("FootnoteText", "Footnote Text", "", 18, false, false))
	b.WriteString(`<w:style w:type="character" w:styleId="Hyperlink"><w:name w:val="Hyperlink"/><w:rPr><w:color w:val="0563C1"/><w:u w:val="single"/></w:rPr></w:style>`)
	b.WriteString(`<w:style w:type="character" w:styleId="FootnoteReference"><w:name w:val="Footnote Reference"/><w:rPr><w:vertAlign w:val="superscript"/></w:rPr></w:style>`)
	b.WriteString(`</w:styles>`)
	return b.String()
}

func styleParagraph(id, name, font string, size int, bold, italic bool) string {
	b := &strings.Builder{}
	b.WriteString(`<w:style w:type="paragraph" w:styleId="`)
	b.WriteString(xmlAttr(id))
	b.WriteString(`"><w:name w:val="`)
	b.WriteString(xmlAttr(name))
	b.WriteString(`"/><w:rPr>`)
	if font != "" {
		b.WriteString(`<w:rFonts w:ascii="`)
		b.WriteString(xmlAttr(font))
		b.WriteString(`" w:hAnsi="`)
		b.WriteString(xmlAttr(font))
		b.WriteString(`" w:cs="`)
		b.WriteString(xmlAttr(font))
		b.WriteString(`"/>`)
	}
	if bold {
		b.WriteString(`<w:b/>`)
	}
	if italic {
		b.WriteString(`<w:i/>`)
	}
	b.WriteString(`<w:sz w:val="`)
	b.WriteString(strconv.Itoa(size))
	b.WriteString(`"/><w:szCs w:val="`)
	b.WriteString(strconv.Itoa(size))
	b.WriteString(`"/></w:rPr></w:style>`)
	return b.String()
}

func xmlHeader() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
}

func xmlText(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(s)
}

func xmlAttr(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return replacer.Replace(s)
}

func imageContentType(ext string) string {
	switch strings.ToLower(ext) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "bmp":
		return "image/bmp"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}
