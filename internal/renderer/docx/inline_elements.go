package docx

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

// paragraphBuilder accumulates paragraph XML with deferred run serialisation.
// Text added via appendText is held in a pending run and merged as long as the
// runStyle is identical.  Any call to WriteString or Write (structural XML such
// as pPr, bookmarks, hyperlinks, or line breaks) automatically flushes the
// pending run first, maintaining correct element ordering without requiring
// callers to manage state.
type paragraphBuilder struct {
	xml          strings.Builder // serialised XML output
	pendingText  strings.Builder // accumulated text for the pending run
	pendingStyle runStyle        // style of the currently buffered run
	hasPending   bool
}

// WriteString flushes the pending run and appends raw XML.
// All non-text content (pPr, bookmarks, hyperlinks, br, etc.) uses this path.
// Returns nothing because strings.Builder.WriteString never errors.
func (pb *paragraphBuilder) WriteString(s string) {
	pb.flushPendingRun()
	pb.xml.WriteString(s)
}

// Write implements io.Writer so that fmt.Fprint can write to the builder.
func (pb *paragraphBuilder) Write(p []byte) (int, error) {
	pb.flushPendingRun()
	return pb.xml.Write(p)
}

// appendText buffers text into the pending run.
// If the incoming style matches the current pending style the text is merged;
// otherwise the pending run is flushed and a new one is started.
func (pb *paragraphBuilder) appendText(text string, style runStyle) {
	if text == "" {
		return
	}
	if pb.hasPending && pb.pendingStyle == style {
		pb.pendingText.WriteString(text)
		return
	}
	pb.flushPendingRun()
	pb.pendingStyle = style
	pb.pendingText.Reset()
	pb.pendingText.WriteString(text)
	pb.hasPending = true
}

func (pb *paragraphBuilder) flushPendingRun() {
	if !pb.hasPending {
		return
	}
	text := pb.pendingText.String()
	pb.hasPending = false
	pb.pendingText.Reset()
	if text == "" {
		return
	}
	pb.xml.WriteString("<w:r>")
	pb.pendingStyle.writeRPr(&pb.xml)
	writeRunTextChildren(&pb.xml, text)
	pb.xml.WriteString("</w:r>")
}

// String flushes any pending run and returns the full serialised XML.
func (pb *paragraphBuilder) String() string {
	pb.flushPendingRun()
	return pb.xml.String()
}

func (pb *paragraphBuilder) Reset() {
	pb.xml.Reset()
	pb.pendingText.Reset()
	pb.hasPending = false
	pb.pendingStyle = runStyle{}
}

var paragraphBuilderPool = sync.Pool{
	New: func() interface{} {
		pb := &paragraphBuilder{}
		pb.xml.Grow(256)
		pb.pendingText.Grow(64)
		return pb
	},
}

// runStyle carries accumulated formatting state for nested inline elements.
type runStyle struct {
	bold        bool
	italic      bool
	monospace   bool
	monoFont    string // font name for monospace (from theme); empty means "Courier New"
	font        string // explicit font override (non-monospace)
	highlight   bool
	subscript   bool
	superscript bool
	underline   bool
	color       string
	charStyle   string
	shading     string // background color for inline shading (w:shd)
}

type paragraphOptions struct {
	style        string
	numID        int
	level        int
	bookmarkName string
	indentLeft   int  // left indent in twips (0 = no indent)
	keepNext     bool // emit <w:keepNext/> to keep this paragraph with the next
}

func (r *docxRenderer) renderInlineElements(para *paragraphBuilder, elements []interface{}, style runStyle) error {
	for _, element := range elements {
		if err := r.renderInlineElement(para, element, style); err != nil {
			return err
		}
	}
	return nil
}

