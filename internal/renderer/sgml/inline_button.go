package sgml

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderInlineButton(b *types.InlineButton) (string, error) {
	result := &strings.Builder{}
	tmpl, err := r.inlineButton()
	if err != nil {
		return "", errors.Wrap(err, "unable to load inline button template")
	}
	if err = tmpl.Execute(result, b.Attributes[types.AttrButtonLabel]); err != nil {
		return "", errors.Wrap(err, "unable to render inline button")
	}
	return result.String(), nil
}
