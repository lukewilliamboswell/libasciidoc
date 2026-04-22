package docx

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderElements(elements []interface{}) error {
	for _, element := range elements {
		if err := r.renderElement(element); err != nil {
			return err
		}
	}
	return nil
}

func (r *docxRenderer) renderElement(element interface{}) error {
	switch e := element.(type) {
	case *types.TableOfContents:
		return r.renderTableOfContents(e)
	case *types.Section:
		return r.renderSection(e)
	case *types.Preamble:
		if err := r.renderElements(e.Elements); err != nil {
			return err
		}
		return r.renderTableOfContents(e.TableOfContents)
	case *types.Paragraph:
		return r.renderParagraph(e)
	case *types.List:
		return r.renderList(e)
	case *types.Table:
		return r.renderTable(e)
	case *types.ImageBlock:
		return r.renderImageBlock(e)
	case *types.DelimitedBlock:
		return r.renderDelimitedBlock(e)
	case *types.ThematicBreak:
		r.renderThematicBreak()
		return nil
	case *types.PageBreak:
		r.writer.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
		return nil
	case *types.InternalCrossReference, *types.ExternalCrossReference, *types.QuotedText,
		*types.InlinePassthrough, *types.InlineButton, *types.InlineImage,
		*types.InlineLink, *types.InlineMenu, *types.Icon, *types.StringElement,
		*types.FootnoteReference, *types.LineBreak, *types.UserMacro,
		*types.IndexTerm, *types.ConcealedIndexTerm, *types.Callout,
		*types.SpecialCharacter, *types.Symbol, *types.PredefinedAttribute:
		para := r.startParagraph(paragraphOptions{})
		if err := r.renderInlineElement(para, e, runStyle{}); err != nil {
			return err
		}
		r.endParagraph(para)
		return nil
	case *types.AttributeDeclaration:
		r.ctx.attributes[e.Name] = e.Value
		return nil
	case *types.AttributeReset:
		delete(r.ctx.attributes, e.Name)
		return nil
	case *types.FrontMatter:
		r.ctx.attributes.AddAll(e.Attributes)
		return nil
	default:
		return fmt.Errorf("docx: unsupported block element type: %T", element)
	}
}

func (r *docxRenderer) renderThematicBreak() {
	r.writer.WriteString(`<w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="6" w:space="1" w:color="auto"/></w:pBdr></w:pPr></w:p>`)
}

func (r *docxRenderer) renderDelimitedBlock(b *types.DelimitedBlock) error {
	if title := b.Attributes.GetAsStringWithDefault(types.AttrTitle, ""); title != "" {
		if err := r.renderTextParagraph(title, paragraphOptions{style: "Caption"}); err != nil {
			return err
		}
	}
	switch b.Kind {
	case types.Comment:
		return nil
	case types.Example:
		if style, _ := b.Attributes.GetAsString(types.AttrStyle); isAdmonitionStyle(style) {
			return r.renderAdmonitionBlock(b, style)
		}
		return r.renderStyledBlock(b, "Example")
	case types.Listing, types.Literal, types.Source, types.Fenced, types.MarkdownCode:
		text, err := r.renderPlainText(b.Elements)
		if err != nil {
			return err
		}
		for _, line := range splitPreserveOne(text) {
			if err := r.renderTextParagraph(line, paragraphOptions{style: "CodeBlock"}); err != nil {
				return err
			}
		}
		return nil
	case types.Quote, types.MarkdownQuote, types.Verse:
		old := r.writer
		tmp := &strings.Builder{}
		r.writer = tmp
		err := r.renderElements(b.Elements)
		r.writer = old
		if err != nil {
			return err
		}
		r.writer.WriteString(strings.ReplaceAll(tmp.String(), `<w:p>`, `<w:p><w:pPr><w:pStyle w:val="Quote"/></w:pPr>`))
		if author := b.Attributes.GetAsStringWithDefault(types.AttrQuoteAuthor, ""); author != "" {
			title := b.Attributes.GetAsStringWithDefault(types.AttrQuoteTitle, "")
			attribution := "— " + author
			if title != "" {
				attribution += ", " + title
			}
			return r.renderTextParagraph(attribution, paragraphOptions{style: "Caption"})
		}
		return nil
	case types.Sidebar:
		return r.renderStyledBlock(b, "Sidebar")
	default:
		return r.renderElements(b.Elements)
	}
}

// renderStyledBlock renders a delimited block by applying a paragraph style to all child paragraphs.
func (r *docxRenderer) renderStyledBlock(b *types.DelimitedBlock, style string) error {
	old := r.writer
	tmp := &strings.Builder{}
	r.writer = tmp
	err := r.renderElements(b.Elements)
	r.writer = old
	if err != nil {
		return err
	}
	// Inject style into all paragraphs
	r.writer.WriteString(strings.ReplaceAll(tmp.String(), `<w:p>`, `<w:p><w:pPr><w:pStyle w:val="`+style+`"/></w:pPr>`))
	return nil
}

func splitPreserveOne(text string) []string {
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return []string{""}
	}
	return strings.Split(text, "\n")
}
