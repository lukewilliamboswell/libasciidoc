package docx

import (
	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *docxRenderer) renderStringElement(para *paragraphBuilder, e *types.StringElement, style runStyle) error {
	r.writeTextRun(para, e.Content, style)
	return nil
}
