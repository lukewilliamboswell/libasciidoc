package sgml

import (
	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *sgmlRenderer) renderPredefinedAttribute(a *types.PredefinedAttribute) (string, error) {
	return predefinedAttribute(a.Name), nil
}
