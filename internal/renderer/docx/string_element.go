package docx

import (
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *docxRenderer) renderStringElement(para *strings.Builder, e *types.StringElement, style runStyle) error {
	r.writeTextRun(para, e.Content, style)
	return nil
}
