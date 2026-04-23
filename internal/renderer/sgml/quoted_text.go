package sgml

import (
	"fmt"
	"strings"
	texttemplate "text/template"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

// TODO: The bold, italic, and monospace items should be refactored to support semantic tags instead.

func (r *sgmlRenderer) renderQuotedText(ctx *context, t *types.QuotedText) (string, error) {
	elementsBuffer := &strings.Builder{}
	for _, element := range t.Elements {
		b, err := r.renderElement(ctx, element)
		if err != nil {
			return "", fmt.Errorf("unable to render text quote: %w", err)
		}
		_, err = elementsBuffer.WriteString(b)
		if err != nil {
			return "", fmt.Errorf("unable to render text quote: %w", err)
		}
	}
	roles, err := r.renderElementRoles(ctx, t.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render quoted text roles: %w", err)
	}
	var tmpl *texttemplate.Template
	switch t.Kind {
	case types.SingleQuoteBold, types.DoubleQuoteBold:
		tmpl, err = r.boldText()
	case types.SingleQuoteItalic, types.DoubleQuoteItalic:
		tmpl, err = r.italicText()
	case types.SingleQuoteMarked, types.DoubleQuoteMarked:
		tmpl, err = r.markedText()
	case types.SingleQuoteMonospace, types.DoubleQuoteMonospace:
		tmpl, err = r.monospaceText()
	case types.SingleQuoteSubscript:
		tmpl, err = r.subscriptText()
	case types.SingleQuoteSuperscript:
		tmpl, err = r.superscriptText()
	default:
		return "", fmt.Errorf("unsupported quoted text kind: '%v'", t.Kind)
	}
	if err != nil {
		return "", fmt.Errorf("unable to load quoted text template: %w", err)
	}
	result := &strings.Builder{}
	if err := tmpl.Execute(result, struct {
		ID         string
		Roles      string
		Attributes types.Attributes
		Content    string
	}{
		Attributes: t.Attributes,
		ID:         r.renderElementID(t.Attributes),
		Roles:      roles,
		Content:    elementsBuffer.String(),
	}); err != nil {
		return "", fmt.Errorf("unable to render quoted text: %w", err)
	}
	return result.String(), nil
}
