package sgml

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderListingBlock(ctx *context, b *types.DelimitedBlock) (string, error) {
	if k, found := b.Attributes[types.AttrStyle]; found && k == types.Source {
		return r.renderSourceBlock(ctx, b)
	}
	previousWithinDelimitedBlock := ctx.withinDelimitedBlock
	defer func() {
		ctx.withinDelimitedBlock = previousWithinDelimitedBlock
	}()
	ctx.withinDelimitedBlock = true
	content, err := r.renderElements(ctx, b.Elements)
	if err != nil {
		return "", fmt.Errorf("unable to render listing block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render listing block roles: %w", err)
	}
	title, err := r.renderElementTitle(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render listing block title: %w", err)
	}
	return r.execute(r.listingBlock, struct {
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

func (r *sgmlRenderer) renderListingParagraph(ctx *context, p *types.Paragraph) (string, error) {
	content, err := r.renderElements(ctx, p.Elements)
	if err != nil {
		return "", fmt.Errorf("unable to render listing block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, p.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render listing block roles: %w", err)
	}
	title, err := r.renderElementTitle(ctx, p.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render listing paragraph roles: %w", err)
	}
	return r.execute(r.listingBlock, struct {
		Context *context
		ID      string
		Title   string
		Roles   string
		Content string
	}{
		Context: ctx,
		ID:      r.renderElementID(p.Attributes),
		Title:   title,
		Roles:   roles,
		Content: content,
	})
}
