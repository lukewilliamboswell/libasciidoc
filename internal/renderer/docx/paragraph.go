package docx

import (
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *docxRenderer) renderParagraph(p *types.Paragraph) error {
	style, _ := p.Attributes.GetAsString(types.AttrStyle)

	// Theme-defined block roles take precedence over the inherent style so an
	// author who applies `[.placeholder]` to e.g. a quote gets the placeholder
	// styling. Undefined roles fall through to the inherent style.
	if roleStyle, ok := r.resolveRoleStyleID(p.Attributes); ok {
		return r.renderRoleParagraph(p, roleStyle)
	}

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
	// Inline elements that resolve to nothing but line breaks (e.g. `{empty} +`
	// stacks used for vertical whitespace, or stray hard breaks at the head or
	// tail of a paragraph) would otherwise emit a `<w:p>` containing only
	// `<w:r><w:br/></w:r>` runs — valid OOXML but ugly hygiene that adds
	// surprise vertical extent in Word. Render them as a single empty paragraph
	// so the paragraph break is preserved without the bare break runs.
	blank, err := r.paragraphIsBlankModuloLineBreaks(p)
	if err != nil {
		return err
	}
	para := r.startParagraph(paragraphOptions{indentLeft: r.effectiveBodyIndent()})
	if err := r.renderCheckPrefix(para, p); err != nil {
		return err
	}
	if !blank {
		if err := r.renderInlineElements(para, p.Elements, runStyle{}); err != nil {
			return err
		}
	}
	r.endParagraph(para)
	return nil
}

// paragraphIsBlankModuloLineBreaks reports whether the paragraph's inline
// elements collapse to nothing more than line breaks and empty-text runs.
// Used to suppress emission of `<w:r><w:br/></w:r>`-only paragraphs.
func (r *docxRenderer) paragraphIsBlankModuloLineBreaks(p *types.Paragraph) (bool, error) {
	if len(p.Elements) == 0 {
		return false, nil
	}
	if p.Attributes[types.AttrCheckStyle] != nil {
		return false, nil
	}
	text, err := r.renderPlainText(p.Elements)
	if err != nil {
		return false, err
	}
	return strings.ReplaceAll(text, "\n", "") == "", nil
}

func (r *docxRenderer) renderRoleParagraph(p *types.Paragraph, styleID string) error {
	para := r.startParagraph(paragraphOptions{style: styleID, indentLeft: r.effectiveBodyIndent()})
	if err := r.renderCheckPrefix(para, p); err != nil {
		return err
	}
	if err := r.renderInlineElements(para, p.Elements, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

// resolveRoleStyleID looks at the block's roles in source order and returns
// the OOXML style id of the first role that has a matching theme entry.
// Roles undefined in the theme are silently skipped (AsciiDoc allows
// arbitrary role names).
func (r *docxRenderer) resolveRoleStyleID(attrs types.Attributes) (string, bool) {
	names := blockRoleNames(attrs)
	if len(names) == 0 {
		return "", false
	}
	role, ok := r.ctx.theme.Roles.FirstDefined(names)
	if !ok {
		return "", false
	}
	return role.StyleID, true
}

// blockRoleNames returns the role names attached to a block, in source order.
// Returns nil if no roles are present.
func blockRoleNames(attrs types.Attributes) []string {
	raw, ok := attrs[types.AttrRoles]
	if !ok {
		return nil
	}
	var out []string
	switch v := raw.(type) {
	case types.Roles:
		for _, r := range v {
			if s, ok := r.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case []interface{}:
		for _, r := range v {
			if s, ok := r.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case string:
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func (r *docxRenderer) renderCodeParagraph(p *types.Paragraph) error {
	para := r.startParagraph(paragraphOptions{style: "CodeBlock", indentLeft: r.effectiveBodyIndent()})
	if err := r.renderInlineElements(para, p.Elements, runStyle{monospace: true, monoFont: r.ctx.theme.Code.FontFamily}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

func (r *docxRenderer) renderParagraphAsListItem(p *types.Paragraph, opts paragraphOptions) error {
	if opts.style == "" {
		opts.style = "ListParagraph"
	}
	segments := splitAtLineBreaks(p.Elements)
	for i, seg := range segments {
		segOpts := opts
		if i > 0 {
			// Subsequent segments: same indent as the list text, no numbering marker.
			segOpts = paragraphOptions{
				style:      "ListParagraph",
				indentLeft: opts.indentLeft,
			}
		}
		para := r.startParagraph(segOpts)
		if i == 0 {
			if err := r.renderCheckPrefix(para, p); err != nil {
				return err
			}
		}
		if err := r.renderInlineElements(para, seg, runStyle{}); err != nil {
			return err
		}
		r.endParagraph(para)
	}
	return nil
}

// splitAtLineBreaks splits an inline element slice at every types.LineBreak.
// If no LineBreak is present the original slice is returned as-is in a
// single-element result (no allocation).
func splitAtLineBreaks(elements []interface{}) [][]interface{} {
	hasBreak := false
	for _, e := range elements {
		if _, ok := e.(*types.LineBreak); ok {
			hasBreak = true
			break
		}
	}
	if !hasBreak {
		return [][]interface{}{elements}
	}
	var segs [][]interface{}
	cur := make([]interface{}, 0, len(elements))
	for _, e := range elements {
		if _, ok := e.(*types.LineBreak); ok {
			segs = append(segs, cur)
			cur = make([]interface{}, 0, len(elements))
		} else {
			cur = append(cur, e)
		}
	}
	return append(segs, cur)
}

func (r *docxRenderer) renderCheckPrefix(para *paragraphBuilder, p *types.Paragraph) error {
	switch p.Attributes[types.AttrCheckStyle] {
	case types.Checked, types.CheckedInteractive:
		r.writeTextRun(para, "☑ ", runStyle{})
	case types.Unchecked, types.UncheckedInteractive:
		r.writeTextRun(para, "☐ ", runStyle{})
	}
	return nil
}

func (r *docxRenderer) renderAdmonitionParagraph(p *types.Paragraph, kind string) error {
	para := r.startParagraph(paragraphOptions{style: "Admonition", indentLeft: r.effectiveBodyIndent()})
	label := r.admonitionLabel(kind) + ": "
	r.writeTextRun(para, label, r.admonitionLabelStyle())
	if err := r.renderInlineElements(para, p.Elements, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)
	return nil
}

func (r *docxRenderer) renderAdmonitionBlock(b *types.DelimitedBlock, kind string) error {
	para := r.startParagraph(paragraphOptions{style: "Admonition"})
	label := r.admonitionLabel(kind) + ": "
	r.writeTextRun(para, label, r.admonitionLabelStyle())
	r.endParagraph(para)
	return r.renderElements(b.Elements)
}

func (r *docxRenderer) admonitionLabelStyle() runStyle {
	adm := r.ctx.theme.Admonition
	bold, italic := fontStyleBoldItalic(adm.LabelFontStyle)
	return runStyle{
		charStyle: "AdmonitionLabel",
		bold:      bold,
		italic:    italic,
		color:     adm.LabelFontColor,
	}
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
