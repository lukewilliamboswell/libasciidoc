package docx

import (
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderParagraph(p *types.Paragraph) error {
	style, _ := p.Attributes.GetAsString(types.AttrStyle)

	switch style {
	case types.Tip, types.Note, types.Important, types.Warning, types.Caution:
		return r.renderAdmonitionParagraph(p, style)
	case types.LiteralParagraph, types.Source, types.Literal:
		return r.renderCodeParagraph(p)
	default:
		return r.renderRegularParagraph(p)
	}
}

func (r *docxRenderer) renderRegularParagraph(p *types.Paragraph) error {
	para := r.startParagraph(paragraphOptions{})
	if err := r.renderCheckPrefix(para, p); err != nil {
		return err
	}
	if err := r.renderInlineElements(para, p.Elements, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

func (r *docxRenderer) renderCodeParagraph(p *types.Paragraph) error {
	para := r.startParagraph(paragraphOptions{style: "CodeBlock"})
	if err := r.renderInlineElements(para, p.Elements, runStyle{monospace: true}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

func (r *docxRenderer) renderParagraphAsListItem(p *types.Paragraph, opts paragraphOptions) error {
	if opts.style == "" {
		opts.style = "ListParagraph"
	}
	para := r.startParagraph(opts)
	if err := r.renderCheckPrefix(para, p); err != nil {
		return err
	}
	if err := r.renderInlineElements(para, p.Elements, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

func (r *docxRenderer) renderCheckPrefix(para *strings.Builder, p *types.Paragraph) error {
	switch p.Attributes[types.AttrCheckStyle] {
	case types.Checked, types.CheckedInteractive:
		r.writeTextRun(para, "☑ ", runStyle{})
	case types.Unchecked, types.UncheckedInteractive:
		r.writeTextRun(para, "☐ ", runStyle{})
	}
	return nil
}

func (r *docxRenderer) renderAdmonitionParagraph(p *types.Paragraph, kind string) error {
	para := r.startParagraph(paragraphOptions{style: "Admonition"})
	label := r.admonitionLabel(kind) + ": "
	r.writeTextRun(para, label, runStyle{bold: true})
	if err := r.renderInlineElements(para, p.Elements, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

func (r *docxRenderer) renderAdmonitionBlock(b *types.DelimitedBlock, kind string) error {
	para := r.startParagraph(paragraphOptions{style: "Admonition"})
	label := r.admonitionLabel(kind) + ": "
	r.writeTextRun(para, label, runStyle{bold: true})
	r.endParagraph(para)
	return r.renderElements(b.Elements)
}

func isAdmonitionStyle(style string) bool {
	switch style {
	case types.Tip, types.Note, types.Important, types.Warning, types.Caution:
		return true
	default:
		return false
	}
}

func (r *docxRenderer) admonitionLabel(kind string) string {
	attr := ""
	switch strings.ToUpper(kind) {
	case types.Tip:
		attr = types.AttrTipCaption
	case types.Note:
		attr = types.AttrNoteCaption
	case types.Important:
		attr = types.AttrImportantCaption
	case types.Warning:
		attr = types.AttrWarningCaption
	case types.Caution:
		attr = types.AttrCautionCaption
	}
	if attr != "" {
		if label := r.ctx.attributes.GetAsStringWithDefault(attr, ""); label != "" {
			return label
		}
	}
	return strings.ToUpper(kind)
}
