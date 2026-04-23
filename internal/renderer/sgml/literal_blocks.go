package sgml

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderLiteralBlock(ctx *context, b *types.DelimitedBlock) (string, error) {
	log.Debugf("rendering literal block")
	content, err := r.renderElements(ctx, b.Elements)
	if err != nil {
		return "", fmt.Errorf("unable to render literal block content: %w", err)
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render literal block roles: %w", err)
	}
	title, err := r.renderElementTitle(ctx, b.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to render literal block title: %w", err)
	}

	return r.execute(r.literalBlock, struct {
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
