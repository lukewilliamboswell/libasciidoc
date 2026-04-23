package renderer

import (
	"fmt"
	"io"

	"github.com/lukewilliamboswell/libasciidoc/configuration"
	"github.com/lukewilliamboswell/libasciidoc/internal/renderer/docx"
	"github.com/lukewilliamboswell/libasciidoc/internal/renderer/sgml/html5"
	"github.com/lukewilliamboswell/libasciidoc/internal/renderer/sgml/xhtml5"
	"github.com/lukewilliamboswell/libasciidoc/types"
)

func Render(doc *types.Document, config *configuration.Configuration, output io.Writer) (types.Metadata, error) {
	switch config.BackEnd {
	case "html", "html5":
		return html5.Render(doc, config, output)
	case "xhtml", "xhtml5":
		return xhtml5.Render(doc, config, output)
	case "docx":
		return docx.Render(doc, config, output)
	default:
		return types.Metadata{}, fmt.Errorf("backend '%s' not supported", config.BackEnd)
	}
}
