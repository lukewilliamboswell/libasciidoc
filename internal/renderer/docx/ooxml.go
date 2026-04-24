package docx

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// flateWriterPool reuses flate compressors across zip entries to avoid
// allocating fresh huffman tables for every XML part in the DOCX archive.
var flateWriterPool = sync.Pool{
	New: func() interface{} {
		fw, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
		return fw
	},
}

type pooledFlateWriter struct {
	fw *flate.Writer
}

func (w *pooledFlateWriter) Write(p []byte) (int, error) { return w.fw.Write(p) }

func (w *pooledFlateWriter) Close() error {
	err := w.fw.Close()
	if err == nil {
		flateWriterPool.Put(w.fw)
	}
	w.fw = nil
	return err
}

const (
	relTypeOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	relTypeStyles         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relTypeNumbering      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	relTypeFootnotes      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes"
	relTypeHyperlink      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	relTypeImage          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	relTypeHeader         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header"
	relTypeFooter         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"
	relTypeSettings       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	relTypeFontTable      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable"

	// twipsPerLevel is the indentation increment per numbering level (in twips).
	twipsPerLevel = 360
	// listHangingTwips is the hanging indent for list items (in twips).
	listHangingTwips = 360
)

type docxDocument struct {
	body             strings.Builder
	footnotes        strings.Builder
	rels             []relationship
	media            []mediaItem
	numbering        []numberingDefinition
	abstractNumByFmt map[string]int    // format -> abstractNumId (shared)
	legalAbsNumID    int               // abstractNum for multi-level legal numbering
	legalNumID       int               // num instance for heading numbering
	legalListNums    []legalListNumDef // per-list num instances for legal lists
	nextRelID        int
	nextMediaID      int
	nextDrawing      int
	nextBookmark     int
	nextNumID        int
	nextAbsNumID     int
	hasFootnotes     bool
	hasHeader        bool
	hasFooter        bool
	headerRelID      string
	footerRelID      string
	theme            *DocxTheme
	title            string         // document title for core properties
	creators         string         // semicolon-joined author names
	created          time.Time      // document creation timestamp
	modified         time.Time      // document modification timestamp
	bookmarkNames    map[string]int // tracks sanitized bookmark names for deduplication
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
	Indent     int // base indentation in twips (set from list nesting depth)
}

func newDocxDocument() *docxDocument {
	return &docxDocument{
		nextRelID:        10,
		nextDrawing:      1,
		nextNumID:        1,
		nextAbsNumID:     1,
		abstractNumByFmt: make(map[string]int),
		bookmarkNames:    make(map[string]int),
	}
}

// uniqueBookmarkName returns a deduplicated version of the sanitized bookmark name.
// The first occurrence of a name is returned unchanged. Subsequent occurrences are
// suffixed with _2, _3, … to satisfy the OOXML requirement that w:name values are
// unique within a document.
func (d *docxDocument) uniqueBookmarkName(raw string) string {
	name := sanitizeBookmarkName(raw)
	count := d.bookmarkNames[name]
	d.bookmarkNames[name]++
	if count == 0 {
		return name
	}
	return fmt.Sprintf("%s_%d", name, count+1)
}

