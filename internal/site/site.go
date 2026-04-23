package site

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"

	log "github.com/sirupsen/logrus"
)

// Config holds the settings for a static site build.
type Config struct {
	SourceDir    string
	OutputDir    string
	TemplatePath string
	BasePath     string
	CSS          []string
	Attributes   map[string]interface{}
}

type page struct {
	relPath    string // relative path from source dir, e.g. "guides/intro.adoc"
	outputPath string // relative output path, e.g. "guides/intro.html"
	title      string
	content    string // rendered HTML fragment
	weight     int
	isIndex    bool
}

// Build walks the source directory, renders all .adoc files, wraps them in a
// layout template with navigation, and copies static assets to the output directory.
func Build(config Config) error {
	if config.BasePath == "" {
		config.BasePath = "/"
	}
	if !strings.HasSuffix(config.BasePath, "/") {
		config.BasePath += "/"
	}

	// 1. Walk source directory
	adocFiles, assets, err := walkSource(config.SourceDir)
	if err != nil {
		return fmt.Errorf("walking source directory: %w", err)
	}
	if len(adocFiles) == 0 {
		return fmt.Errorf("no .adoc files found in %s", config.SourceDir)
	}
	log.Infof("found %d .adoc files and %d assets", len(adocFiles), len(assets))

	// 2. Render each .adoc file
	pages := make([]*page, 0, len(adocFiles))
	for _, relPath := range adocFiles {
		p, err := renderPage(config, relPath)
		if err != nil {
			return fmt.Errorf("rendering %s: %w", relPath, err)
		}
		pages = append(pages, p)
	}

	// 3. Build nav tree
	nav := buildNav(pages, config.BasePath)

	// 4. Load template
	tmpl, err := loadTemplate(config.TemplatePath)
	if err != nil {
		return err
	}

	// 5. Write each page
	for _, p := range pages {
		pageNav := cloneNav(nav)
		currentHref := config.BasePath + strings.ReplaceAll(p.outputPath, string(filepath.Separator), "/")
		setActive(pageNav, currentHref)

		data := TemplateData{
			Title:    p.title,
			Content:  htmltemplate.HTML(p.content),
			Nav:      pageNav,
			BasePath: config.BasePath,
			CSS:      config.CSS,
		}

		outPath := filepath.Join(config.OutputDir, p.outputPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating %s: %w", outPath, err)
		}
		err = tmpl.Execute(f, data)
		f.Close()
		if err != nil {
			return fmt.Errorf("rendering template for %s: %w", p.relPath, err)
		}
		log.Infof("wrote %s", outPath)
	}

	// 6. Copy static assets
	if err := copyAssets(config.SourceDir, config.OutputDir, assets); err != nil {
		return err
	}

	log.Infof("site built: %d pages, %d assets -> %s", len(pages), len(assets), config.OutputDir)
	return nil
}

func renderPage(config Config, relPath string) (*page, error) {
	sourcePath := filepath.Join(config.SourceDir, relPath)
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	cfg := configuration.NewConfiguration(
		configuration.WithFilename(absPath),
		configuration.WithAttributes(config.Attributes),
		configuration.WithCSS(config.CSS),
		configuration.WithBackEnd("html5"),
		configuration.WithHeaderFooter(false),
	)
	metadata, err := asciidoc.ConvertFile(&buf, cfg)
	if err != nil {
		return nil, err
	}

	outputPath := strings.TrimSuffix(relPath, filepath.Ext(relPath)) + ".html"
	name := filepath.Base(relPath)

	p := &page{
		relPath:    relPath,
		outputPath: outputPath,
		title:      metadata.Title,
		content:    buf.String(),
		isIndex:    strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name))) == "index",
	}

	// extract weight from document attributes
	if metadata.Attributes != nil {
		if w, ok := metadata.Attributes["weight"]; ok {
			switch v := w.(type) {
			case string:
				if n, err := strconv.Atoi(v); err == nil {
					p.weight = n
				}
			case int:
				p.weight = v
			case float64:
				p.weight = int(v)
			}
		}
	}

	return p, nil
}
