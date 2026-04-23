package sgml

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderFencedBlock(ctx *context, b *types.DelimitedBlock) (string, error) {
	previousWithinDelimitedBlock := ctx.withinDelimitedBlock
	defer func() {
		ctx.withinDelimitedBlock = previousWithinDelimitedBlock
	}()
	ctx.withinDelimitedBlock = true
	content, err := r.renderElements(ctx, b.Elements)
	if err != nil {
		return "", fmt.Errorf("unable to render fenced block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render fenced block roles: %w", err)
	}
	title, err := r.renderElementTitle(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render fenced block roles: %w", err)
	}
	return r.execute(r.fencedBlock, struct {
		Context *context
		ID      string
		Title   string
		Roles   string
		Content string
	}{
		Context: ctx,
		ID:      r.renderElementID(b.Attributes),
		Title:   title,
		Roles:   roles,
		Content: strings.Trim(content, "\n"),
	})
}