// textWidthTwips returns the printable text width in twips, derived from the
// theme's page size and left/right margin settings.  This is the authoritative
// base for all table column-width calculations.
func (d *docxDocument) textWidthTwips() int {
	t := d.theme
	pageW, _ := pageSizeTwips(t.Page.Size, t.Page.Layout)
	return pageW - mmToTwips(t.Page.Margin[3]) - mmToTwips(t.Page.Margin[1])
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

func (d *docxDocument) addNumbering(format string, start, indent int) (numID int) {
	if start <= 0 {
		start = 1
	}
	// Reuse a shared abstractNum for each format type. Multiple w:num
	// instances reference the same abstractNum but use startOverride to
	// restart numbering. This matches how Word generates numbering.xml
	// and prevents counter-merging bugs.
	absID, ok := d.abstractNumByFmt[format]
	if !ok {
		absID = d.nextAbsNumID
		d.nextAbsNumID++
		d.abstractNumByFmt[format] = absID
	}
	def := numberingDefinition{
		AbstractID: absID,
		NumID:      d.nextNumID,
		Format:     format,
		Start:      start,
		Indent:     indent,
	}
	d.nextNumID++
	d.numbering = append(d.numbering, def)
	return def.NumID
}

// addLegalNumbering creates a single multi-level numbering definition for
// legal document style: decimal headings (1., 1.1, 1.1.1) followed by
// parenthetical lists ((a), (i), (A)). Returns the numID for headings.
func (d *docxDocument) addLegalNumbering() int {
	absID := d.nextAbsNumID
	d.nextAbsNumID++
	numID := d.nextNumID
	d.nextNumID++
	d.legalAbsNumID = absID
	d.legalNumID = numID
	return numID
}

// addLegalListNum creates a new w:num instance that references the shared
// legal abstractNum but overrides the start at the given ilvl. This ensures
// each list restarts at (a)/(i)/(A) independently while sharing the heading
// numbering format from the legal abstractNum.
func (d *docxDocument) addLegalListNum(ilvl int) int {
	numID := d.nextNumID
	d.nextNumID++
	d.legalListNums = append(d.legalListNums, legalListNumDef{
		NumID: numID,
		Ilvl:  ilvl,
	})
	return numID
}

type legalListNumDef struct {
	NumID int
	Ilvl  int
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

// setupHeaderFooter creates the header/footer parts if the theme defines them.
func (d *docxDocument) setupHeaderFooter() {
	if d.theme.RunningHeader.Content != "" {
		d.hasHeader = true
		d.headerRelID = "rId" + strconv.Itoa(d.nextRelID)
		d.nextRelID++
	}
	if d.theme.RunningFooter.Content != "" {
		d.hasFooter = true
		d.footerRelID = "rId" + strconv.Itoa(d.nextRelID)
		d.nextRelID++
	}
}

func (d *docxDocument) WriteTo(output io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	zw.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
		fw := flateWriterPool.Get().(*flate.Writer)
		fw.Reset(w)
		return &pooledFlateWriter{fw: fw}, nil
	})
	files := map[string]string{
		"[Content_Types].xml":          d.contentTypesXML(),
		"_rels/.rels":                  d.packageRelsXML(),
		"word/document.xml":            d.documentXML(),
		"word/styles.xml":              d.stylesXML(),
		"word/numbering.xml":           d.numberingXML(),
		"word/settings.xml":            settingsXML(),
		"word/fontTable.xml":           d.fontTableXML(),
		"word/_rels/document.xml.rels": d.documentRelsXML(),
		"docProps/core.xml":            d.corePropertiesXML(),
		"docProps/app.xml":             appPropertiesXML(),
	}
	if d.hasFootnotes {
		files["word/footnotes.xml"] = d.footnotesXML()
	}
	if d.hasHeader {
		files["word/header1.xml"] = d.headerXML()
	}
	if d.hasFooter {
		files["word/footer1.xml"] = d.footerXML()
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := writeZipFile(zw, name, []byte(files[name])); err != nil {
			zw.Close()
			return 0, err
		}
	}
	for _, media := range d.media {
		if err := writeZipFile(zw, "word/media/"+media.Name, media.Data); err != nil {
			zw.Close()
			return 0, err
		}
	}
	if err := zw.Close(); err != nil {
		return 0, err
	}
	n, err := output.Write(buf.Bytes())
	return int64(n), err
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (d *docxDocument) packageRelsXML() string {
	const relTypeCoreProperties = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"
	const relTypeExtendedProperties = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties"
	return xmlHeader() +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="` + relTypeOfficeDocument + `" Target="word/document.xml"/>` +
		`<Relationship Id="rId2" Type="` + relTypeCoreProperties + `" Target="docProps/core.xml"/>` +
		`<Relationship Id="rId3" Type="` + relTypeExtendedProperties + `" Target="docProps/app.xml"/>` +
		`</Relationships>`
}

func (d *docxDocument) documentRelsXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	b.WriteString(`<Relationship Id="rId1" Type="` + relTypeStyles + `" Target="styles.xml"/>`)
	b.WriteString(`<Relationship Id="rId2" Type="` + relTypeNumbering + `" Target="numbering.xml"/>`)
	b.WriteString(`<Relationship Id="rId4" Type="` + relTypeSettings + `" Target="settings.xml"/>`)
	b.WriteString(`<Relationship Id="rId5" Type="` + relTypeFontTable + `" Target="fontTable.xml"/>`)
	if d.hasFootnotes {
		b.WriteString(`<Relationship Id="rId3" Type="` + relTypeFootnotes + `" Target="footnotes.xml"/>`)
	}
	if d.hasHeader {
		b.WriteString(`<Relationship Id="`)
		b.WriteString(d.headerRelID)
		b.WriteString(`" Type="`)
		b.WriteString(relTypeHeader)
		b.WriteString(`" Target="header1.xml"/>`)
	}
	if d.hasFooter {
		b.WriteString(`<Relationship Id="`)
		b.WriteString(d.footerRelID)
		b.WriteString(`" Type="`)
		b.WriteString(relTypeFooter)
		b.WriteString(`" Target="footer1.xml"/>`)
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
		`<w:body>` + d.body.String() + d.sectionPropertiesXML() + `</w:body></w:document>`
}

func (d *docxDocument) sectionPropertiesXML() string {
	t := d.theme
	w, h := pageSizeTwips(t.Page.Size, t.Page.Layout)
	sp := sectionProps{
		pageW:        w,
		pageH:        h,
		marginTop:    mmToTwips(t.Page.Margin[0]),
		marginRight:  mmToTwips(t.Page.Margin[1]),
		marginBottom: mmToTwips(t.Page.Margin[2]),
		marginLeft:   mmToTwips(t.Page.Margin[3]),
	}
	if d.hasHeader {
		sp.headerRelID = d.headerRelID
	}
	if d.hasFooter {
		sp.footerRelID = d.footerRelID
	}
	b := &strings.Builder{}
	sp.xml(b)
	return b.String()
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
	if d.hasHeader {
		b.WriteString(`<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`)
	}
	if d.hasFooter {
		b.WriteString(`<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`)
	}
	b.WriteString(`<Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/>`)
	b.WriteString(`<Override PartName="/word/fontTable.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"/>`)
	b.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	b.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
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

// headerXML generates the word/header1.xml part.
func (d *docxDocument) headerXML() string {
	return d.runningHFXML("hdr", &d.theme.RunningHeader)
}

// footerXML generates the word/footer1.xml part.
func (d *docxDocument) footerXML() string {
	return d.runningHFXML("ftr", &d.theme.RunningFooter)
}

// runningHFXML generates the XML for a running header or footer.
func (d *docxDocument) runningHFXML(tag string, hf *RunningHFTheme) string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<w:`)
	b.WriteString(tag)
	b.WriteString(` xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	b.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr>`)

	// Run properties
	hasRPr := hf.FontFamily != "" || hf.FontSize > 0 || hf.FontColor != "" || hf.FontStyle != ""
	rprStr := ""
	if hasRPr {
		rpr := &strings.Builder{}
		rpr.WriteString(`<w:rPr>`)
		if hf.FontFamily != "" {
			rpr.WriteString(`<w:rFonts w:ascii="`)
			rpr.WriteString(xmlAttr(hf.FontFamily))
			rpr.WriteString(`" w:hAnsi="`)
			rpr.WriteString(xmlAttr(hf.FontFamily))
			rpr.WriteString(`"/>`)
		}
		bold, italic := fontStyleBoldItalic(hf.FontStyle)
		if bold {
			rpr.WriteString(`<w:b/>`)
		}
		if italic {
			rpr.WriteString(`<w:i/>`)
		}
		if hf.FontColor != "" {
			rpr.WriteString(`<w:color w:val="`)
			rpr.WriteString(xmlAttr(hf.FontColor))
			rpr.WriteString(`"/>`)
		}
		if hf.FontSize > 0 {
			sz := itoa(ptToHalfPt(hf.FontSize))
			rpr.WriteString(`<w:sz w:val="`)
			rpr.WriteString(sz)
			rpr.WriteString(`"/><w:szCs w:val="`)
			rpr.WriteString(sz)
			rpr.WriteString(`"/>`)
		}
		rpr.WriteString(`</w:rPr>`)
		rprStr = rpr.String()
	}

	// Expand content template: {page-number} becomes a PAGE field
	content := hf.Content
	if strings.Contains(content, "{page-number}") {
		parts := strings.Split(content, "{page-number}")
		for i, part := range parts {
			if part != "" {
				b.WriteString(`<w:r>`)
				if hasRPr {
					b.WriteString(rprStr)
				}
				b.WriteString(`<w:t xml:space="preserve">`)
				b.WriteString(xmlText(part))
				b.WriteString(`</w:t></w:r>`)
			}
			if i < len(parts)-1 {
				// PAGE field
				b.WriteString(`<w:r>`)
				if hasRPr {
					b.WriteString(rprStr)
				}
				b.WriteString(`<w:fldChar w:fldCharType="begin"/></w:r>`)
				b.WriteString(`<w:r>`)
				if hasRPr {
					b.WriteString(rprStr)
				}
				b.WriteString(`<w:instrText xml:space="preserve"> PAGE </w:instrText></w:r>`)
				b.WriteString(`<w:r>`)
				if hasRPr {
					b.WriteString(rprStr)
				}
				b.WriteString(`<w:fldChar w:fldCharType="end"/></w:r>`)
			}
		}
	} else {
		// Plain text content
		b.WriteString(`<w:r>`)
		if hasRPr {
			b.WriteString(rprStr)
		}
		b.WriteString(`<w:t xml:space="preserve">`)
		b.WriteString(xmlText(content))
		b.WriteString(`</w:t></w:r>`)
	}

	b.WriteString(`</w:p></w:`)
	b.WriteString(tag)
	b.WriteString(`>`)
	return b.String()
}