func (r *docxRenderer) renderInlineElement(para *paragraphBuilder, element interface{}, style runStyle) error {
	switch e := element.(type) {
	case *types.StringElement:
		return r.renderStringElement(para, e, style)
	case *types.QuotedText:
		return r.renderQuotedText(para, e, style)
	case *types.InlineLink:
		return r.renderLink(para, e)
	case *types.InlineImage:
		return r.renderInlineImage(para, e)
	case *types.ExternalCrossReference:
		return r.renderExternalCrossReference(para, e, style)
	case *types.LineBreak:
		r.renderLineBreak(para)
		return nil
	case *types.SpecialCharacter:
		return r.renderSpecialCharacter(para, e, style)
	case *types.Symbol:
		return r.renderSymbol(para, e, style)
	case *types.PredefinedAttribute:
		return r.renderPredefinedAttribute(para, e, style)
	case *types.InternalCrossReference:
		return r.renderCrossReference(para, e, style)
	case *types.FootnoteReference:
		return r.renderFootnoteRef(para, e, style)
	case *types.InlinePassthrough:
		return r.renderInlinePassthrough(para, e, style)
	case *types.InlineButton:
		label := e.Attributes.GetAsStringWithDefault(types.AttrButtonLabel, "")
		r.writeTextRun(para, label, style)
		return nil
	case *types.InlineMenu:
		r.writeTextRun(para, strings.Join(e.Path, " > "), style)
		return nil
	case *types.Icon:
		label := e.Attributes.GetAsStringWithDefault(types.AttrImageAlt, e.Class)
		r.writeTextRun(para, label, style)
		return nil
	case *types.UserMacro:
		return r.renderUserMacroInline(para, e, style)
	case *types.IndexTerm:
		return r.renderInlineElements(para, e.Term, style)
	case *types.ConcealedIndexTerm:
		return nil
	case *types.Callout:
		r.writeTextRun(para, fmt.Sprintf("<%d>", e.Ref), style)
		return nil
	case *types.AttributeDeclaration:
		r.ctx.attributes[e.Name] = e.Value
		return nil
	case *types.AttributeReset:
		delete(r.ctx.attributes, e.Name)
		return nil
	default:
		return fmt.Errorf("docx: unsupported inline element type: %T", element)
	}
}

func (r *docxRenderer) startParagraph(opts paragraphOptions) *paragraphBuilder {
	para := paragraphBuilderPool.Get().(*paragraphBuilder)
	para.Reset()
	para.WriteString("<w:p>")
	opts.writePPr(para)
	if opts.bookmarkName != "" {
		id := r.doc.nextBookmarkID()
		name := r.doc.uniqueBookmarkName(opts.bookmarkName)
		para.WriteString(`<w:bookmarkStart w:id="`)
		fmt.Fprint(para, id)
		para.WriteString(`" w:name="`)
		para.WriteString(xmlAttr(name))
		para.WriteString(`"/>`)
		para.WriteString(`<w:bookmarkEnd w:id="`)
		fmt.Fprint(para, id)
		para.WriteString(`"/>`)
	}
	return para
}

func (r *docxRenderer) endParagraph(para *paragraphBuilder) {
	para.WriteString("</w:p>")
	r.writer.WriteString(para.String())
	paragraphBuilderPool.Put(para)
}

func (opts paragraphOptions) writePPr(para *paragraphBuilder) {
	if opts.style == "" && opts.numID == 0 && opts.indentLeft == 0 && !opts.keepNext {
		return
	}
	para.WriteString("<w:pPr>")
	if opts.style != "" {
		para.WriteString(`<w:pStyle w:val="`)
		para.WriteString(xmlAttr(opts.style))
		para.WriteString(`"/>`)
	}
	if opts.keepNext {
		para.WriteString(`<w:keepNext/>`)
	}
	if opts.numID > 0 {
		para.WriteString(`<w:numPr><w:ilvl w:val="`)
		fmt.Fprint(para, opts.level)
		para.WriteString(`"/><w:numId w:val="`)
		fmt.Fprint(para, opts.numID)
		para.WriteString(`"/></w:numPr>`)
	}
	if opts.indentLeft > 0 {
		para.WriteString(`<w:ind w:left="`)
		fmt.Fprint(para, opts.indentLeft)
		para.WriteString(`"/>`)
	}
	para.WriteString("</w:pPr>")
}

