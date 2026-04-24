package sgml

import (
	"fmt"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderOpenBlock(ctx *context, b *types.DelimitedBlock) (string, error) {
	blocks := discardBlankLines(b.Elements)
	content, err := r.renderElements(ctx, blocks)
	if err != nil {
		return "", fmt.Errorf("unable to render open block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render open block roles: %w", err)
	}
	title, err := r.renderElementTitle(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render open block title: %w", err)
	}
	return r.execute(r.openBlock, struct {
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
		Content: content,
	})
}
