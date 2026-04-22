package docx

import (
	"strings"

	"github.com/bytesparadise/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderQuotedText(para *strings.Builder, qt *types.QuotedText, style runStyle) error {
	merged := mergeStyle(style, qt.Kind)
	return r.renderInlineElements(para, qt.Elements, merged)
}

func mergeStyle(current runStyle, kind types.QuotedTextKind) runStyle {
	switch kind {
	case types.SingleQuoteBold, types.DoubleQuoteBold:
		current.bold = true
	case types.SingleQuoteItalic, types.DoubleQuoteItalic:
		current.italic = true
	case types.SingleQuoteMonospace, types.DoubleQuoteMonospace:
		current.monospace = true
	case types.SingleQuoteMarked, types.DoubleQuoteMarked:
		current.highlight = true
	case types.SingleQuoteSubscript:
		current.subscript = true
	case types.SingleQuoteSuperscript:
		current.superscript = true
	}
	return current
}