func (d *docxDocument) numberingXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)

	// Legal multi-level numbering abstractNum (if used).
	if d.legalNumID > 0 {
		d.writeLegalAbstractNum(b)
	}

	// Emit each shared abstractNum once (OOXML: all abstractNum before all num).
	emitted := make(map[int]bool)
	for _, def := range d.numbering {
		if emitted[def.AbstractID] {
			continue
		}
		emitted[def.AbstractID] = true
		b.WriteString(`<w:abstractNum w:abstractNumId="`)
		b.WriteString(strconv.Itoa(def.AbstractID))
		b.WriteString(`">`)
		b.WriteString(`<w:nsid w:val="`)
		fmt.Fprintf(b, "%08X", 0x10000000+def.AbstractID)
		b.WriteString(`"/>`)
		b.WriteString(`<w:multiLevelType w:val="singleLevel"/>`)
		for level := 0; level < 9; level++ {
			d.writeNumberingLevel(b, def, level)
		}
		b.WriteString(`</w:abstractNum>`)
	}

	// Legal numbering w:num (for headings — continuous counter).
	if d.legalNumID > 0 {
		b.WriteString(`<w:num w:numId="`)
		b.WriteString(strconv.Itoa(d.legalNumID))
		b.WriteString(`"><w:abstractNumId w:val="`)
		b.WriteString(strconv.Itoa(d.legalAbsNumID))
		b.WriteString(`"/></w:num>`)
	}

	// Per-list w:num instances for legal lists — each restarts at (a)/(i)/(A).
	for _, ln := range d.legalListNums {
		b.WriteString(`<w:num w:numId="`)
		b.WriteString(strconv.Itoa(ln.NumID))
		b.WriteString(`"><w:abstractNumId w:val="`)
		b.WriteString(strconv.Itoa(d.legalAbsNumID))
		b.WriteString(`"/>`)
		b.WriteString(`<w:lvlOverride w:ilvl="`)
		b.WriteString(strconv.Itoa(ln.Ilvl))
		b.WriteString(`"><w:startOverride w:val="1"/>`)
		b.WriteString(`</w:lvlOverride>`)
		b.WriteString(`</w:num>`)
	}

	// Each regular list instance gets its own w:num with startOverride.
	for _, def := range d.numbering {
		b.WriteString(`<w:num w:numId="`)
		b.WriteString(strconv.Itoa(def.NumID))
		b.WriteString(`"><w:abstractNumId w:val="`)
		b.WriteString(strconv.Itoa(def.AbstractID))
		b.WriteString(`"/>`)
		b.WriteString(`<w:lvlOverride w:ilvl="0">`)
		b.WriteString(`<w:startOverride w:val="`)
		b.WriteString(strconv.Itoa(def.Start))
		b.WriteString(`"/>`)
		b.WriteString(`</w:lvlOverride>`)
		b.WriteString(`</w:num>`)
	}

	b.WriteString(`</w:numbering>`)
	return b.String()
}

