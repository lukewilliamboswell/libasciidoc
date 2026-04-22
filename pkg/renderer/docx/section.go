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

	number := r.ctx.sectionNumbering[s.GetID()]

	// Legal numbering: headings get w:numPr at the appropriate ilvl
	// instead of a plain text number prefix.
	if number != "" && r.legalNumID > 0 {
		// Map AsciiDoc section level to legal numbering ilvl.
		// The first numbered level (usually level 1, i.e. ==) maps to ilvl=0.
		ilvl := s.Level - 1
		if ilvl < 0 {
			ilvl = 0
		}
		if ilvl > 2 {
			ilvl = 2
		}
		opts.numID = r.legalNumID
		opts.level = ilvl
	}

	para := r.startParagraph(opts)

	// Plain text number prefix (fallback when legal numbering is not active).
	if number != "" && r.legalNumID == 0 {
		r.writeTextRun(para, number+" ", runStyle{})
	}

	// Render the section title as inline elements
	if err := r.renderInlineElements(para, s.Title, runStyle{}); err != nil {
		return err
	}
	r.endParagraph(para)

	// Track legal numbering scope for child elements (lists).
	wasLegal := r.inLegalNumbering
	if number != "" && r.legalNumID > 0 {
		r.inLegalNumbering = true
	}
	defer func() { r.inLegalNumbering = wasLegal }()

	// Render child elements
	return r.renderElements(s.Elements)
}
