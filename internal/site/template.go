package site

import (
	_ "embed"
	"fmt"
	htmltemplate "html/template"
	"os"
)

//go:embed default_template.html
var defaultTemplate string

// TemplateData is passed to the layout template for each page.
type TemplateData struct {
	Title    string
	Content  htmltemplate.HTML
	Nav      *NavNode
	BasePath string
	CSS      []string
}

// loadTemplate parses the layout template from a file path, or uses the embedded default.
func loadTemplate(templatePath string) (*htmltemplate.Template, error) {
	var src string
	if templatePath != "" {
		data, err := os.ReadFile(templatePath)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", templatePath, err)
		}
		src = string(data)
	} else {
		src = defaultTemplate
	}
	tmpl, err := htmltemplate.New("layout").Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	return tmpl, nil
}
