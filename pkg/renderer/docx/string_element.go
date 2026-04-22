package docx

import (
	"strings"

	"github.com/bytesparadise/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderStringElement(para *strings.Builder, e *types.StringElement, style runStyle) error {
	r.writeTextRun(para, e.Content, style)
	return nil
}
