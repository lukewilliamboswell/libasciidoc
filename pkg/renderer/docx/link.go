package docx

import (
	"path/filepath"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderLink(para *strings.Builder, l *types.InlineLink) error {
	if l.Location == nil {
		if id := l.Attributes.GetAsStringWithDefault(types.AttrID, ""); id != "" {
			r.writeBookmark(para, id)
		}
		return nil
	}
	url := l.Location.ToString()
	label := l.Attributes[types.AttrInlineLinkText]
	if label == nil {
		label = l.Location.ToDisplayString()
	}
	return r.renderExternalHyperlink(para, url, label, runStyle{charStyle: "Hyperlink"})
}

func (r *docxRenderer) renderExternalCrossReference(para *strings.Builder, xref *types.ExternalCrossReference, style runStyle) error {
	label := xref.Attributes[types.AttrXRefLabel]
	if label == nil {
		label = defaultCrossReferenceLabel(xref)
	}
	target := crossReferenceLocation(xref)
	if strings.HasPrefix(target, "#") {
		return r.renderInternalHyperlink(para, strings.TrimPrefix(target, "#"), label, style)
	}
	return r.renderExternalHyperlink(para, target, label, style)
}

func (r *docxRenderer) renderExternalHyperlink(para *strings.Builder, url string, label interface{}, style runStyle) error {
	id := r.doc.addExternalRelationship(relTypeHyperlink, url)
	para.WriteString(`<w:hyperlink r:id="`)
	para.WriteString(xmlAttr(id))
	para.WriteString(`" w:history="1">`)
	linkColor := r.ctx.theme.Link.FontColor
	if err := r.renderLabelInline(para, label, mergeRunStyle(style, runStyle{charStyle: "Hyperlink", underline: true, color: linkColor})); err != nil {
		return err
	}
	para.WriteString(`</w:hyperlink>`)
	return nil
}

func (r *docxRenderer) renderInternalHyperlink(para *strings.Builder, id string, label interface{}, style runStyle) error {
	para.WriteString(`<w:hyperlink w:anchor="`)
	para.WriteString(xmlAttr(sanitizeBookmarkName(id)))
	para.WriteString(`" w:history="1">`)
	linkColor := r.ctx.theme.Link.FontColor
	if err := r.renderLabelInline(para, label, mergeRunStyle(style, runStyle{charStyle: "Hyperlink", underline: true, color: linkColor})); err != nil {
		return err
	}
	para.WriteString(`</w:hyperlink>`)
	return nil
}

func (r *docxRenderer) renderLabelInline(para *strings.Builder, label interface{}, style runStyle) error {
	switch label := label.(type) {
	case string:
		r.writeTextRun(para, label, style)
		return nil
	case []interface{}:
		return r.renderInlineElements(para, label, style)
	case nil:
		return nil
	default:
		text, err := r.renderPlainText(label)
		if err != nil {
			return err
		}
		r.writeTextRun(para, text, style)
		return nil
	}
}

func mergeRunStyle(base, extra runStyle) runStyle {
	if extra.bold {
		base.bold = true
	}
	if extra.italic {
		base.italic = true
	}
	if extra.monospace {
		base.monospace = true
	}
	if extra.highlight {
		base.highlight = true
	}
	if extra.subscript {
		base.subscript = true
	}
	if extra.superscript {
		base.superscript = true
	}
	if extra.underline {
		base.underline = true
	}
	if extra.font != "" {
		base.font = extra.font
	}
	if extra.color != "" {
		base.color = extra.color
	}
	if extra.charStyle != "" {
		base.charStyle = extra.charStyle
	}
	return base
}

func defaultCrossReferenceLabel(xref *types.ExternalCrossReference) string {
	loc := xref.Location.ToDisplayString()
	ext := filepath.Ext(loc)
	if ext == "" {
		return "[" + loc + "]"
	}
	return loc[:len(loc)-len(ext)] + ".html"
}

func crossReferenceLocation(xref *types.ExternalCrossReference) string {
	loc := xref.Location.ToDisplayString()
	ext := filepath.Ext(loc)
	if ext == "" {
		return "#" + loc
	}
	return loc[:len(loc)-len(ext)] + ".html"
}
