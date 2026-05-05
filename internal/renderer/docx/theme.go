package docx

import (
	"math"
	"strconv"
	"strings"
)

// DocxTheme holds styling properties for DOCX rendering.
// Zero values in loaded themes are replaced by defaults.
type DocxTheme struct {
	Page            PageTheme
	Base            BaseTheme
	Heading         HeadingTheme
	Title           TitlePageTheme
	Table           TableTheme
	List            ListTheme
	Code            CodeTheme
	Codespan        CodespanTheme
	Link            LinkTheme
	Prose           ProseTheme
	Quote           QuoteTheme
	Admonition      AdmonitionTheme
	Sidebar         SidebarTheme
	Example         ExampleTheme
	Caption         CaptionTheme
	DescriptionList DescriptionListTheme
	RunningHeader   RunningHFTheme
	RunningFooter   RunningHFTheme
	Roles           RoleThemes
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
	FontStyle  string  // "bold", "italic", "bold_italic"
	TextAlign  string  // "left", "center", "right", "justify"
	LineHeight float64 // ratio (e.g. 1.15); 0 means use default 1.08
}

// HeadingTheme controls heading styles.
type HeadingTheme struct {
	FontFamily      string
	FontColor       string
	FontStyle       string  // "bold", "italic", "bold_italic"
	TextTransform   string  // "uppercase", "lowercase", "capitalize", or "" for none
	MarginTop       float64 // space before heading in pt
	MarginBottom    float64 // space after heading in pt
	H1FontSize      float64 // in pt; 0 means use formula fallback
	H2FontSize      float64
	H3FontSize      float64
	H4FontSize      float64
	H5FontSize      float64
	H6FontSize      float64
	H1TextTransform string // per-level overrides; empty means inherit from TextTransform
	H2TextTransform string
	H3TextTransform string
	H4TextTransform string
	H1FontColor     string // per-level color overrides
	H2FontColor     string
	H3FontColor     string
	H4FontColor     string
	H5FontColor     string
	H6FontColor     string
	H1FontStyle     string // per-level style overrides
	H2FontStyle     string
	H3FontStyle     string
	H4FontStyle     string
	H5FontStyle     string
	H6FontStyle     string
	H1MarginTop     float64 // per-level space before in pt; 0 means use global MarginTop
	H1MarginBottom  float64
	H2MarginTop     float64
	H2MarginBottom  float64
	H3MarginTop     float64
	H3MarginBottom  float64
	H4MarginTop     float64
	H4MarginBottom  float64
	H5MarginTop     float64
	H5MarginBottom  float64
	H6MarginTop     float64
	H6MarginBottom  float64
}

// TitlePageTheme controls the document title and subtitle.
type TitlePageTheme struct {
	TitleFontSize      float64
	TitleFontStyle     string // "bold", "italic", "bold_italic"
	TitleFontColor     string
	TitleFontFamily    string
	TitleTextTransform string // "uppercase", "lowercase", "capitalize", or ""
	TitleTextAlign     string // "left", "center", "right", "justify"
	SubtitleFontSize   float64
	SubtitleFontColor  string
	SubtitleFontFamily string
	SubtitleFontStyle  string // "bold", "italic", "bold_italic"
}

// TableTheme controls table rendering.
type TableTheme struct {
	Width         string // "auto" (fit content), "full" (fit page/window), or percentage like "80%"
	FontSize      float64
	BorderColor   string
	BorderWidth   float64 // in pt
	HeadBgColor   string  // header row background, e.g. "F0F0F0"
	HeadFontStyle string  // "bold", "italic", "bold_italic"
	CellPadding   float64 // in pt (applied as cell margins)
	GridColor     string  // color for internal grid lines
	GridWidth     float64 // width for internal grid lines in pt
	StripeBgColor string  // alternating row background
	FootBgColor   string  // footer row background
	FootFontStyle string  // "bold", "italic", "bold_italic"
}

