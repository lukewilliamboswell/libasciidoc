package sgml

import (
	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderInlineMenu(m *types.InlineMenu) (string, error) {
	return r.execute(r.inlineMenu, struct {
		Path []string
	}{
		Path: m.Path,
	})
}
