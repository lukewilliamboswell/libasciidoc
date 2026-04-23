package sgml

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderInlineButton(b *types.InlineButton) (string, error) {
	result := &strings.Builder{}
	tmpl, err := r.inlineButton()
	if err != nil {
		return "", fmt.Errorf("unable to load inline button template: %w", err)
	}
	if err = tmpl.Execute(result, b.Attributes[types.AttrButtonLabel]); err != nil {
		return "", fmt.Errorf("unable to render inline button: %w", err)
	}
	return result.String(), nil
}