// ListTheme controls list indentation and styling.
type ListTheme struct {
	Indent          float64 // base indent in pt
	ItemSpacing     float64 // vertical space between items in pt
	MarkerFontColor string  // color of bullets/numbers
}

// CodeTheme controls code/literal block styling.
type CodeTheme struct {
	FontFamily      string
	FontSize        float64 // in pt
	FontColor       string
	BackgroundColor string
	BorderColor     string
	BorderWidth     float64 // in pt
	LineHeight      float64 // ratio
}

// CodespanTheme controls inline monospace styling.
type CodespanTheme struct {
	FontFamily      string
	FontSize        float64
	FontColor       string
	BackgroundColor string
}

// LinkTheme controls hyperlink styling.
type LinkTheme struct {
	FontColor      string // 6-hex e.g. "0563C1"; used for Hyperlink character style
	FontStyle      string // "bold", "italic", "bold_italic"
	TextDecoration string // "underline", "none"; default underline
}

// ProseTheme controls paragraph text spacing and alignment.
type ProseTheme struct {
	MarginBottom float64 // space after paragraphs in pt
	TextAlign    string  // "left", "center", "right", "justify"
	TextIndent   float64 // first-line indent in pt
}

// QuoteTheme controls blockquote styling.
type QuoteTheme struct {
	FontSize    float64
	FontColor   string
	FontStyle   string // "bold", "italic", "bold_italic"
	FontFamily  string
	BorderColor string
	BorderWidth float64 // in pt
}

// AdmonitionTheme controls tip/note/warning/caution/important styling.
type AdmonitionTheme struct {
	FontColor          string
	FontSize           float64
	BackgroundColor    string
	BorderColor        string
	BorderWidth        float64
	LabelFontStyle     string  // "bold", "italic", "bold_italic"
	LabelFontColor     string
	LabelFontSize      float64 // in pt; 0 means inherit body size
	LabelTextTransform string  // "uppercase" | "" (only "uppercase" is honoured)
}

// SidebarTheme controls sidebar block styling.
type SidebarTheme struct {
	BackgroundColor string
	BorderColor     string
	BorderWidth     float64
	FontColor       string
	FontSize        float64
}

// ExampleTheme controls example block styling.
type ExampleTheme struct {
	BackgroundColor string
	BorderColor     string
	BorderWidth     float64
	FontColor       string
	FontSize        float64
}

// CaptionTheme controls figure/table caption styling.
type CaptionTheme struct {
	FontSize   float64
	FontStyle  string // "bold", "italic", "bold_italic"
	FontColor  string
	FontFamily string
	TextAlign  string
}

// DescriptionListTheme controls labeled/description list styling.
type DescriptionListTheme struct {
	TermFontStyle  string
	TermFontColor  string
	TermFontFamily string
	TermFontSize   float64
}

// RunningHFTheme controls running header or footer content.
type RunningHFTheme struct {
	Content    string  // legacy flat template string, used when no per-position content set
	FontSize   float64 // in pt
	FontColor  string
	FontFamily string
	FontStyle  string  // "bold", "italic", "bold_italic"
	Height     float64 // in mm (defaults from page margin header/footer distance)
	Recto      RunningHFSide
	Verso      RunningHFSide
}

// RunningHFSide holds the three positional slots for one side of a running
// header or footer (recto = odd / right-hand pages, verso = even / left-hand).
type RunningHFSide struct {
	Left   RunningHFSlot
	Center RunningHFSlot
	Right  RunningHFSlot
}

// IsSet reports whether any position on this side carries content.
func (s RunningHFSide) IsSet() bool {
	return s.Left.IsSet() || s.Center.IsSet() || s.Right.IsSet()
}

// RunningHFSlot represents a single positional content cell. Modeled as a
// struct with an explicit set bit so that "absent" is distinct from "present
// but empty" — important because applying a partial side over defaults must
// not silently clear earlier content.
type RunningHFSlot struct {
	text string
	set  bool
}

