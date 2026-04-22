package docx

import (
	"math"
	"strconv"
)

// DocxTheme holds styling properties for DOCX rendering.
// Zero values in loaded themes are replaced by defaults.
type DocxTheme struct {
	Page    PageTheme
	Base    BaseTheme
	Heading HeadingTheme
	Title   TitlePageTheme
	Table   TableTheme
	List    ListTheme
	Code    CodeTheme
	Link    LinkTheme
}

// PageTheme controls page dimensions and margins.
type PageTheme struct {
	Layout string     // "portrait" or "landscape"
	Size   string     // "A4", "Letter", "Legal"
	Margin [4]float64 // [top, right, bottom, left] in mm
}

// BaseTheme controls the default document font.
type BaseTheme struct {
	FontFamily string  // e.g. "Helvetica"
	FontColor  string  // 6-hex e.g. "000000"
	FontSize   float64 // in pt, e.g. 10.5
}

// HeadingTheme controls heading styles.
type HeadingTheme struct {
	FontFamily    string
	FontColor     string
	FontStyle     string // "bold", "italic", "bold_italic"
	TextTransform string // "uppercase", "lowercase", "capitalize", or "" for none
	H1FontSize    float64 // in pt; 0 means use formula fallback
	H2FontSize    float64
	H3FontSize    float64
	H4FontSize    float64
	H1TextTransform string // per-level overrides; empty means inherit from TextTransform
	H2TextTransform string
	H3TextTransform string
	H4TextTransform string
}

// TitlePageTheme controls the document title and subtitle.
type TitlePageTheme struct {
	TitleFontSize     float64
	TitleFontStyle    string // "bold", "italic", "bold_italic"
	TitleFontColor    string
	SubtitleFontSize  float64
	SubtitleFontColor string
}

// TableTheme controls table rendering.
type TableTheme struct {
	FontSize    float64
	BorderColor string
	BorderWidth float64 // in pt
	HeadBgColor string  // header row background, e.g. "F0F0F0"
}

// ListTheme controls list indentation.
type ListTheme struct {
	Indent float64 // base indent in pt
}

// CodeTheme controls code/literal text styling.
type CodeTheme struct {
	FontFamily string
	FontSize   float64 // in pt
}

// LinkTheme controls hyperlink styling.
type LinkTheme struct {
	FontColor string // 6-hex e.g. "0563C1"; used for Hyperlink character style
}

// DefaultTheme returns a theme matching the previously hardcoded values,
// ensuring zero behavioral change when no theme file is provided.
func DefaultTheme() *DocxTheme {
	return &DocxTheme{
		Page: PageTheme{
			Layout: "portrait",
			Size:   "A4",
			Margin: [4]float64{20, 20, 20, 20}, // 1134 twips ≈ 20mm
		},
		Base: BaseTheme{
			FontFamily: "Aptos",
			FontSize:   11, // 22 half-points
		},
		Heading: HeadingTheme{
			FontStyle: "bold",
		},
		Title: TitlePageTheme{
			TitleFontSize:  20, // 40 half-points
			TitleFontStyle: "bold",
			SubtitleFontSize: 12, // 24 half-points
		},
		Table: TableTheme{
			BorderColor: "auto",
			BorderWidth: 0.5, // w:sz="4" = 4 eighths of a point = 0.5pt
		},
		List: ListTheme{
			Indent: 36, // 720 twips / 20 = 36pt
		},
		Code: CodeTheme{
			FontFamily: "Courier New",
			FontSize:   10, // 20 half-points
		},
		Link: LinkTheme{
			FontColor: "0563C1",
		},
	}
}

// --- unit conversion helpers ---

// ptToHalfPt converts points to OOXML half-points (w:sz units).
func ptToHalfPt(pt float64) int {
	return int(math.Round(pt * 2))
}

// mmToTwips converts millimeters to twips (1 twip = 1/1440 inch, 1mm ≈ 56.693 twips).
func mmToTwips(mm float64) int {
	return int(math.Round(mm * 1440 / 25.4))
}

// ptToTwips converts points to twips (1pt = 20 twips).
func ptToTwips(pt float64) int {
	return int(math.Round(pt * 20))
}

// ptToEighths converts points to eighths of a point (OOXML w:sz for borders).
func ptToEighths(pt float64) int {
	return int(math.Round(pt * 8))
}

// pageSizeTwips returns page width and height in twips for the given size and layout.
func pageSizeTwips(size, layout string) (w, h int) {
	switch size {
	case "Letter":
		w, h = 12240, 15840
	case "Legal":
		w, h = 12240, 20160
	default: // A4
		w, h = 11906, 16838
	}
	if layout == "landscape" {
		w, h = h, w
	}
	return w, h
}

// headingSizeHalfPt returns the heading size in half-points for the given level (1-9).
// It uses theme-specified sizes for h1-h4 and falls back to a formula for h5-h9.
func (t *DocxTheme) headingSizeHalfPt(level int) int {
	var pt float64
	switch level {
	case 1:
		pt = t.Heading.H1FontSize
	case 2:
		pt = t.Heading.H2FontSize
	case 3:
		pt = t.Heading.H3FontSize
	case 4:
		pt = t.Heading.H4FontSize
	}
	if pt > 0 {
		return ptToHalfPt(pt)
	}
	// formula fallback matching original: 32 - level*2, min 20 (half-points)
	size := 32 - level*2
	if size < 20 {
		size = 20
	}
	return size
}

// headingTextTransform returns the text_transform value for the given heading level (1-9).
// Per-level overrides take precedence over the general heading text_transform.
func (t *DocxTheme) headingTextTransform(level int) string {
	var perLevel string
	switch level {
	case 1:
		perLevel = t.Heading.H1TextTransform
	case 2:
		perLevel = t.Heading.H2TextTransform
	case 3:
		perLevel = t.Heading.H3TextTransform
	case 4:
		perLevel = t.Heading.H4TextTransform
	}
	if perLevel != "" {
		return perLevel
	}
	return t.Heading.TextTransform
}

// itoa is a shorthand for strconv.Itoa.
func itoa(i int) string {
	return strconv.Itoa(i)
}