// writeLegalAbstractNum writes the multi-level legal numbering definition.
//
//	ilvl 0: "1."        decimal     (clause headings)
//	ilvl 1: "1.1"       decimal     (sub-clause headings)
//	ilvl 2: "1.1.1"     decimal     (sub-sub-clause headings)
//	ilvl 3: "(a)"       lowerLetter (enumerated items)
//	ilvl 4: "(i)"       lowerRoman  (sub-items)
//	ilvl 5: "(A)"       upperLetter (sub-sub-items)
func (d *docxDocument) writeLegalAbstractNum(b *strings.Builder) {
	b.WriteString(`<w:abstractNum w:abstractNumId="`)
	b.WriteString(strconv.Itoa(d.legalAbsNumID))
	b.WriteString(`">`)
	b.WriteString(`<w:nsid w:val="`)
	fmt.Fprintf(b, "%08X", 0x20000000+d.legalAbsNumID)
	b.WriteString(`"/>`)
	b.WriteString(`<w:multiLevelType w:val="multilevel"/>`)

	type legalLevel struct {
		numFmt     string
		lvlText    string
		pStyle     string // paragraph style linked to this level
		lvlRestart int    // -1 = no restart; N = restart after level N
		indent     int    // left indent in twips
		hanging    int    // hanging indent in twips
	}

	// Indent step from theme (list.indent in pt, converted to twips).
	step := ptToTwips(d.theme.List.Indent)

	levels := []legalLevel{
		// Level 0: "1." — number at 0mm, text at 15mm
		{numFmt: "decimal", lvlText: "%1.", pStyle: "Heading2", lvlRestart: -1, indent: step * 1, hanging: step},
		// Level 1: "1.1" — number at 15mm, text at 30mm
		{numFmt: "decimal", lvlText: "%1.%2", pStyle: "Heading3", lvlRestart: -1, indent: step * 2, hanging: step},
		// Level 2: "1.1.1" — number at 30mm, text at 45mm
		{numFmt: "decimal", lvlText: "%1.%2.%3", pStyle: "Heading4", lvlRestart: -1, indent: step * 3, hanging: step},
		// Level 3: "(a)" — number at 15mm, text at 30mm
		{numFmt: "lowerLetter", lvlText: "(%4)", pStyle: "ListParagraph", lvlRestart: 1, indent: step * 2, hanging: step},
		// Level 4: "(i)" — number at 30mm, text at 45mm
		{numFmt: "lowerRoman", lvlText: "(%5)", lvlRestart: -1, indent: step * 3, hanging: step},
		// Level 5: "(A)" — number at 45mm, text at 60mm
		{numFmt: "upperLetter", lvlText: "(%6)", lvlRestart: -1, indent: step * 4, hanging: step},
	}

	for i, lvl := range levels {
		b.WriteString(`<w:lvl w:ilvl="`)
		b.WriteString(strconv.Itoa(i))
		b.WriteString(`">`)
		b.WriteString(`<w:start w:val="1"/>`)
		b.WriteString(`<w:numFmt w:val="`)
		b.WriteString(xmlAttr(lvl.numFmt))
		b.WriteString(`"/>`)
		if lvl.lvlRestart >= 0 {
			b.WriteString(`<w:lvlRestart w:val="`)
			b.WriteString(strconv.Itoa(lvl.lvlRestart))
			b.WriteString(`"/>`)
		}
		if lvl.pStyle != "" {
			b.WriteString(`<w:pStyle w:val="`)
			b.WriteString(xmlAttr(lvl.pStyle))
			b.WriteString(`"/>`)
		}
		b.WriteString(`<w:lvlText w:val="`)
		b.WriteString(xmlAttr(lvl.lvlText))
		b.WriteString(`"/>`)
		b.WriteString(`<w:lvlJc w:val="left"/>`)
		b.WriteString(`<w:pPr><w:ind w:left="`)
		b.WriteString(strconv.Itoa(lvl.indent))
		b.WriteString(`" w:hanging="`)
		b.WriteString(strconv.Itoa(lvl.hanging))
		b.WriteString(`"/></w:pPr>`)
		b.WriteString(`</w:lvl>`)
	}

	// Pad remaining levels 6-8 with decimal fallback.
	for i := len(levels); i < 9; i++ {
		indent := 480 + i*480
		b.WriteString(`<w:lvl w:ilvl="`)
		b.WriteString(strconv.Itoa(i))
		b.WriteString(`"><w:start w:val="1"/>`)
		b.WriteString(`<w:numFmt w:val="decimal"/>`)
		b.WriteString(`<w:lvlText w:val="%`)
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(`."/><w:lvlJc w:val="left"/>`)
		b.WriteString(`<w:pPr><w:ind w:left="`)
		b.WriteString(strconv.Itoa(indent))
		b.WriteString(`" w:hanging="480"/></w:pPr>`)
		b.WriteString(`</w:lvl>`)
	}

	b.WriteString(`</w:abstractNum>`)
}