func (r *docxRenderer) renderTextParagraph(text string, opts paragraphOptions) error {
	para := r.startParagraph(opts)
	r.writeTextRun(para, text, runStyle{})
	r.endParagraph(para)
	return nil
}

// writeTextRun buffers text into the paragraph's pending run.
// Adjacent calls with the same runStyle are automatically merged into a single
// <w:r> by the paragraphBuilder's deferred serialisation.
func (r *docxRenderer) writeTextRun(para *paragraphBuilder, text string, style runStyle) {
	para.appendText(text, style)
}

// writeRPr emits w:rPr children in the order required by
// ECMA-376 CT_RPr (§17.3.2.28): rStyle, rFonts, b, i, caps, …,
// color, …, highlight, u, …, vertAlign.
func (style runStyle) writeRPr(b *strings.Builder) {
	if !style.bold && !style.italic && !style.monospace && !style.highlight && !style.subscript && !style.superscript && !style.underline && style.color == "" && style.charStyle == "" && style.font == "" && style.shading == "" {
		return
	}
	b.WriteString("<w:rPr>")
	// 1. rStyle
	if style.charStyle != "" {
		b.WriteString(`<w:rStyle w:val="`)
		b.WriteString(xmlAttr(style.charStyle))
		b.WriteString(`"/>`)
	}
	// 2. rFonts (monospace takes precedence over explicit font)
	if style.monospace {
		font := style.monoFont
		if font == "" {
			font = "Courier New"
		}
		b.WriteString(`<w:rFonts w:ascii="`)
		b.WriteString(xmlAttr(font))
		b.WriteString(`" w:hAnsi="`)
		b.WriteString(xmlAttr(font))
		b.WriteString(`" w:cs="`)
		b.WriteString(xmlAttr(font))
		b.WriteString(`"/>`)
	} else if style.font != "" {
		b.WriteString(`<w:rFonts w:ascii="`)
		b.WriteString(xmlAttr(style.font))
		b.WriteString(`" w:hAnsi="`)
		b.WriteString(xmlAttr(style.font))
		b.WriteString(`" w:cs="`)
		b.WriteString(xmlAttr(style.font))
		b.WriteString(`"/>`)
	}
	// 3. b
	if style.bold {
		b.WriteString("<w:b/>")
	}
	// 4. i
	if style.italic {
		b.WriteString("<w:i/>")
	}
	// 5. color
	if style.color != "" {
		b.WriteString(`<w:color w:val="`)
		b.WriteString(xmlAttr(style.color))
		b.WriteString(`"/>`)
	}
	// 6. shading (inline background)
	if style.shading != "" {
		b.WriteString(`<w:shd w:val="clear" w:color="auto" w:fill="`)
		b.WriteString(xmlAttr(style.shading))
		b.WriteString(`"/>`)
	}
	// 7. highlight
	if style.highlight {
		b.WriteString(`<w:highlight w:val="yellow"/>`)
	}
	// 8. u
	if style.underline {
		b.WriteString(`<w:u w:val="single"/>`)
	}
	// 9. vertAlign
	if style.subscript {
		b.WriteString(`<w:vertAlign w:val="subscript"/>`)
	}
	if style.superscript {
		b.WriteString(`<w:vertAlign w:val="superscript"/>`)
	}
	b.WriteString("</w:rPr>")
}

func writeRunTextChildren(b *strings.Builder, text string) {
	if text == "" {
		return
	}
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			b.WriteString("<w:br/>")
		}
		for j, part := range strings.Split(line, "\t") {
			if j > 0 {
				b.WriteString("<w:tab/>")
			}
			if part == "" {
				continue
			}
			b.WriteString(`<w:t xml:space="preserve">`)
			b.WriteString(xmlText(part))
			b.WriteString("</w:t>")
		}
	}
}