// NewRunningHFSlot constructs a slot with the given text marked as present.
func NewRunningHFSlot(text string) RunningHFSlot { return RunningHFSlot{text: text, set: true} }

// Text returns the slot's template string.
func (s RunningHFSlot) Text() string { return s.text }

// IsSet reports whether the slot was supplied in the source theme.
func (s RunningHFSlot) IsSet() bool { return s.set }

// HasPositions reports whether any per-position slot on either side is set.
func (t RunningHFTheme) HasPositions() bool {
	return t.Recto.IsSet() || t.Verso.IsSet()
}

// IsActive reports whether the running region should be emitted at all —
// either the legacy flat content or any per-position slot is present.
func (t RunningHFTheme) IsActive() bool {
	return t.Content != "" || t.HasPositions()
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
			FontStyle:      "bold",
			H1MarginTop:    math.MaxFloat64,
			H1MarginBottom: math.MaxFloat64,
			H2MarginTop:    math.MaxFloat64,
			H2MarginBottom: math.MaxFloat64,
			H3MarginTop:    math.MaxFloat64,
			H3MarginBottom: math.MaxFloat64,
			H4MarginTop:    math.MaxFloat64,
			H4MarginBottom: math.MaxFloat64,
			H5MarginTop:    math.MaxFloat64,
			H5MarginBottom: math.MaxFloat64,
			H6MarginTop:    math.MaxFloat64,
			H6MarginBottom: math.MaxFloat64,
		},
		Title: TitlePageTheme{
			TitleFontSize:    20, // 40 half-points
			TitleFontStyle:   "bold",
			SubtitleFontSize: 12, // 24 half-points
		},
		Table: TableTheme{
			Width:       "full", // auto-fit to page width (window)
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
		Admonition: AdmonitionTheme{
			LabelFontStyle: "bold",
		},
		Caption: CaptionTheme{
			FontStyle: "italic",
		},
		DescriptionList: DescriptionListTheme{
			TermFontStyle: "bold",
		},
		Roles: NewRoleThemes(),
	}
}