func (d *docxDocument) writeNumberingLevel(b *strings.Builder, def numberingDefinition, level int) {
	baseIndent := ptToTwips(d.theme.List.Indent)
	indent := def.Indent + baseIndent + level*twipsPerLevel
	hanging := listHangingTwips
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
		b.WriteString(`"/><w:lvlJc w:val="left"/><w:rPr><w:rFonts w:ascii="Symbol" w:hAnsi="Symbol"/>`)
		if d.theme.List.MarkerFontColor != "" {
			b.WriteString(`<w:color w:val="`)
			b.WriteString(xmlAttr(d.theme.List.MarkerFontColor))
			b.WriteString(`"/>`)
		}
		b.WriteString(`</w:rPr>`)
	} else {
		b.WriteString(`<w:numFmt w:val="`)
		b.WriteString(xmlAttr(def.Format))
		b.WriteString(`"/>`)
		// Restart numbering after any heading (outline level 0 = Heading1).
		b.WriteString(`<w:lvlRestart w:val="0"/>`)
		b.WriteString(`<w:lvlText w:val="%`)
		b.WriteString(strconv.Itoa(level + 1))
		b.WriteString(`."/><w:lvlJc w:val="left"/>`)
		if d.theme.List.MarkerFontColor != "" {
			b.WriteString(`<w:rPr><w:color w:val="`)
			b.WriteString(xmlAttr(d.theme.List.MarkerFontColor))
			b.WriteString(`"/></w:rPr>`)
		}
	}
	b.WriteString(`<w:pPr><w:ind w:left="`)
	b.WriteString(strconv.Itoa(indent))
	b.WriteString(`" w:hanging="`)
	b.WriteString(strconv.Itoa(hanging))
	b.WriteString(`"/></w:pPr></w:lvl>`)
}

