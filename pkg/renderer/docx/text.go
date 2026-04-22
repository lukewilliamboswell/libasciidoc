package docx

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderPlainText(element interface{}) (string, error) {
	switch e := element.(type) {
	case nil:
		return "", nil
	case []interface{}:
		buf := &strings.Builder{}
		for _, child := range e {
			text, err := r.renderPlainText(child)
			if err != nil {
				return "", err
			}
			buf.WriteString(text)
		}
		return buf.String(), nil
	case *types.StringElement:
		return e.Content, nil
	case *types.QuotedText:
		return r.renderPlainText(e.Elements)
	case *types.SpecialCharacter:
		if mapped, ok := specialCharacters[e.Name]; ok {
			return mapped, nil
		}
		return e.Name, nil
	case *types.Symbol:
		if mapped, ok := symbols[e.Name]; ok {
			return mapped, nil
		}
		return e.Name, nil
	case *types.PredefinedAttribute:
		if mapped, ok := predefinedAttributes[e.Name]; ok {
			return mapped, nil
		}
		return e.Name, nil
	case *types.InlineLink:
		if label, ok := e.Attributes[types.AttrInlineLinkText]; ok {
			return r.renderPlainText(label)
		}
		if e.Location == nil {
			return "", nil
		}
		return e.Location.ToDisplayString(), nil
	case *types.InternalCrossReference:
		if e.Label != nil {
			return r.renderPlainText(e.Label)
		}
		if id, ok := e.ID.(string); ok {
			if target, found := r.ctx.elementReferences[id]; found {
				return r.renderPlainText(target)
			}
			return "[" + id + "]", nil
		}
		return r.renderPlainText(e.ID)
	case *types.ExternalCrossReference:
		if label, ok := e.Attributes[types.AttrXRefLabel]; ok {
			return r.renderPlainText(label)
		}
		return defaultCrossReferenceLabel(e), nil
	case *types.InlineImage:
		src := ""
		if e.Location != nil {
			src = e.Location.ToString()
		}
		return imageAlt(e.Attributes, src), nil
	case *types.Icon:
		return e.Attributes.GetAsStringWithDefault(types.AttrImageAlt, e.Class), nil
	case *types.InlineButton:
		return e.Attributes.GetAsStringWithDefault(types.AttrButtonLabel, ""), nil
	case *types.InlineMenu:
		return strings.Join(e.Path, " > "), nil
	case *types.FootnoteReference:
		if e.ID == types.InvalidFootnoteReference {
			return "", nil
		}
		return "[" + strconv.Itoa(e.ID) + "]", nil
	case *types.InlinePassthrough:
		return r.renderPlainText(e.Elements)
	case *types.LineBreak:
		return "\n", nil
	case *types.Callout:
		return fmt.Sprintf("<%d>", e.Ref), nil
	case *types.UserMacro:
		return e.RawText, nil
	case *types.IndexTerm:
		return r.renderPlainText(e.Term)
	case *types.ConcealedIndexTerm:
		return "", nil
	default:
		return "", fmt.Errorf("docx: unable to render plain text for %T", element)
	}
}

func (r *docxRenderer) renderUserMacroInline(para *strings.Builder, m *types.UserMacro, style runStyle) error {
	tmpl, ok := r.ctx.config.Macros[m.Name]
	if !ok {
		r.writeTextRun(para, m.RawText, style)
		return nil
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, m); err != nil {
		return err
	}
	r.writeTextRun(para, buf.String(), style)
	return nil
}

func (r *docxRenderer) renderFootnotes(notes []*types.Footnote) error {
	if len(notes) == 0 {
		return nil
	}
	r.doc.hasFootnotes = true
	oldWriter := r.writer
	for _, note := range notes {
		r.doc.footnotes.WriteString(`<w:footnote w:id="`)
		r.doc.footnotes.WriteString(strconv.Itoa(note.ID))
		r.doc.footnotes.WriteString(`">`)
		r.writer = &r.doc.footnotes
		para := r.startParagraph(paragraphOptions{style: "FootnoteText"})
		para.WriteString(`<w:r><w:rPr><w:rStyle w:val="FootnoteReference"/></w:rPr><w:footnoteRef/></w:r>`)
		if err := r.renderInlineElements(para, note.Elements, runStyle{}); err != nil {
			r.writer = oldWriter
			return err
		}
		r.endParagraph(para)
		r.doc.footnotes.WriteString(`</w:footnote>`)
	}
	r.writer = oldWriter
	return nil
}

