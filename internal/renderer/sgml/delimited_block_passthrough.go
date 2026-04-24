package sgml

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderPassthroughBlock(ctx *context, b *types.DelimitedBlock) (string, error) {
	previousWithinDelimitedBlock := ctx.withinDelimitedBlock
	defer func() {
		ctx.withinDelimitedBlock = previousWithinDelimitedBlock
	}()
	ctx.withinDelimitedBlock = true
	content, err := r.renderElements(ctx, b.Elements)
	if err != nil {
		return "", fmt.Errorf("unable to render passthrough block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render passthrough block roles: %w", err)
	}
	return r.execute(r.passthroughBlock, struct {
		Context *context
		ID      string
		Roles   string
		Content string
	}{
		Context: ctx,
		ID:      r.renderElementID(b.Attributes),
		Roles:   roles,
		Content: strings.Trim(content, "\n"),
	})
}
