package docx

import (
	"strings"
)

type docxRenderer struct {
	doc              *docxDocument
	ctx              *context
	writer           *strings.Builder
	listLevel        int
	legalNumID       int  // shared multi-level numID for legal numbering (0 = not active)
	inLegalNumbering bool // true when inside a section with legal numbering
	legalIndent      int  // left indent in twips for body text under current heading
	listContIndent   int  // left indent in twips for continuation blocks inside the current list item
}

// effectiveBodyIndent returns the combined left indent (in twips) for any
// body-level block rendered in the current context.  It adds together the
// legal-section indent (driven by heading depth) and the list-continuation
// indent (driven by the nesting level of the enclosing list item).
func (r *docxRenderer) effectiveBodyIndent() int {
	return r.legalIndent + r.listContIndent
}
