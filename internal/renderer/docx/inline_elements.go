package docx

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

// runStyle carries accumulated formatting state for nested inline elements.
type runStyle struct {
	bold          bool
	italic        bool
	monospace     bool
	monoFont      string // font name for monospace (from theme); empty means "Courier New"
	font          string // explicit font override (non-monospace)
	highlight     bool
	subscript     bool
	superscript   bool
	underline     bool
	color         string
	charStyle     string
	shading       string // background color for inline shading (w:shd)
}

type paragraphOptions struct {
	style        string
	numID        int
	level        int
	bookmarkName string
	indentLeft   int // left indent in twips (0 = no indent)
}

func (r *docxRenderer) renderInlineElements(para *strings.Builder, elements []interface{}, style runStyle) error {
	for _, element := range elements {
		if err := r.renderInlineElement(para, element, style); err != nil {
			return err
		}
	}
	return nil
}

func (r *docxRenderer) renderInlineElement(para *strings.Builder, element interface{}, style runStyle) error {
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

func (r *docxRenderer) startParagraph(opts paragraphOptions) *strings.Builder {
	para := &strings.Builder{}
	para.WriteString("<w:p>")
	writeParagraphProperties(para, opts)
	if opts.bookmarkName != "" {
		id := r.doc.nextBookmarkID()
		para.WriteString(`<w:bookmarkStart w:id="`)
		para.WriteString(fmt.Sprint(id))
		para.WriteString(`" w:name="`)
		para.WriteString(xmlAttr(sanitizeBookmarkName(opts.bookmarkName)))
		para.WriteString(`"/>`)
		para.WriteString(`<w:bookmarkEnd w:id="`)
		para.WriteString(fmt.Sprint(id))
		para.WriteString(`"/>`)
	}
	return para
}

func (r *docxRenderer) endParagraph(para *strings.Builder) {
	para.WriteString("</w:p>")
	r.writer.WriteString(para.String())
}

func writeParagraphProperties(para *strings.Builder, opts paragraphOptions) {
	if opts.style == "" && opts.numID == 0 && opts.indentLeft == 0 {
		return
	}
	para.WriteString("<w:pPr>")
	if opts.style != "" {
		para.WriteString(`<w:pStyle w:val="`)
		para.WriteString(xmlAttr(opts.style))
		para.WriteString(`"/>`)
	}
	if opts.numID > 0 {
		para.WriteString(`<w:numPr><w:ilvl w:val="`)
		para.WriteString(fmt.Sprint(opts.level))
		para.WriteString(`"/><w:numId w:val="`)
		para.WriteString(fmt.Sprint(opts.numID))
		para.WriteString(`"/></w:numPr>`)
	}
	if opts.indentLeft > 0 {
		para.WriteString(`<w:ind w:left="`)
		para.WriteString(fmt.Sprint(opts.indentLeft))
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

func (r *docxRenderer) writeTextRun(para *strings.Builder, text string, style runStyle) {
	if text == "" {
		return
	}
	para.WriteString("<w:r>")
	writeRunProperties(para, style)
	writeRunTextChildren(para, text)
	para.WriteString("</w:r>")
}

// writeRunProperties emits w:rPr children in the order required by
// ECMA-376 CT_RPr (§17.3.2.28): rStyle, rFonts, b, i, caps, …,
// color, …, highlight, u, …, vertAlign.
func writeRunProperties(para *strings.Builder, style runStyle) {
	if !style.bold && !style.italic && !style.monospace && !style.highlight && !style.subscript && !style.superscript && !style.underline && style.color == "" && style.charStyle == "" && style.font == "" && style.shading == "" {
		return
	}
	para.WriteString("<w:rPr>")
	// 1. rStyle
	if style.charStyle != "" {
		para.WriteString(`<w:rStyle w:val="`)
		para.WriteString(xmlAttr(style.charStyle))
		para.WriteString(`"/>`)
	}
	// 2. rFonts (monospace takes precedence over explicit font)
	if style.monospace {
		font := style.monoFont
		if font == "" {
			font = "Courier New"
		}
		para.WriteString(`<w:rFonts w:ascii="`)
		para.WriteString(xmlAttr(font))
		para.WriteString(`" w:hAnsi="`)
		para.WriteString(xmlAttr(font))
		para.WriteString(`" w:cs="`)
		para.WriteString(xmlAttr(font))
		para.WriteString(`"/>`)
	} else if style.font != "" {
		para.WriteString(`<w:rFonts w:ascii="`)
		para.WriteString(xmlAttr(style.font))
		para.WriteString(`" w:hAnsi="`)
		para.WriteString(xmlAttr(style.font))
		para.WriteString(`" w:cs="`)
		para.WriteString(xmlAttr(style.font))
		para.WriteString(`"/>`)
	}
	// 3. b
	if style.bold {
		para.WriteString("<w:b/>")
	}
	// 4. i
	if style.italic {
		para.WriteString("<w:i/>")
	}
	// 5. color
	if style.color != "" {
		para.WriteString(`<w:color w:val="`)
		para.WriteString(xmlAttr(style.color))
		para.WriteString(`"/>`)
	}
	// 6. shading (inline background)
	if style.shading != "" {
		para.WriteString(`<w:shd w:val="clear" w:color="auto" w:fill="`)
		para.WriteString(xmlAttr(style.shading))
		para.WriteString(`"/>`)
	}
	// 7. highlight
	if style.highlight {
		para.WriteString(`<w:highlight w:val="yellow"/>`)
	}
	// 8. u
	if style.underline {
		para.WriteString(`<w:u w:val="single"/>`)
	}
	// 9. vertAlign
	if style.subscript {
		para.WriteString(`<w:vertAlign w:val="subscript"/>`)
	}
	if style.superscript {
		para.WriteString(`<w:vertAlign w:val="superscript"/>`)
	}
	para.WriteString("</w:rPr>")
}

func writeRunTextChildren(para *strings.Builder, text string) {
	if text == "" {
		return
	}
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			para.WriteString("<w:br/>")
		}
		for j, part := range strings.Split(line, "\t") {
			if j > 0 {
				para.WriteString("<w:tab/>")
			}
			if part == "" {
				continue
			}
			para.WriteString(`<w:t xml:space="preserve">`)
			para.WriteString(xmlText(part))
			para.WriteString("</w:t>")
		}
	}
}
