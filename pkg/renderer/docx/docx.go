package docx

import (
	"io"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
	"github.com/pkg/errors"
)

// Render renders the document as a DOCX file and writes it to the output writer.
func Render(doc *types.Document, config *configuration.Configuration, output io.Writer) (types.Metadata, error) {
	d := newDocxDocument()

	ctx := newContext(doc, config)
	r := &docxRenderer{
		doc:    d,
		ctx:    ctx,
		writer: &d.body,
	}

	var metadata types.Metadata
	metadata.LastUpdated = config.LastUpdated.Format(configuration.LastUpdatedFormat)
	metadata.TableOfContents = doc.TableOfContents

	// Process header attribute declarations
	if header, _ := doc.Header(); header != nil {
		for _, e := range header.Elements {
			switch e := e.(type) {
			case *types.AttributeDeclaration:
				ctx.attributes[e.Name] = e.Value
			case *types.AttributeReset:
				delete(ctx.attributes, e.Name)
			}
		}
	}

	// Process standalone attribute declarations before the first rendered block.
bodyAttributes:
	for _, e := range doc.BodyElements() {
		switch e := e.(type) {
		case *types.AttributeDeclaration:
			ctx.attributes[e.Name] = e.Value
		case *types.AttributeReset:
			delete(ctx.attributes, e.Name)
		default:
			break bodyAttributes
		}
	}

	// Section numbering
	var err error
	if ctx.sectionNumbering, err = doc.SectionNumbers(); err != nil {
		return metadata, errors.Wrap(err, "unable to render docx document")
	}

	// Render document title
	if header, _ := doc.Header(); header != nil && header.Title != nil {
		title, err := r.renderPlainText(header.Title)
		if err != nil {
			return metadata, errors.Wrap(err, "unable to render document title")
		}
		metadata.Title = title

		// Add title to the document
		if err := r.renderTextParagraph(title, paragraphOptions{style: "Title"}); err != nil {
			return metadata, errors.Wrap(err, "unable to render document title")
		}

		// Add authors
		if authors := header.Authors(); authors != nil {
			metadata.Authors = authors
			authorNames := make([]string, len(authors))
			for i, author := range authors {
				authorNames[i] = author.FullName()
			}
			if err := r.renderTextParagraph(joinStrings(authorNames, "; "), paragraphOptions{style: "Subtitle"}); err != nil {
				return metadata, errors.Wrap(err, "unable to render document authors")
			}
		}

		// Add revision
		if revision := header.Revision(); revision != nil {
			metadata.Revision = *revision
		}
	}

	// Render body elements
	elements, err := r.bodyElementsWithTableOfContents(doc)
	if err != nil {
		return metadata, errors.Wrap(err, "unable to render docx document")
	}
	if err := r.renderElements(elements); err != nil {
		return metadata, errors.Wrap(err, "unable to render docx document")
	}
	if err := r.renderFootnotes(doc.Footnotes); err != nil {
		return metadata, errors.Wrap(err, "unable to render docx footnotes")
	}

	// Write the document
	if err := d.WriteTo(output); err != nil {
		return metadata, errors.Wrap(err, "unable to write docx document")
	}

	return metadata, nil
}

func (r *docxRenderer) bodyElementsWithTableOfContents(doc *types.Document) ([]interface{}, error) {
	elements := doc.BodyElements()
	toc := doc.TableOfContents
	if toc == nil {
		return elements, nil
	}
	if err := r.prerenderTableOfContents(toc); err != nil {
		return nil, err
	}
	placement, found := r.ctx.attributes[types.AttrTableOfContents]
	if !found {
		return elements, nil
	}
	switch placement {
	case "preamble":
		for _, element := range elements {
			if p, ok := element.(*types.Preamble); ok {
				p.TableOfContents = toc
				break
			}
		}
		return elements, nil
	case "", nil:
		result := make([]interface{}, len(elements)+1)
		result[0] = toc
		copy(result[1:], elements)
		return result, nil
	default:
		return nil, errors.Errorf("unsupported table of contents placement: '%v'", placement)
	}
}

func joinStrings(values []string, sep string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += sep + value
	}
	return result
}