func (r *docxRenderer) prerenderTableOfContents(toc *types.TableOfContents) error {
	if toc == nil {
		return nil
	}
	return r.prerenderTableOfContentsSections(toc.Sections)
}

func (r *docxRenderer) prerenderTableOfContentsSections(sections []*types.ToCSection) error {
	for _, section := range sections {
		if err := r.prerenderTableOfContentsEntry(section); err != nil {
			return err
		}
	}
	return nil
}

func (r *docxRenderer) prerenderTableOfContentsEntry(entry *types.ToCSection) error {
	if entry == nil {
		return nil
	}
	if err := r.prerenderTableOfContentsSections(entry.Children); err != nil {
		return err
	}
	entry.Number = r.ctx.sectionNumbering[entry.ID]
	target, found := r.ctx.elementReferences[entry.ID]
	if !found {
		return fmt.Errorf("unable to render table of contents entry title for '%s'", entry.ID)
	}
	title, err := r.renderPlainText(target)
	if err != nil {
		return err
	}
	entry.Title = title
	return nil
}

func (r *docxRenderer) renderTableOfContents(toc *types.TableOfContents) error {
	if toc == nil || len(toc.Sections) == 0 {
		return nil
	}
	title := "Table of Contents"
	if value, found := r.ctx.attributes[types.AttrTableOfContentsTitle]; found {
		rendered, err := r.renderPlainText(value)
		if err != nil {
			return err
		}
		if rendered != "" {
			title = rendered
		}
	}
	if err := r.renderTextParagraph(title, paragraphOptions{style: "Heading1"}); err != nil {
		return err
	}
	return r.renderTableOfContentsSections(toc.Sections)
}

func (r *docxRenderer) renderTableOfContentsSections(sections []*types.ToCSection) error {
	for _, section := range sections {
		if err := r.renderTableOfContentsEntry(section); err != nil {
			return err
		}
	}
	return nil
}

func (r *docxRenderer) renderTableOfContentsEntry(entry *types.ToCSection) error {
	text := entry.Title
	if entry.Number != "" {
		text = entry.Number + " " + text
	}
	para := r.startParagraph(paragraphOptions{style: "ListParagraph"})
	if entry.ID != "" {
		if err := r.renderInternalHyperlink(para, entry.ID, text, runStyle{}); err != nil {
			return err
		}
	} else {
		r.writeTextRun(para, text, runStyle{})
	}
	r.endParagraph(para)
	return r.renderTableOfContentsSections(entry.Children)
}

func (r *docxRenderer) writeBookmark(para *strings.Builder, id string) {
	if id == "" {
		return
	}
	bookmarkID := r.doc.nextBookmarkID()
	name := sanitizeBookmarkName(id)
	para.WriteString(`<w:bookmarkStart w:id="`)
	para.WriteString(strconv.Itoa(bookmarkID))
	para.WriteString(`" w:name="`)
	para.WriteString(xmlAttr(name))
	para.WriteString(`"/><w:bookmarkEnd w:id="`)
	para.WriteString(strconv.Itoa(bookmarkID))
	para.WriteString(`"/>`)
}

var invalidBookmarkChars = regexp.MustCompile(`[^A-Za-z0-9_]`)

func sanitizeBookmarkName(id string) string {
	id = strings.TrimPrefix(id, "#")
	id = invalidBookmarkChars.ReplaceAllString(id, "_")
	if id == "" {
		return "_"
	}
	if id[0] >= '0' && id[0] <= '9' {
		id = "_" + id
	}
	if len(id) > 40 {
		id = id[:40]
	}
	return id
}
