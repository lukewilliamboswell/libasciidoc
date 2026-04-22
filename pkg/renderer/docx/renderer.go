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
}