// tableWidthAttrs returns the OOXML w:tblW attributes for the given Width
// theme value. "full" = 100% of page width, "auto" = fit to content,
// a percentage like "80%" = that fraction of page width.
func tableWidthAttrs(width string) (w string, wType string) {
	switch {
	case width == "full" || width == "":
		return "5000", "pct" // 5000 fiftieths-of-a-percent = 100%
	case width == "auto":
		return "0", "auto"
	case len(width) > 1 && width[len(width)-1] == '%':
		pctStr := width[:len(width)-1]
		pct, err := strconv.ParseFloat(pctStr, 64)
		if err != nil || pct <= 0 || pct > 100 {
			return "5000", "pct" // fallback to full
		}
		return strconv.Itoa(int(math.Round(pct * 50))), "pct" // 50 per percent
	default:
		return "5000", "pct" // fallback to full
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
// It uses theme-specified sizes for h1-h6 and falls back to a formula for h7-h9.
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
	case 5:
		pt = t.Heading.H5FontSize
	case 6:
		pt = t.Heading.H6FontSize
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

// headingFontColor returns the font color for the given heading level (1-9).
// Per-level overrides take precedence over the general heading font color.
func (t *DocxTheme) headingFontColor(level int) string {
	var perLevel string
	switch level {
	case 1:
		perLevel = t.Heading.H1FontColor
	case 2:
		perLevel = t.Heading.H2FontColor
	case 3:
		perLevel = t.Heading.H3FontColor
	case 4:
		perLevel = t.Heading.H4FontColor
	case 5:
		perLevel = t.Heading.H5FontColor
	case 6:
		perLevel = t.Heading.H6FontColor
	}
	if perLevel != "" {
		return perLevel
	}
	return t.Heading.FontColor
}

// headingFontStyle returns the resolved bold/italic for the given heading level.
// Per-level overrides take precedence over the general heading font style.
func (t *DocxTheme) headingFontStyle(level int) (bold, italic bool) {
	style := t.Heading.FontStyle
	var perLevel string
	switch level {
	case 1:
		perLevel = t.Heading.H1FontStyle
	case 2:
		perLevel = t.Heading.H2FontStyle
	case 3:
		perLevel = t.Heading.H3FontStyle
	case 4:
		perLevel = t.Heading.H4FontStyle
	case 5:
		perLevel = t.Heading.H5FontStyle
	case 6:
		perLevel = t.Heading.H6FontStyle
	}
	if perLevel != "" {
		style = perLevel
	}
	bold = style == "bold" || style == "bold_italic"
	italic = style == "italic" || style == "bold_italic"
	return bold, italic
}

// headingMarginTop returns the space-before value (in pt) for the given heading level.
// math.MaxFloat64 is the sentinel meaning "not set"; fall back to global MarginTop.
func (t *DocxTheme) headingMarginTop(level int) float64 {
	perLevel := math.MaxFloat64
	switch level {
	case 1:
		perLevel = t.Heading.H1MarginTop
	case 2:
		perLevel = t.Heading.H2MarginTop
	case 3:
		perLevel = t.Heading.H3MarginTop
	case 4:
		perLevel = t.Heading.H4MarginTop
	case 5:
		perLevel = t.Heading.H5MarginTop
	case 6:
		perLevel = t.Heading.H6MarginTop
	}
	if perLevel != math.MaxFloat64 {
		return perLevel
	}
	return t.Heading.MarginTop
}

// headingMarginBottom returns the space-after value (in pt) for the given heading level.
// math.MaxFloat64 is the sentinel meaning "not set"; fall back to global MarginBottom.
func (t *DocxTheme) headingMarginBottom(level int) float64 {
	perLevel := math.MaxFloat64
	switch level {
	case 1:
		perLevel = t.Heading.H1MarginBottom
	case 2:
		perLevel = t.Heading.H2MarginBottom
	case 3:
		perLevel = t.Heading.H3MarginBottom
	case 4:
		perLevel = t.Heading.H4MarginBottom
	case 5:
		perLevel = t.Heading.H5MarginBottom
	case 6:
		perLevel = t.Heading.H6MarginBottom
	}
	if perLevel != math.MaxFloat64 {
		return perLevel
	}
	return t.Heading.MarginBottom
}

// lineHeightValue returns the OOXML w:line value for the base line height.
// OOXML line spacing in auto mode: 240 = single (1.0), 276 = 1.15, 360 = 1.5, etc.
func (t *DocxTheme) lineHeightValue() int {
	lh := t.Base.LineHeight
	if lh <= 0 {
		return 259 // default ~1.08 (Word default)
	}
	return int(math.Round(lh * 240))
}

// itoa is a shorthand for strconv.Itoa.
func itoa(i int) string {
	return strconv.Itoa(i)
}

// fontStyleBoldItalic parses a font_style string into bold and italic booleans.
func fontStyleBoldItalic(style string) (bold, italic bool) {
	bold = style == "bold" || style == "bold_italic"
	italic = style == "italic" || style == "bold_italic"
	return bold, italic
}

// RoleTheme is a normalised, validated set of paragraph-style overrides that
// apply when an AsciiDoc block carries the matching role attribute. Fields
// use IsSet-bearing types so that "absent" is distinct from an explicit zero.
type RoleTheme struct {
	Name            string // canonical (lower-case) role name
	StyleID         string // OOXML style id (e.g. "RolePlaceholder")
	StyleName       string // OOXML style display name
	FontFamily      string
	FontColor       string  // 6-hex RRGGBB; empty means inherit
	FontSize        float64 // pt; 0 means inherit
	FontStyle       string  // "bold" | "italic" | "bold_italic" | "normal" | ""
	BackgroundColor string  // 6-hex RRGGBB shading; empty means none
	BorderColor     string
	BorderWidth     float64 // pt
	PaddingTop      float64 // pt -> w:spacing/before
	PaddingRight    float64 // pt -> w:ind/right
	PaddingBottom   float64 // pt -> w:spacing/after
	PaddingLeft     float64 // pt -> w:ind/left
	MarginTop       float64 // pt; merges with PaddingTop additively at emission
	MarginBottom    float64 // pt; merges with PaddingBottom additively at emission
	TextAlign       string
	TextTransform   string // "uppercase" | "lowercase" | "" (none)
}

// RoleThemes is an ordered, case-insensitively-keyed collection of RoleTheme
// entries. The order is the order roles appear in the source YAML so that
// styles.xml emission is deterministic.
type RoleThemes struct {
	order   []string // canonical names in insertion order
	byName  map[string]RoleTheme
}

// NewRoleThemes returns an empty, ready-to-populate RoleThemes.
func NewRoleThemes() RoleThemes {
	return RoleThemes{byName: map[string]RoleTheme{}}
}

// canonicaliseRoleName lower-cases the role name; AsciiDoc role names are
// conventionally lower-kebab-case and `[.Foo]` is treated as the same role
// as `[.foo]`. Hyphens and underscores are preserved as written.
func canonicaliseRoleName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Set inserts or replaces a role definition. The role's Name field is
// canonicalised; StyleID / StyleName are derived if empty.
func (r *RoleThemes) Set(role RoleTheme) {
	if r.byName == nil {
		r.byName = map[string]RoleTheme{}
	}
	canon := canonicaliseRoleName(role.Name)
	if canon == "" {
		return
	}
	role.Name = canon
	if role.StyleID == "" {
		role.StyleID = roleStyleID(canon)
	}
	if role.StyleName == "" {
		role.StyleName = "Role " + canon
	}
	if _, exists := r.byName[canon]; !exists {
		r.order = append(r.order, canon)
	}
	r.byName[canon] = role
}

// Get looks up a role by name (case-insensitive). Returns the resolved role
// and true if defined.
func (r RoleThemes) Get(name string) (RoleTheme, bool) {
	if r.byName == nil {
		return RoleTheme{}, false
	}
	role, ok := r.byName[canonicaliseRoleName(name)]
	return role, ok
}

// Each iterates roles in insertion order (deterministic for emission).
func (r RoleThemes) Each(fn func(RoleTheme)) {
	for _, name := range r.order {
		fn(r.byName[name])
	}
}

// Len reports the number of roles defined.
func (r RoleThemes) Len() int { return len(r.order) }

// FirstDefined returns the first role from the candidate slice that has a
// matching definition. Used to implement "multiple roles, first wins" when
// a block carries `[.foo.bar]` — we keep AsciiDoc's source order.
func (r RoleThemes) FirstDefined(candidates []string) (RoleTheme, bool) {
	for _, c := range candidates {
		if role, ok := r.Get(c); ok {
			return role, true
		}
	}
	return RoleTheme{}, false
}

// roleStyleID builds the OOXML style id from a canonicalised role name.
// Scheme: "Role" + UpperCamelCase(name) with non-alphanumerics treated as
// word separators. `placeholder` -> `RolePlaceholder`; `legal-note` ->
// `RoleLegalNote`. Stable across invocations.
func roleStyleID(canon string) string {
	var b strings.Builder
	b.WriteString("Role")
	upper := true
	for _, r := range canon {
		switch {
		case r >= 'a' && r <= 'z':
			if upper {
				b.WriteRune(r - 32)
				upper = false
			} else {
				b.WriteRune(r)
			}
		case r >= 'A' && r <= 'Z':
			if upper {
				b.WriteRune(r)
				upper = false
			} else {
				b.WriteRune(r + 32)
			}
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			upper = false
		default:
			upper = true
		}
	}
	return b.String()
}