func (d *docxDocument) stylesXML() string {
	t := d.theme
	headingFont := t.Heading.FontFamily

	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)

	afterTwips := 160 // default ~8pt after
	if t.Prose.MarginBottom > 0 {
		afterTwips = ptToTwips(t.Prose.MarginBottom)
	}
	dd := docDefaults{
		font:       t.Base.FontFamily,
		size:       itoa(ptToHalfPt(t.Base.FontSize)),
		afterTwips: afterTwips,
		lineVal:    t.lineHeightValue(),
	}
	if align := resolveTextAlign(t.Prose.TextAlign, t.Base.TextAlign); align != "" {
		dd.align = ooxmlAlignment(align)
	}
	dd.xml(b)

	// Normal style
	baseBold, baseItalic := fontStyleBoldItalic(t.Base.FontStyle)
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Normal", name: "Normal",
		size: ptToHalfPt(t.Base.FontSize), bold: baseBold, italic: baseItalic,
		color: t.Base.FontColor,
	}))

	// Title style
	titleBold, titleItalic := fontStyleBoldItalic(t.Title.TitleFontStyle)
	titleCaps := t.Title.TitleTextTransform == "uppercase"
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Title", name: "Title",
		font: t.Title.TitleFontFamily,
		size: ptToHalfPt(t.Title.TitleFontSize), bold: titleBold, italic: titleItalic,
		caps: titleCaps, color: t.Title.TitleFontColor,
		align: t.Title.TitleTextAlign,
	}))

	// Subtitle style
	subtitleBold, subtitleItalic := fontStyleBoldItalic(t.Title.SubtitleFontStyle)
	if !subtitleBold && !subtitleItalic {
		subtitleItalic = true // default: italic
	}
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Subtitle", name: "Subtitle",
		font: t.Title.SubtitleFontFamily,
		size: ptToHalfPt(t.Title.SubtitleFontSize), bold: subtitleBold, italic: subtitleItalic,
		color: t.Title.SubtitleFontColor,
	}))

	// Heading styles (1-9)
	for i := 1; i <= 9; i++ {
		caps := t.headingTextTransform(i) == "uppercase"
		hBold, hItalic := t.headingFontStyle(i)
		hColor := t.headingFontColor(i)
		opts := styleParaOpts{
			id: "Heading" + strconv.Itoa(i), name: "heading " + strconv.Itoa(i),
			font: headingFont, size: t.headingSizeHalfPt(i),
			bold: hBold, italic: hItalic, caps: caps, color: hColor,
			outlineLevel: i - 1,
			keepNext:     true,
		}
		if mt := t.headingMarginTop(i); mt > 0 {
			opts.spaceBefore = ptToTwips(mt)
		}
		if mb := t.headingMarginBottom(i); mb > 0 {
			opts.spaceAfter = ptToTwips(mb)
		}
		b.WriteString(styleParaXML(opts))
	}

	// Quote style
	qBold, qItalic := fontStyleBoldItalic(t.Quote.FontStyle)
	if !qBold && !qItalic && t.Quote.FontStyle == "" {
		qItalic = true // default: italic
	}
	qSize := ptToHalfPt(t.Base.FontSize)
	if t.Quote.FontSize > 0 {
		qSize = ptToHalfPt(t.Quote.FontSize)
	}
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Quote", name: "Quote",
		font: t.Quote.FontFamily, size: qSize,
		bold: qBold, italic: qItalic, color: t.Quote.FontColor,
		borderLeft: t.Quote.BorderColor, borderLeftWidth: t.Quote.BorderWidth,
	}))

	// Admonition style
	admSize := ptToHalfPt(t.Base.FontSize)
	if t.Admonition.FontSize > 0 {
		admSize = ptToHalfPt(t.Admonition.FontSize)
	}
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Admonition", name: "Admonition",
		size: admSize, color: t.Admonition.FontColor,
		shading:   t.Admonition.BackgroundColor,
		borderAll: t.Admonition.BorderColor, borderAllWidth: t.Admonition.BorderWidth,
	}))

	// Caption style
	capSize := ptToHalfPt(t.Code.FontSize) // default: same as code size
	if t.Caption.FontSize > 0 {
		capSize = ptToHalfPt(t.Caption.FontSize)
	}
	capBold, capItalic := fontStyleBoldItalic(t.Caption.FontStyle)
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Caption", name: "Caption",
		font: t.Caption.FontFamily, size: capSize,
		bold: capBold, italic: capItalic, color: t.Caption.FontColor,
		align: t.Caption.TextAlign,
	}))

	// CodeBlock style
	codeBgColor := t.Code.BackgroundColor
	codeOpts := styleParaOpts{
		id: "CodeBlock", name: "Code Block",
		font: t.Code.FontFamily, size: ptToHalfPt(t.Code.FontSize),
		color: t.Code.FontColor, shading: codeBgColor,
	}
	if t.Code.BorderColor != "" {
		codeOpts.borderAll = t.Code.BorderColor
		codeOpts.borderAllWidth = t.Code.BorderWidth
	}
	if t.Code.LineHeight > 0 {
		codeOpts.lineSpacing = int(t.Code.LineHeight * 240)
	}
	b.WriteString(styleParaXML(codeOpts))

	// Sidebar style
	sidebarSize := ptToHalfPt(t.Base.FontSize)
	if t.Sidebar.FontSize > 0 {
		sidebarSize = ptToHalfPt(t.Sidebar.FontSize)
	}
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Sidebar", name: "Sidebar",
		size: sidebarSize, color: t.Sidebar.FontColor,
		shading:   t.Sidebar.BackgroundColor,
		borderAll: t.Sidebar.BorderColor, borderAllWidth: t.Sidebar.BorderWidth,
	}))

	// Example style
	exampleSize := ptToHalfPt(t.Base.FontSize)
	if t.Example.FontSize > 0 {
		exampleSize = ptToHalfPt(t.Example.FontSize)
	}
	b.WriteString(styleParaXML(styleParaOpts{
		id: "Example", name: "Example",
		size: exampleSize, color: t.Example.FontColor,
		shading:   t.Example.BackgroundColor,
		borderAll: t.Example.BorderColor, borderAllWidth: t.Example.BorderWidth,
	}))

	// Remaining styles
	b.WriteString(styleParaXML(styleParaOpts{
		id: "ListParagraph", name: "List Paragraph",
		size: ptToHalfPt(t.Base.FontSize),
		// Canonical Word ListParagraph definition: base indent + suppress inter-item spacing.
		indentLeft:        ptToTwips(t.List.Indent),
		contextualSpacing: true,
	}))
	// TOC entry styles: level 1 = bold, level 2 = indented, level 3 = further indented + smaller
	tocIndent := ptToTwips(t.List.Indent)
	b.WriteString(styleParaXML(styleParaOpts{
		id: "TOCEntry1", name: "TOC Entry 1",
		size: ptToHalfPt(t.Base.FontSize), bold: true,
	}))
	b.WriteString(styleParaXML(styleParaOpts{
		id: "TOCEntry2", name: "TOC Entry 2",
		size: ptToHalfPt(t.Base.FontSize), indentLeft: tocIndent,
	}))
	b.WriteString(styleParaXML(styleParaOpts{
		id: "TOCEntry3", name: "TOC Entry 3",
		size: ptToHalfPt(t.Base.FontSize) - 2, indentLeft: tocIndent * 2,
	}))
	b.WriteString(styleParaXML(styleParaOpts{id: "FootnoteText", name: "Footnote Text", size: 18}))

	// Hyperlink character style
	linkBold, linkItalic := fontStyleBoldItalic(t.Link.FontStyle)
	charStyle{
		id: "Hyperlink", name: "Hyperlink",
		color: t.Link.FontColor, bold: linkBold, italic: linkItalic,
		underline: t.Link.TextDecoration != "none",
	}.xml(b)

	charStyle{id: "FootnoteReference", name: "Footnote Reference", vertAlign: "superscript"}.xml(b)
	b.WriteString(`</w:styles>`)
	return b.String()
}

