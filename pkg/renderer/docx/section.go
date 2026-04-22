package docx

import (
	"strconv"

	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
)

func (r *docxRenderer) renderSection(s *types.Section) error {
	// Render heading: AsciiDoc level 0 = "Heading1", level 1 = "Heading2", etc.
	headingLevel := s.Level + 1
	if headingLevel > 9 {
		headingLevel = 9
	}
	opts := paragraphOptions{
		style:        "Heading" + strconv.Itoa(headingLevel),
		bookmarkName: s.GetID(),
	}
	para := r.startParagraph(opts)
	if number := r.ctx.sectionNumbering[s.GetID()]; number != "" {
		r.writeTextRun(para, number+" ", runStyle{})
	}

	// Render the section title as inline elements
	if err := r.renderInlineElements(para, s.Title, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)

	// Render child elements
	return r.renderElements(s.Elements)
}
