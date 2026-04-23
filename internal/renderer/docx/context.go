package docx

import (
	"github.com/lukewilliamboswell/libasciidoc/configuration"
	"github.com/lukewilliamboswell/libasciidoc/types"
)

type context struct {
	config            *configuration.Configuration
	counters          map[string]int
	attributes        types.Attributes
	elementReferences types.ElementReferences
	hasHeader         bool
	sectionNumbering  types.SectionNumbers
	theme             *DocxTheme
}

func newContext(doc *types.Document, config *configuration.Configuration) (*context, error) {
	header, _ := doc.Header()
	attributes := config.Attributes.Clone()
	if attributes == nil {
		attributes = types.Attributes{}
	}
	theme := DefaultTheme()
	if config.ThemePath != "" {
		var err error
		theme, err = LoadTheme(config.ThemePath)
		if err != nil {
			return nil, err
		}
	}
	ctx := &context{
		config:            config,
		counters:          make(map[string]int),
		attributes:        attributes,
		elementReferences: doc.ElementReferences,
		hasHeader:         header != nil,
		theme:             theme,
	}
	if !ctx.attributes.Has(types.AttrFigureCaption) {
		ctx.attributes[types.AttrFigureCaption] = "Figure"
	}
	if !ctx.attributes.Has(types.AttrTableCaption) {
		ctx.attributes[types.AttrTableCaption] = "Table"
	}
	if header != nil {
		if authors := header.Authors(); authors != nil {
			ctx.attributes.AddAll(authors.Expand())
		}
		if revision := header.Revision(); revision != nil {
			ctx.attributes.AddAll(revision.Expand())
		}
	}
	return ctx, nil
}

const tableCounter = "tableCounter"

func (ctx *context) GetAndIncrementTableCounter() int {
	return ctx.getAndIncrementCounter(tableCounter)
}

const imageCounter = "imageCounter"

func (ctx *context) GetAndIncrementImageCounter() int {
	return ctx.getAndIncrementCounter(imageCounter)
}

func (ctx *context) getAndIncrementCounter(name string) int {
	if _, found := ctx.counters[name]; !found {
		ctx.counters[name] = 1
		return 1
	}
	ctx.counters[name]++
	return ctx.counters[name]
}
