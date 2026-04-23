package docx

import (
	"strconv"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *docxRenderer) renderList(l *types.List) error {
	r.listLevel++
	defer func() { r.listLevel-- }()
	switch l.Kind {
	case types.OrderedListKind:
		return r.renderOrderedList(l)
	case types.UnorderedListKind:
		return r.renderUnorderedList(l)
	case types.LabeledListKind:
		return r.renderLabeledList(l)
	case types.CalloutListKind:
		return r.renderCalloutList(l)
	default:
		return r.renderOrderedList(l)
	}
}

func (r *docxRenderer) renderOrderedList(l *types.List) error {
	// Legal numbering: use the shared multi-level numID at ilvl 3+.
	// Each list gets its own w:num (with startOverride) referencing the
	// legal abstractNum so that (a) restarts under each heading.
	if r.inLegalNumbering && r.legalNumID > 0 {
		// ilvl 3 = (a), ilvl 4 = (i), ilvl 5 = (A)
		ilvl := 2 + r.listLevel
		if ilvl > 5 {
			ilvl = 5
		}
		numID := r.doc.addLegalListNum(ilvl)
		for _, item := range l.Elements {
			if ole, ok := item.(*types.OrderedListElement); ok {
				if err := r.renderListItem(numID, ole.Elements, ilvl); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Regular numbering outside legal sections.
	indent := (r.listLevel - 1) * twipsPerLevel
	numID := r.doc.addNumbering(orderedListFormat(l), l.Attributes.GetAsIntWithDefault(types.AttrStart, 1), indent)
	for _, item := range l.Elements {
		if ole, ok := item.(*types.OrderedListElement); ok {
			if err := r.renderListItem(numID, ole.Elements, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *docxRenderer) renderUnorderedList(l *types.List) error {
	indent := (r.listLevel - 1) * twipsPerLevel
	numID := r.doc.addNumbering("bullet", 1, indent)
	for _, item := range l.Elements {
		if ule, ok := item.(*types.UnorderedListElement); ok {
			if ule.CheckStyle != types.NoCheck {
				if err := r.renderChecklistItem(ule); err != nil {
					return err
				}
				continue
			}
			if err := r.renderListItem(numID, ule.Elements, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *docxRenderer) renderLabeledList(l *types.List) error {
	dl := r.ctx.theme.DescriptionList
	for _, item := range l.Elements {
		if lle, ok := item.(*types.LabeledListElement); ok {
			// Render term with description list theme
			termPara := r.startParagraph(paragraphOptions{indentLeft: r.legalIndent})
			bold, italic := fontStyleBoldItalic(dl.TermFontStyle)
			termStyle := runStyle{bold: bold, italic: italic, color: dl.TermFontColor}
			if dl.TermFontFamily != "" {
				termStyle.font = dl.TermFontFamily
			} else if r.ctx.theme.Heading.FontFamily != "" {
				termStyle.font = r.ctx.theme.Heading.FontFamily
			}
			if err := r.renderInlineElements(termPara, lle.Term, termStyle); err != nil {
				return err
			}
			r.endParagraph(termPara)
			// Render description elements
			for _, elem := range lle.Elements {
				if err := r.renderElement(elem); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *docxRenderer) renderCalloutList(l *types.List) error {
	for _, item := range l.Elements {
		if cle, ok := item.(*types.CalloutListElement); ok {
			prefix := "<" + strconv.Itoa(cle.Ref) + "> "
			if err := r.renderListItemWithPrefix(prefix, cle.Elements); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *docxRenderer) renderListItem(numID int, elements []interface{}, ilvl int) error {
	for i, elem := range elements {
		switch e := elem.(type) {
		case *types.Paragraph:
			if i == 0 {
				if err := r.renderParagraphAsListItem(e, paragraphOptions{numID: numID, level: ilvl}); err != nil {
					return err
				}
			} else if err := r.renderParagraphAsListItem(e, paragraphOptions{style: "ListParagraph"}); err != nil {
				return err
			}
		default:
			if err := r.renderElement(elem); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *docxRenderer) renderListItemWithPrefix(prefix string, elements []interface{}) error {
	for i, elem := range elements {
		switch e := elem.(type) {
		case *types.Paragraph:
			para := r.startParagraph(paragraphOptions{style: "ListParagraph"})
			if i == 0 {
				r.writeTextRun(para, prefix, runStyle{})
			}
			if err := r.renderInlineElements(para, e.Elements, runStyle{}); err != nil {
				return err
			}
			r.endParagraph(para)
		default:
			if err := r.renderElement(elem); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *docxRenderer) renderChecklistItem(ule *types.UnorderedListElement) error {
	prefix := "☐ "
	switch ule.CheckStyle {
	case types.Checked, types.CheckedInteractive:
		prefix = "☑ "
	}
	return r.renderListItemWithPrefix(prefix, ule.Elements)
}

func orderedListFormat(l *types.List) string {
	style := ""
	if len(l.Elements) > 0 {
		if first, ok := l.Elements[0].(*types.OrderedListElement); ok {
			style = first.Style
		}
	}
	if attrStyle := l.Attributes.GetAsStringWithDefault(types.AttrStyle, ""); attrStyle != "" {
		style = attrStyle
	}
	switch style {
	case types.LowerAlpha:
		return "lowerLetter"
	case types.UpperAlpha:
		return "upperLetter"
	case types.LowerRoman:
		return "lowerRoman"
	case types.UpperRoman:
		return "upperRoman"
	default:
		return "decimal"
	}
}