// styleParaOpts configures a paragraph style definition.
type styleParaOpts struct {
	id                string
	name              string
	font              string
	size              int
	bold              bool
	italic            bool
	caps              bool
	color             string
	shading           string  // background color (w:shd fill)
	borderLeft        string  // left border color
	borderLeftWidth   float64 // left border width in pt
	borderAll         string  // all-side border color
	borderAllWidth    float64 // all-side border width in pt
	spaceBefore       int     // w:spacing w:before (twips)
	spaceAfter        int     // w:spacing w:after (twips)
	lineSpacing       int     // w:spacing w:line (240 = single)
	align             string  // text alignment
	indentLeft        int     // w:ind w:left (twips)
	outlineLevel      int     // -1 = not set; 0–8 = Word outline level (w:outlineLvl)
	keepNext          bool    // w:keepNext: keep paragraph on same page as next
	contextualSpacing bool    // w:contextualSpacing: suppress spacing between same-style paragraphs
}

func styleParaXML(opts styleParaOpts) string {
	b := &strings.Builder{}
	b.WriteString(`<w:style w:type="paragraph" w:styleId="`)
	b.WriteString(xmlAttr(opts.id))
	b.WriteString(`"><w:name w:val="`)
	b.WriteString(xmlAttr(opts.name))
	b.WriteString(`"/>`)
	opts.writePPr(b)
	opts.writeRPr(b)
	b.WriteString(`</w:style>`)
	return b.String()
}

func (opts styleParaOpts) writePPr(b *strings.Builder) {
	hasPPr := opts.spaceBefore > 0 || opts.spaceAfter > 0 || opts.lineSpacing > 0 ||
		opts.shading != "" || opts.borderLeft != "" || opts.borderAll != "" || opts.align != "" || opts.indentLeft > 0 ||
		opts.keepNext || opts.contextualSpacing || opts.outlineLevel >= 0
	if !hasPPr {
		return
	}
	b.WriteString(`<w:pPr>`)
	if opts.keepNext {
		b.WriteString(`<w:keepNext/>`)
	}
	if opts.contextualSpacing {
		b.WriteString(`<w:contextualSpacing/>`)
	}
	if opts.spaceBefore > 0 || opts.spaceAfter > 0 || opts.lineSpacing > 0 {
		b.WriteString(`<w:spacing`)
		if opts.spaceBefore > 0 {
			b.WriteString(` w:before="`)
			b.WriteString(itoa(opts.spaceBefore))
			b.WriteString(`"`)
		}
		if opts.spaceAfter > 0 {
			b.WriteString(` w:after="`)
			b.WriteString(itoa(opts.spaceAfter))
			b.WriteString(`"`)
		}
		if opts.lineSpacing > 0 {
			b.WriteString(` w:line="`)
			b.WriteString(itoa(opts.lineSpacing))
			b.WriteString(`" w:lineRule="auto"`)
		}
		b.WriteString(`/>`)
	}
	writeShading(b, opts.shading)
	if opts.borderAll != "" {
		bw := opts.borderAllWidth
		if ptToEighths(bw) < 1 {
			bw = 0.5 // default 0.5pt = 4 eighths
		}
		bl := borderLine{sizePt: bw, space: 4, color: opts.borderAll}
		b.WriteString(`<w:pBdr>`)
		b.WriteString(`<w:top `)
		bl.writeAttrs(b)
		b.WriteString(`/><w:left `)
		bl.writeAttrs(b)
		b.WriteString(`/><w:bottom `)
		bl.writeAttrs(b)
		b.WriteString(`/><w:right `)
		bl.writeAttrs(b)
		b.WriteString(`/>`)
		b.WriteString(`</w:pBdr>`)
	} else if opts.borderLeft != "" {
		bw := opts.borderLeftWidth
		if ptToEighths(bw) < 1 {
			bw = 0.5
		}
		bl := borderLine{sizePt: bw, space: 4, color: opts.borderLeft}
		b.WriteString(`<w:pBdr><w:left `)
		bl.writeAttrs(b)
		b.WriteString(`/></w:pBdr>`)
	}
	if opts.indentLeft > 0 {
		b.WriteString(`<w:ind w:left="`)
		b.WriteString(strconv.Itoa(opts.indentLeft))
		b.WriteString(`"/>`)
	}
	if opts.align != "" {
		b.WriteString(`<w:jc w:val="`)
		b.WriteString(xmlAttr(ooxmlAlignment(opts.align)))
		b.WriteString(`"/>`)
	}
	if opts.outlineLevel >= 0 {
		b.WriteString(`<w:outlineLvl w:val="`)
		b.WriteString(itoa(opts.outlineLevel))
		b.WriteString(`"/>`)
	}
	b.WriteString(`</w:pPr>`)
}

