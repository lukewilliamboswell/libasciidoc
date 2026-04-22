package docx

import (
	"strings"
)

type docxRenderer struct {
	doc       *docxDocument
	ctx       *context
	writer    *strings.Builder
	listLevel int
}
