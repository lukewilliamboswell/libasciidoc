package docx

import (
	"strconv"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *docxRenderer) renderCrossReference(para *paragraphBuilder, ref *types.InternalCrossReference, style runStyle) error {
	refID, ok := ref.ID.(string)
	if !ok {
		text, err := r.renderPlainText(ref.ID)
		if err != nil {
			return err
		}
		r.writeTextRun(para, text, style)
		return nil
	}
	// Resolve the canonical element reference ID, which may differ in case
	// from the xref ID (e.g. parser produces "_target_section" but section
	// ID is "_Target_Section"). Use the canonical form for both label
	// lookup and bookmark anchor so they match exactly.
	canonicalID := r.resolveElementReferenceID(refID)

	var label interface{}
	if ref.Label != nil {
		label = ref.Label
	} else if canonicalID != "" {
		if title, found := r.ctx.elementReferences[canonicalID]; found {
			label = title
		} else {
			label = refID
		}
	}
	if canonicalID == "" {
		text, err := r.renderPlainText(label)
		if err != nil {
			return err
		}
		r.writeTextRun(para, text, style)
		return nil
	}
	return r.renderInternalHyperlink(para, canonicalID, label, style)
}

// resolveElementReferenceID finds the canonical element reference ID
// matching the given id. It tries an exact match first, then falls back
// to a case-insensitive match so that xref anchors and section bookmarks
// use identical names.
func (r *docxRenderer) resolveElementReferenceID(id string) string {
	if id == "" {
		return ""
	}
	if _, found := r.ctx.elementReferences[id]; found {
		return id
	}
	for key := range r.ctx.elementReferences {
		if strings.EqualFold(key, id) {
			return key
		}
	}
	return id
}

func (r *docxRenderer) renderFootnoteRef(para *paragraphBuilder, ref *types.FootnoteReference, style runStyle) error {
	if ref.ID == types.InvalidFootnoteReference {
		r.writeTextRun(para, "[missing footnote: "+ref.Ref+"]", style)
		return nil
	}
	// This is a structural run (footnote reference marker), not a text run.
	// WriteString flushes any pending text run first, then writes to para.xml.
	// writeRPr writes directly to para.xml (the underlying builder).
	para.WriteString(`<w:r>`)
	mergeRunStyle(style, runStyle{charStyle: "FootnoteReference", superscript: true}).writeRPr(&para.xml)
	para.WriteString(`<w:footnoteReference w:id="`)
	para.WriteString(strconv.Itoa(ref.ID))
	para.WriteString(`"/></w:r>`)
	return nil
}

func (r *docxRenderer) renderInlinePassthrough(para *paragraphBuilder, p *types.InlinePassthrough, style runStyle) error {
	// Render passthrough content as plain text
	return r.renderInlineElements(para, p.Elements, style)
}

func (r *docxRenderer) renderLineBreak(para *paragraphBuilder) {
	para.WriteString("<w:r><w:br/></w:r>")
}
