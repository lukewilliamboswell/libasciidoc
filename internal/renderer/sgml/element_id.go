package sgml

import (
	texttemplate "text/template"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderElementID(attrs types.Attributes) string {
	if id, ok := attrs[types.AttrID].(string); ok {
		return texttemplate.HTMLEscapeString(id)
	}
	return ""
}