func (opts styleParaOpts) writeRPr(b *strings.Builder) {
	b.WriteString(`<w:rPr>`)
	if opts.font != "" {
		b.WriteString(`<w:rFonts w:ascii="`)
		b.WriteString(xmlAttr(opts.font))
		b.WriteString(`" w:hAnsi="`)
		b.WriteString(xmlAttr(opts.font))
		b.WriteString(`" w:cs="`)
		b.WriteString(xmlAttr(opts.font))
		b.WriteString(`"/>`)
	}
	if opts.bold {
		b.WriteString(`<w:b/>`)
	}
	if opts.italic {
		b.WriteString(`<w:i/>`)
	}
	if opts.caps {
		b.WriteString(`<w:caps/>`)
	}
	if opts.color != "" {
		b.WriteString(`<w:color w:val="`)
		b.WriteString(xmlAttr(opts.color))
		b.WriteString(`"/>`)
	}
	b.WriteString(`<w:sz w:val="`)
	b.WriteString(strconv.Itoa(opts.size))
	b.WriteString(`"/><w:szCs w:val="`)
	b.WriteString(strconv.Itoa(opts.size))
	b.WriteString(`"/></w:rPr>`)
}

// ooxmlAlignment maps theme alignment values to OOXML w:jc values.
func ooxmlAlignment(align string) string {
	switch align {
	case "justify":
		return "both"
	case "left", "center", "right":
		return align
	default:
		return "left"
	}
}

// resolveTextAlign returns the first non-empty alignment value.
func resolveTextAlign(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func xmlHeader() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
}

var (
	xmlTextReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	xmlAttrReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
)

func xmlText(s string) string {
	return xmlTextReplacer.Replace(s)
}

func xmlAttr(s string) string {
	return xmlAttrReplacer.Replace(s)
}

func settingsXML() string {
	return xmlHeader() +
		`<w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:defaultTabStop w:val="720"/>` +
		`<w:compat>` +
		`<w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/>` +
		`</w:compat>` +
		`</w:settings>`
}

func (d *docxDocument) fontTableXML() string {
	type fontEntry struct {
		name, family, pitch string
	}
	// Collect fonts in precedence order; first occurrence wins on dedup.
	entries := []fontEntry{
		{name: d.theme.Base.FontFamily, family: "swiss", pitch: "variable"},
		{name: d.theme.Code.FontFamily, family: "modern", pitch: "fixed"},
	}
	if d.theme.Heading.FontFamily != "" && d.theme.Heading.FontFamily != d.theme.Base.FontFamily {
		entries = append(entries, fontEntry{name: d.theme.Heading.FontFamily, family: "swiss", pitch: "variable"})
	}
	entries = append(entries, fontEntry{name: "Symbol", family: "auto", pitch: "default"})

	// Deduplicate by name, keeping first occurrence.
	seen := make(map[string]bool)
	var unique []fontEntry
	for _, e := range entries {
		if e.name == "" || seen[e.name] {
			continue
		}
		seen[e.name] = true
		unique = append(unique, e)
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i].name < unique[j].name })

	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<w:fonts xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	for _, f := range unique {
		b.WriteString(`<w:font w:name="`)
		b.WriteString(xmlAttr(f.name))
		b.WriteString(`"><w:charset w:val="00"/><w:family w:val="`)
		b.WriteString(xmlAttr(f.family))
		b.WriteString(`"/><w:pitch w:val="`)
		b.WriteString(xmlAttr(f.pitch))
		b.WriteString(`"/></w:font>`)
	}
	b.WriteString(`</w:fonts>`)
	return b.String()
}

func (d *docxDocument) corePropertiesXML() string {
	b := &strings.Builder{}
	b.WriteString(xmlHeader())
	b.WriteString(`<cp:coreProperties`)
	b.WriteString(` xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"`)
	b.WriteString(` xmlns:dc="http://purl.org/dc/elements/1.1/"`)
	b.WriteString(` xmlns:dcterms="http://purl.org/dc/terms/"`)
	b.WriteString(` xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">`)
	if d.title != "" {
		b.WriteString(`<dc:title>`)
		b.WriteString(xmlText(d.title))
		b.WriteString(`</dc:title>`)
	}
	if d.creators != "" {
		b.WriteString(`<dc:creator>`)
		b.WriteString(xmlText(d.creators))
		b.WriteString(`</dc:creator>`)
	}
	if !d.created.IsZero() {
		b.WriteString(`<dcterms:created xsi:type="dcterms:W3CDTF">`)
		b.WriteString(d.created.UTC().Format("2006-01-02T15:04:05Z"))
		b.WriteString(`</dcterms:created>`)
	}
	if !d.modified.IsZero() {
		b.WriteString(`<dcterms:modified xsi:type="dcterms:W3CDTF">`)
		b.WriteString(d.modified.UTC().Format("2006-01-02T15:04:05Z"))
		b.WriteString(`</dcterms:modified>`)
	}
	b.WriteString(`</cp:coreProperties>`)
	return b.String()
}

func appPropertiesXML() string {
	return xmlHeader() +
		`<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">` +
		`<Application>libasciidoc</Application>` +
		`</Properties>`
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
