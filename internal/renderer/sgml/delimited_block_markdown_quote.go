package sgml

import (
	"fmt"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderMarkdownQuoteBlock(ctx *context, b *types.DelimitedBlock) (string, error) {
	content, err := r.renderElements(ctx, b.Elements)
	if err != nil {
		return "", fmt.Errorf("unable to render markdown quote block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render markdown quote block roles: %w", err)
	}
	attribution := newAttribution(b)
	title, err := r.renderElementTitle(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render markdown quote block title: %w", err)
	}
	return r.execute(r.markdownQuoteBlock, struct {
		Context     *context
		ID          string
		Title       string
		Roles       string
		Attribution Attribution
		Content     string
	}{
		Context:     ctx,
		ID:          r.renderElementID(b.Attributes),
		Title:       title,
		Roles:       roles,
		Attribution: attribution,
		Content:     strings.Trim(content, "\n"),
	})
}
