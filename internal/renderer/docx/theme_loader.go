package docx

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadTheme reads an Asciidoctor PDF theme YAML file and returns a DocxTheme
// with the specified values merged onto the defaults.
func LoadTheme(path string) (*DocxTheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseTheme(data)
}

func parseTheme(data []byte) (*DocxTheme, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if err := resolveThemeVariables(&root); err != nil {
		return nil, err
	}
	var yt yamlTheme
	if err := root.Decode(&yt); err != nil {
		return nil, err
	}
	theme := DefaultTheme()
	yt.applyTo(theme)
	return theme, nil
}

// Length is a typed length value used in theme YAML. It is stored internally
// in points (the canonical Asciidoctor PDF unit). Bare numeric YAML scalars
// are interpreted as points; strings may carry "mm", "in", or "pt" suffixes.
//
// The zero value represents "unset" so apply* helpers can decide whether to
// override the corresponding default.
type Length struct {
	pt  float64
	set bool
}

// Pt returns the length in points.
func (l Length) Pt() float64 { return l.pt }

// Mm returns the length in millimetres.
func (l Length) Mm() float64 { return l.pt * 25.4 / 72 }

// IsSet reports whether the length was present in the source YAML.
func (l Length) IsSet() bool { return l.set }

// UnmarshalYAML parses a numeric scalar (treated as pt) or a unit-suffixed
// string ("16mm", "0.5in", "24pt"). Returns an error on malformed input.
func (l *Length) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: length must be a scalar, got %v", value.Line, value.Kind)
	}
	pt, err := parseLengthPtStrict(value.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", value.Line, err)
	}
	l.pt = pt
	l.set = true
	return nil
}

// yamlColor is a typed colour value used in theme YAML. Stored internally as
// the bare 6-hex-digit form (no leading "#") that ECMA-376 §17.18.38
// (ST_HexColor) requires for `<w:color w:val=>`, `<w:shd w:fill=>`, and
// border-colour attributes; `auto` is the only other accepted literal.
//
// Normalisation lives at the YAML decode boundary so the renderer never has
// to think about whether a colour string carries a leading `#`.
type yamlColor struct {
	hex string
	set bool
}

// IsSet reports whether the colour was present in the source YAML.
func (c yamlColor) IsSet() bool { return c.set }

// String returns the normalised colour value (`RRGGBB` or `auto`).
func (c yamlColor) String() string { return c.hex }

// UnmarshalYAML accepts either a bare 6-hex string ("1A2B3C"), a
// `#`-prefixed hex string ("#1A2B3C"), or the literal "auto". Any other
// shape returns an error.
func (c *yamlColor) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: colour must be a scalar, got %v", value.Line, value.Kind)
	}
	s, err := normaliseHexColor(value.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", value.Line, err)
	}
	c.hex = s
	c.set = true
	return nil
}

// normaliseHexColor strips a single leading "#" and validates that the
// remainder is either six hex digits (returned upper-cased) or the literal
// "auto" (preserved as-is).
func normaliseHexColor(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("invalid colour: empty value")
	}
	if s == "auto" {
		return "auto", nil
	}
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return "", fmt.Errorf("invalid colour %q: expected 6 hex digits or \"auto\"", raw)
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return "", fmt.Errorf("invalid colour %q: expected 6 hex digits or \"auto\"", raw)
		}
	}
	return strings.ToUpper(s), nil
}

// parseLengthPtStrict parses a length string into points. Supports bare
// numbers (treated as pt) and the suffixes mm, in, pt.
func parseLengthPtStrict(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty length value")
	}
	for _, suf := range []struct {
		name   string
		toPt   float64
	}{
		{"mm", 72.0 / 25.4},
		{"in", 72.0},
		{"pt", 1.0},
	} {
		if strings.HasSuffix(s, suf.name) {
			f, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(s, suf.name)), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid length %q: %w", s, err)
			}
			return f * suf.toPt, nil
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid length %q: %w", s, err)
	}
	return f, nil
}

// --- YAML structures matching Asciidoctor PDF theme format ---

type yamlTheme struct {
	Page            *yamlPageTheme            `yaml:"page"`
	Base            *yamlBaseTheme            `yaml:"base"`
	Heading         *yamlHeadingTheme         `yaml:"heading"`
	Title           *yamlTitleTheme           `yaml:"title_page"`
	Table           *yamlTableTheme           `yaml:"table"`
	List            *yamlListTheme            `yaml:"list"`
	Code            *yamlCodeTheme            `yaml:"code"`
	Codespan        *yamlCodespanTheme        `yaml:"codespan"`
	Link            *yamlLinkTheme            `yaml:"link"`
	Prose           *yamlProseTheme           `yaml:"prose"`
	Quote           *yamlQuoteTheme           `yaml:"quote"`
	Admonition      *yamlAdmonitionTheme      `yaml:"admonition"`
	Sidebar         *yamlSidebarTheme         `yaml:"sidebar"`
	Example         *yamlExampleTheme         `yaml:"example"`
	Caption         *yamlCaptionTheme         `yaml:"caption"`
	DescriptionList *yamlDescriptionListTheme `yaml:"description_list"`
	Header          *yamlRunningHFTheme       `yaml:"header"`
	Footer          *yamlRunningHFTheme       `yaml:"footer"`
	Role            yamlRoleMap               `yaml:"role"`
}

type yamlPageTheme struct {
	Layout string      `yaml:"layout"`
	Size   string      `yaml:"size"`
	Margin yamlMargins `yaml:"margin"`
}

// yamlMargins handles the [25mm, 20mm, 20mm, 20mm] format.
type yamlMargins [4]float64

func (m *yamlMargins) UnmarshalYAML(value *yaml.Node) error {
	var raw []interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	for i := 0; i < 4 && i < len(raw); i++ {
		m[i] = parseLengthMM(raw[i])
	}
	return nil
}

type yamlBaseTheme struct {
	FontFamily string    `yaml:"font_family"`
	FontColor  yamlColor `yaml:"font_color"`
	FontSize   Length    `yaml:"font_size"`
	FontStyle  string    `yaml:"font_style"`
	TextAlign  string    `yaml:"text_align"`
	LineHeight float64   `yaml:"line_height"` // ratio, not a length
}

type yamlHeadingTheme struct {
	FontFamily    string            `yaml:"font_family"`
	FontColor     yamlColor         `yaml:"font_color"`
	FontStyle     string            `yaml:"font_style"`
	TextTransform string            `yaml:"text_transform"`
	MarginTop     Length            `yaml:"margin_top"`
	MarginBottom  Length            `yaml:"margin_bottom"`
	H1            *yamlHeadingLevel `yaml:"h1"`
	H2            *yamlHeadingLevel `yaml:"h2"`
	H3            *yamlHeadingLevel `yaml:"h3"`
	H4            *yamlHeadingLevel `yaml:"h4"`
	H5            *yamlHeadingLevel `yaml:"h5"`
	H6            *yamlHeadingLevel `yaml:"h6"`
}

type yamlHeadingLevel struct {
	FontSize      Length    `yaml:"font_size"`
	TextTransform string    `yaml:"text_transform"`
	FontColor     yamlColor `yaml:"font_color"`
	FontStyle     string    `yaml:"font_style"`
	MarginTop     Length    `yaml:"margin_top"`
	MarginBottom  Length    `yaml:"margin_bottom"`
}

type yamlTitleTheme struct {
	Title    *yamlTitleEntry `yaml:"title"`
	Subtitle *yamlTitleEntry `yaml:"subtitle"`
}

type yamlTitleEntry struct {
	FontSize      Length    `yaml:"font_size"`
	FontStyle     string    `yaml:"font_style"`
	FontColor     yamlColor `yaml:"font_color"`
	FontFamily    string    `yaml:"font_family"`
	TextTransform string    `yaml:"text_transform"`
	TextAlign     string    `yaml:"text_align"`
}

type yamlTableTheme struct {
	Width         string         `yaml:"width"` // "auto", "full", or percentage like "80%"
	FontSize      Length         `yaml:"font_size"`
	BorderColor   yamlColor      `yaml:"border_color"`
	BorderWidth   Length         `yaml:"border_width"`
	CellPadding   Length         `yaml:"cell_padding"`
	GridColor     yamlColor      `yaml:"grid_color"`
	GridWidth     Length         `yaml:"grid_width"`
	StripeBgColor yamlColor      `yaml:"stripe_background_color"`
	Head          *yamlTableHead `yaml:"head"`
	Foot          *yamlTableFoot `yaml:"foot"`
}

type yamlTableHead struct {
	BackgroundColor yamlColor `yaml:"background_color"`
	FontStyle       string    `yaml:"font_style"`
}

type yamlTableFoot struct {
	BackgroundColor yamlColor `yaml:"background_color"`
	FontStyle       string    `yaml:"font_style"`
}

type yamlListTheme struct {
	Indent          Length    `yaml:"indent"`
	ItemSpacing     Length    `yaml:"item_spacing"`
	MarkerFontColor yamlColor `yaml:"marker_font_color"`
}

type yamlCodeTheme struct {
	FontFamily      string    `yaml:"font_family"`
	FontSize        Length    `yaml:"font_size"`
	FontColor       yamlColor `yaml:"font_color"`
	BackgroundColor yamlColor `yaml:"background_color"`
	BorderColor     yamlColor `yaml:"border_color"`
	BorderWidth     Length    `yaml:"border_width"`
	LineHeight      float64   `yaml:"line_height"` // ratio, not a length
}

type yamlCodespanTheme struct {
	FontFamily      string    `yaml:"font_family"`
	FontSize        Length    `yaml:"font_size"`
	FontColor       yamlColor `yaml:"font_color"`
	BackgroundColor yamlColor `yaml:"background_color"`
}

type yamlLinkTheme struct {
	FontColor      yamlColor `yaml:"font_color"`
	FontStyle      string    `yaml:"font_style"`
	TextDecoration string    `yaml:"text_decoration"`
}

type yamlProseTheme struct {
	MarginBottom Length `yaml:"margin_bottom"`
	TextAlign    string `yaml:"text_align"`
	TextIndent   Length `yaml:"text_indent"`
}

type yamlQuoteTheme struct {
	FontSize    Length    `yaml:"font_size"`
	FontColor   yamlColor `yaml:"font_color"`
	FontStyle   string    `yaml:"font_style"`
	FontFamily  string    `yaml:"font_family"`
	BorderColor yamlColor `yaml:"border_color"`
	BorderWidth Length    `yaml:"border_width"`
}

type yamlAdmonitionTheme struct {
	FontColor       yamlColor            `yaml:"font_color"`
	FontSize        Length               `yaml:"font_size"`
	BackgroundColor yamlColor            `yaml:"background_color"`
	BorderColor     yamlColor            `yaml:"border_color"`
	BorderWidth     Length               `yaml:"border_width"`
	Label           *yamlAdmonitionLabel `yaml:"label"`
}

type yamlAdmonitionLabel struct {
	FontStyle     string    `yaml:"font_style"`
	FontColor     yamlColor `yaml:"font_color"`
	FontSize      Length    `yaml:"font_size"`
	TextTransform string    `yaml:"text_transform"`
}

type yamlSidebarTheme struct {
	BackgroundColor yamlColor `yaml:"background_color"`
	BorderColor     yamlColor `yaml:"border_color"`
	BorderWidth     Length    `yaml:"border_width"`
	FontColor       yamlColor `yaml:"font_color"`
	FontSize        Length    `yaml:"font_size"`
}

type yamlExampleTheme struct {
	BackgroundColor yamlColor `yaml:"background_color"`
	BorderColor     yamlColor `yaml:"border_color"`
	BorderWidth     Length    `yaml:"border_width"`
	FontColor       yamlColor `yaml:"font_color"`
	FontSize        Length    `yaml:"font_size"`
}

type yamlCaptionTheme struct {
	FontSize   Length    `yaml:"font_size"`
	FontStyle  string    `yaml:"font_style"`
	FontColor  yamlColor `yaml:"font_color"`
	FontFamily string    `yaml:"font_family"`
	TextAlign  string    `yaml:"text_align"`
}

type yamlDescriptionListTheme struct {
	TermFontStyle  string    `yaml:"term_font_style"`
	TermFontColor  yamlColor `yaml:"term_font_color"`
	TermFontFamily string    `yaml:"term_font_family"`
	TermFontSize   Length    `yaml:"term_font_size"`
}

// yamlRoleMap preserves source order so that downstream emission of role
// styles is deterministic. yaml.v3's default map decoding into a Go map drops
// insertion order; an ordered slice keeps it.
type yamlRoleMap struct {
	entries []yamlRoleEntry
}

type yamlRoleEntry struct {
	Name  string
	Theme yamlRoleTheme
}

func (m *yamlRoleMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: role must be a mapping", value.Line)
	}
	for i := 0; i+1 < len(value.Content); i += 2 {
		key := value.Content[i]
		val := value.Content[i+1]
		if val.Kind != yaml.MappingNode {
			continue
		}
		var rt yamlRoleTheme
		if err := val.Decode(&rt); err != nil {
			return fmt.Errorf("line %d: role %q: %w", key.Line, key.Value, err)
		}
		m.entries = append(m.entries, yamlRoleEntry{Name: key.Value, Theme: rt})
	}
	return nil
}

type yamlRoleTheme struct {
	FontFamily      string       `yaml:"font_family"`
	FontColor       yamlColor    `yaml:"font_color"`
	FontSize        Length       `yaml:"font_size"`
	FontStyle       string       `yaml:"font_style"`
	BackgroundColor yamlColor    `yaml:"background_color"`
	BorderColor     yamlColor    `yaml:"border_color"`
	BorderWidth     Length       `yaml:"border_width"`
	Padding         yamlPadding  `yaml:"padding"`
	MarginTop       Length       `yaml:"margin_top"`
	MarginBottom    Length       `yaml:"margin_bottom"`
	TextAlign       string       `yaml:"text_align"`
	TextTransform   string       `yaml:"text_transform"`
}

// yamlPadding accepts the CSS-style 1/2/3/4-element form Asciidoctor PDF uses
// (e.g. `padding: 4` = all sides; `padding: [4, 6]` = vertical/horizontal;
// `padding: [4, 6, 4, 6]` = top/right/bottom/left). All values are stored as
// points internally.
type yamlPadding struct {
	top, right, bottom, left float64
	set                      bool
}

// IsSet reports whether padding was supplied in the source YAML.
func (p yamlPadding) IsSet() bool { return p.set }

// Sides returns padding in t/r/b/l order, in points.
func (p yamlPadding) Sides() (top, right, bottom, left float64) {
	return p.top, p.right, p.bottom, p.left
}

func (p *yamlPadding) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		v, err := parseLengthPtStrict(value.Value)
		if err != nil {
			return fmt.Errorf("line %d: padding: %w", value.Line, err)
		}
		p.top, p.right, p.bottom, p.left = v, v, v, v
		p.set = true
		return nil
	case yaml.SequenceNode:
		var raw []string
		for _, n := range value.Content {
			raw = append(raw, n.Value)
		}
		vals := make([]float64, len(raw))
		for i, s := range raw {
			v, err := parseLengthPtStrict(s)
			if err != nil {
				return fmt.Errorf("line %d: padding[%d]: %w", value.Line, i, err)
			}
			vals[i] = v
		}
		switch len(vals) {
		case 1:
			p.top, p.right, p.bottom, p.left = vals[0], vals[0], vals[0], vals[0]
		case 2:
			p.top, p.bottom = vals[0], vals[0]
			p.right, p.left = vals[1], vals[1]
		case 3:
			p.top, p.bottom = vals[0], vals[2]
			p.right, p.left = vals[1], vals[1]
		case 4:
			p.top, p.right, p.bottom, p.left = vals[0], vals[1], vals[2], vals[3]
		default:
			return fmt.Errorf("line %d: padding must have 1-4 entries, got %d", value.Line, len(vals))
		}
		p.set = true
		return nil
	default:
		return fmt.Errorf("line %d: padding must be a scalar or sequence", value.Line)
	}
}

type yamlRunningHFTheme struct {
	Content    string             `yaml:"content"`
	FontSize   Length             `yaml:"font_size"`
	FontColor  yamlColor          `yaml:"font_color"`
	FontFamily string             `yaml:"font_family"`
	FontStyle  string             `yaml:"font_style"`
	Height     Length             `yaml:"height"`
	Recto      *yamlRunningHFSide `yaml:"recto"`
	Verso      *yamlRunningHFSide `yaml:"verso"`
}

type yamlRunningHFSide struct {
	Left   yamlRunningHFSlot `yaml:"left"`
	Center yamlRunningHFSlot `yaml:"center"`
	Right  yamlRunningHFSlot `yaml:"right"`
}

// yamlRunningHFSlot accepts either a bare string scalar
// (`left: '{revnumber}'`) or a mapping with a `content` key
// (`left: { content: '{revnumber}' }`); the Asciidoctor PDF theme uses the
// nested form, so support both to stay compatible with existing themes.
type yamlRunningHFSlot struct {
	value RunningHFSlot
}

func (s *yamlRunningHFSlot) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		s.value = NewRunningHFSlot(value.Value)
		return nil
	case yaml.MappingNode:
		var inner struct {
			Content string `yaml:"content"`
		}
		if err := value.Decode(&inner); err != nil {
			return err
		}
		s.value = NewRunningHFSlot(inner.Content)
		return nil
	default:
		return fmt.Errorf("line %d: running header/footer slot must be a string or mapping with 'content'", value.Line)
	}
}

// applyTo merges parsed YAML values onto the default theme.
// Only non-zero/non-empty values are applied.
func (yt *yamlTheme) applyTo(t *DocxTheme) {
	yt.applyPageTo(t)
	yt.applyBaseTo(t)
	yt.applyHeadingTo(t)
	yt.applyTitleTo(t)
	yt.applyTableTo(t)
	yt.applyListTo(t)
	yt.applyCodeTo(t)
	yt.applyCodespanTo(t)
	yt.applyLinkTo(t)
	yt.applyProseTo(t)
	yt.applyQuoteTo(t)
	yt.applyAdmonitionTo(t)
	yt.applySidebarTo(t)
	yt.applyExampleTo(t)
	yt.applyCaptionTo(t)
	yt.applyDescriptionListTo(t)
	applyRunningHF(yt.Header, &t.RunningHeader)
	applyRunningHF(yt.Footer, &t.RunningFooter)
	yt.applyRolesTo(t)
}

func (yt *yamlTheme) applyRolesTo(t *DocxTheme) {
	if t.Roles.byName == nil {
		t.Roles = NewRoleThemes()
	}
	for _, entry := range yt.Role.entries {
		role := RoleTheme{Name: entry.Name}
		yr := entry.Theme
		if yr.FontFamily != "" {
			role.FontFamily = yr.FontFamily
		}
		if yr.FontColor.IsSet() {
			role.FontColor = yr.FontColor.String()
		}
		if yr.FontSize.IsSet() {
			role.FontSize = yr.FontSize.Pt()
		}
		if yr.FontStyle != "" {
			role.FontStyle = yr.FontStyle
		}
		if yr.BackgroundColor.IsSet() {
			role.BackgroundColor = yr.BackgroundColor.String()
		}
		if yr.BorderColor.IsSet() {
			role.BorderColor = yr.BorderColor.String()
		}
		if yr.BorderWidth.IsSet() {
			role.BorderWidth = yr.BorderWidth.Pt()
		}
		if yr.Padding.IsSet() {
			role.PaddingTop, role.PaddingRight, role.PaddingBottom, role.PaddingLeft = yr.Padding.Sides()
		}
		if yr.MarginTop.IsSet() {
			role.MarginTop = yr.MarginTop.Pt()
		}
		if yr.MarginBottom.IsSet() {
			role.MarginBottom = yr.MarginBottom.Pt()
		}
		if yr.TextAlign != "" {
			role.TextAlign = yr.TextAlign
		}
		if yr.TextTransform != "" {
			role.TextTransform = yr.TextTransform
		}
		t.Roles.Set(role)
	}
}

func (yt *yamlTheme) applyPageTo(t *DocxTheme) {
	if yt.Page == nil {
		return
	}
	if yt.Page.Layout != "" {
		t.Page.Layout = yt.Page.Layout
	}
	if yt.Page.Size != "" {
		t.Page.Size = yt.Page.Size
	}
	if yt.Page.Margin != [4]float64{} {
		t.Page.Margin = yt.Page.Margin
	}
}

func (yt *yamlTheme) applyBaseTo(t *DocxTheme) {
	if yt.Base == nil {
		return
	}
	if yt.Base.FontFamily != "" {
		t.Base.FontFamily = yt.Base.FontFamily
	}
	if yt.Base.FontColor.IsSet() {
		t.Base.FontColor = yt.Base.FontColor.String()
	}
	if yt.Base.FontSize.IsSet() {
		t.Base.FontSize = yt.Base.FontSize.Pt()
	}
	if yt.Base.FontStyle != "" {
		t.Base.FontStyle = yt.Base.FontStyle
	}
	if yt.Base.TextAlign != "" {
		t.Base.TextAlign = yt.Base.TextAlign
	}
	if yt.Base.LineHeight > 0 {
		t.Base.LineHeight = yt.Base.LineHeight
	}
}

func (yt *yamlTheme) applyHeadingTo(t *DocxTheme) {
	if yt.Heading == nil {
		return
	}
	if yt.Heading.FontFamily != "" {
		t.Heading.FontFamily = yt.Heading.FontFamily
	}
	if yt.Heading.FontColor.IsSet() {
		t.Heading.FontColor = yt.Heading.FontColor.String()
	}
	if yt.Heading.FontStyle != "" {
		t.Heading.FontStyle = yt.Heading.FontStyle
	}
	if yt.Heading.TextTransform != "" {
		t.Heading.TextTransform = yt.Heading.TextTransform
	}
	if yt.Heading.MarginTop.IsSet() {
		t.Heading.MarginTop = yt.Heading.MarginTop.Pt()
	}
	if yt.Heading.MarginBottom.IsSet() {
		t.Heading.MarginBottom = yt.Heading.MarginBottom.Pt()
	}
	applyHeadingLevel(yt.Heading.H1, &t.Heading.H1FontSize, &t.Heading.H1TextTransform, &t.Heading.H1FontColor, &t.Heading.H1FontStyle, &t.Heading.H1MarginTop, &t.Heading.H1MarginBottom)
	applyHeadingLevel(yt.Heading.H2, &t.Heading.H2FontSize, &t.Heading.H2TextTransform, &t.Heading.H2FontColor, &t.Heading.H2FontStyle, &t.Heading.H2MarginTop, &t.Heading.H2MarginBottom)
	applyHeadingLevel(yt.Heading.H3, &t.Heading.H3FontSize, &t.Heading.H3TextTransform, &t.Heading.H3FontColor, &t.Heading.H3FontStyle, &t.Heading.H3MarginTop, &t.Heading.H3MarginBottom)
	applyHeadingLevel(yt.Heading.H4, &t.Heading.H4FontSize, &t.Heading.H4TextTransform, &t.Heading.H4FontColor, &t.Heading.H4FontStyle, &t.Heading.H4MarginTop, &t.Heading.H4MarginBottom)
	applyHeadingLevel(yt.Heading.H5, &t.Heading.H5FontSize, nil, &t.Heading.H5FontColor, &t.Heading.H5FontStyle, &t.Heading.H5MarginTop, &t.Heading.H5MarginBottom)
	applyHeadingLevel(yt.Heading.H6, &t.Heading.H6FontSize, nil, &t.Heading.H6FontColor, &t.Heading.H6FontStyle, &t.Heading.H6MarginTop, &t.Heading.H6MarginBottom)
}

func (yt *yamlTheme) applyTitleTo(t *DocxTheme) {
	if yt.Title == nil {
		return
	}
	if yt.Title.Title != nil {
		if yt.Title.Title.FontSize.IsSet() {
			t.Title.TitleFontSize = yt.Title.Title.FontSize.Pt()
		}
		if yt.Title.Title.FontStyle != "" {
			t.Title.TitleFontStyle = yt.Title.Title.FontStyle
		}
		if yt.Title.Title.FontColor.IsSet() {
			t.Title.TitleFontColor = yt.Title.Title.FontColor.String()
		}
		if yt.Title.Title.FontFamily != "" {
			t.Title.TitleFontFamily = yt.Title.Title.FontFamily
		}
		if yt.Title.Title.TextTransform != "" {
			t.Title.TitleTextTransform = yt.Title.Title.TextTransform
		}
		if yt.Title.Title.TextAlign != "" {
			t.Title.TitleTextAlign = yt.Title.Title.TextAlign
		}
	}
	if yt.Title.Subtitle != nil {
		if yt.Title.Subtitle.FontSize.IsSet() {
			t.Title.SubtitleFontSize = yt.Title.Subtitle.FontSize.Pt()
		}
		if yt.Title.Subtitle.FontColor.IsSet() {
			t.Title.SubtitleFontColor = yt.Title.Subtitle.FontColor.String()
		}
		if yt.Title.Subtitle.FontFamily != "" {
			t.Title.SubtitleFontFamily = yt.Title.Subtitle.FontFamily
		}
		if yt.Title.Subtitle.FontStyle != "" {
			t.Title.SubtitleFontStyle = yt.Title.Subtitle.FontStyle
		}
	}
}

func (yt *yamlTheme) applyTableTo(t *DocxTheme) {
	if yt.Table == nil {
		return
	}
	if yt.Table.Width != "" {
		t.Table.Width = yt.Table.Width
	}
	if yt.Table.FontSize.IsSet() {
		t.Table.FontSize = yt.Table.FontSize.Pt()
	}
	if yt.Table.BorderColor.IsSet() {
		t.Table.BorderColor = yt.Table.BorderColor.String()
	}
	if yt.Table.BorderWidth.IsSet() {
		t.Table.BorderWidth = yt.Table.BorderWidth.Pt()
	}
	if yt.Table.CellPadding.IsSet() {
		t.Table.CellPadding = yt.Table.CellPadding.Pt()
	}
	if yt.Table.GridColor.IsSet() {
		t.Table.GridColor = yt.Table.GridColor.String()
	}
	if yt.Table.GridWidth.IsSet() {
		t.Table.GridWidth = yt.Table.GridWidth.Pt()
	}
	if yt.Table.StripeBgColor.IsSet() {
		t.Table.StripeBgColor = yt.Table.StripeBgColor.String()
	}
	if yt.Table.Head != nil {
		if yt.Table.Head.BackgroundColor.IsSet() {
			t.Table.HeadBgColor = yt.Table.Head.BackgroundColor.String()
		}
		if yt.Table.Head.FontStyle != "" {
			t.Table.HeadFontStyle = yt.Table.Head.FontStyle
		}
	}
	if yt.Table.Foot != nil {
		if yt.Table.Foot.BackgroundColor.IsSet() {
			t.Table.FootBgColor = yt.Table.Foot.BackgroundColor.String()
		}
		if yt.Table.Foot.FontStyle != "" {
			t.Table.FootFontStyle = yt.Table.Foot.FontStyle
		}
	}
}

func (yt *yamlTheme) applyListTo(t *DocxTheme) {
	if yt.List == nil {
		return
	}
	if yt.List.Indent.IsSet() {
		t.List.Indent = yt.List.Indent.Pt()
	}
	if yt.List.ItemSpacing.IsSet() {
		t.List.ItemSpacing = yt.List.ItemSpacing.Pt()
	}
	if yt.List.MarkerFontColor.IsSet() {
		t.List.MarkerFontColor = yt.List.MarkerFontColor.String()
	}
}

func (yt *yamlTheme) applyCodeTo(t *DocxTheme) {
	if yt.Code == nil {
		return
	}
	if yt.Code.FontFamily != "" {
		t.Code.FontFamily = yt.Code.FontFamily
	}
	if yt.Code.FontSize.IsSet() {
		t.Code.FontSize = yt.Code.FontSize.Pt()
	}
	if yt.Code.FontColor.IsSet() {
		t.Code.FontColor = yt.Code.FontColor.String()
	}
	if yt.Code.BackgroundColor.IsSet() {
		t.Code.BackgroundColor = yt.Code.BackgroundColor.String()
	}
	if yt.Code.BorderColor.IsSet() {
		t.Code.BorderColor = yt.Code.BorderColor.String()
	}
	if yt.Code.BorderWidth.IsSet() {
		t.Code.BorderWidth = yt.Code.BorderWidth.Pt()
	}
	if yt.Code.LineHeight > 0 {
		t.Code.LineHeight = yt.Code.LineHeight
	}
}

func (yt *yamlTheme) applyCodespanTo(t *DocxTheme) {
	if yt.Codespan == nil {
		return
	}
	if yt.Codespan.FontFamily != "" {
		t.Codespan.FontFamily = yt.Codespan.FontFamily
	}
	if yt.Codespan.FontSize.IsSet() {
		t.Codespan.FontSize = yt.Codespan.FontSize.Pt()
	}
	if yt.Codespan.FontColor.IsSet() {
		t.Codespan.FontColor = yt.Codespan.FontColor.String()
	}
	if yt.Codespan.BackgroundColor.IsSet() {
		t.Codespan.BackgroundColor = yt.Codespan.BackgroundColor.String()
	}
}

func (yt *yamlTheme) applyLinkTo(t *DocxTheme) {
	if yt.Link == nil {
		return
	}
	if yt.Link.FontColor.IsSet() {
		t.Link.FontColor = yt.Link.FontColor.String()
	}
	if yt.Link.FontStyle != "" {
		t.Link.FontStyle = yt.Link.FontStyle
	}
	if yt.Link.TextDecoration != "" {
		t.Link.TextDecoration = yt.Link.TextDecoration
	}
}

func (yt *yamlTheme) applyProseTo(t *DocxTheme) {
	if yt.Prose == nil {
		return
	}
	if yt.Prose.MarginBottom.IsSet() {
		t.Prose.MarginBottom = yt.Prose.MarginBottom.Pt()
	}
	if yt.Prose.TextAlign != "" {
		t.Prose.TextAlign = yt.Prose.TextAlign
	}
	if yt.Prose.TextIndent.IsSet() {
		t.Prose.TextIndent = yt.Prose.TextIndent.Pt()
	}
}

func (yt *yamlTheme) applyQuoteTo(t *DocxTheme) {
	if yt.Quote == nil {
		return
	}
	if yt.Quote.FontSize.IsSet() {
		t.Quote.FontSize = yt.Quote.FontSize.Pt()
	}
	if yt.Quote.FontColor.IsSet() {
		t.Quote.FontColor = yt.Quote.FontColor.String()
	}
	if yt.Quote.FontStyle != "" {
		t.Quote.FontStyle = yt.Quote.FontStyle
	}
	if yt.Quote.FontFamily != "" {
		t.Quote.FontFamily = yt.Quote.FontFamily
	}
	if yt.Quote.BorderColor.IsSet() {
		t.Quote.BorderColor = yt.Quote.BorderColor.String()
	}
	if yt.Quote.BorderWidth.IsSet() {
		t.Quote.BorderWidth = yt.Quote.BorderWidth.Pt()
	}
}

func (yt *yamlTheme) applyAdmonitionTo(t *DocxTheme) {
	if yt.Admonition == nil {
		return
	}
	if yt.Admonition.FontColor.IsSet() {
		t.Admonition.FontColor = yt.Admonition.FontColor.String()
	}
	if yt.Admonition.FontSize.IsSet() {
		t.Admonition.FontSize = yt.Admonition.FontSize.Pt()
	}
	if yt.Admonition.BackgroundColor.IsSet() {
		t.Admonition.BackgroundColor = yt.Admonition.BackgroundColor.String()
	}
	if yt.Admonition.BorderColor.IsSet() {
		t.Admonition.BorderColor = yt.Admonition.BorderColor.String()
	}
	if yt.Admonition.BorderWidth.IsSet() {
		t.Admonition.BorderWidth = yt.Admonition.BorderWidth.Pt()
	}
	if yt.Admonition.Label != nil {
		if yt.Admonition.Label.FontStyle != "" {
			t.Admonition.LabelFontStyle = yt.Admonition.Label.FontStyle
		}
		if yt.Admonition.Label.FontColor.IsSet() {
			t.Admonition.LabelFontColor = yt.Admonition.Label.FontColor.String()
		}
		if yt.Admonition.Label.FontSize.IsSet() {
			t.Admonition.LabelFontSize = yt.Admonition.Label.FontSize.Pt()
		}
		if yt.Admonition.Label.TextTransform != "" {
			t.Admonition.LabelTextTransform = yt.Admonition.Label.TextTransform
		}
	}
}

func (yt *yamlTheme) applySidebarTo(t *DocxTheme) {
	if yt.Sidebar == nil {
		return
	}
	if yt.Sidebar.BackgroundColor.IsSet() {
		t.Sidebar.BackgroundColor = yt.Sidebar.BackgroundColor.String()
	}
	if yt.Sidebar.BorderColor.IsSet() {
		t.Sidebar.BorderColor = yt.Sidebar.BorderColor.String()
	}
	if yt.Sidebar.BorderWidth.IsSet() {
		t.Sidebar.BorderWidth = yt.Sidebar.BorderWidth.Pt()
	}
	if yt.Sidebar.FontColor.IsSet() {
		t.Sidebar.FontColor = yt.Sidebar.FontColor.String()
	}
	if yt.Sidebar.FontSize.IsSet() {
		t.Sidebar.FontSize = yt.Sidebar.FontSize.Pt()
	}
}

func (yt *yamlTheme) applyExampleTo(t *DocxTheme) {
	if yt.Example == nil {
		return
	}
	if yt.Example.BackgroundColor.IsSet() {
		t.Example.BackgroundColor = yt.Example.BackgroundColor.String()
	}
	if yt.Example.BorderColor.IsSet() {
		t.Example.BorderColor = yt.Example.BorderColor.String()
	}
	if yt.Example.BorderWidth.IsSet() {
		t.Example.BorderWidth = yt.Example.BorderWidth.Pt()
	}
	if yt.Example.FontColor.IsSet() {
		t.Example.FontColor = yt.Example.FontColor.String()
	}
	if yt.Example.FontSize.IsSet() {
		t.Example.FontSize = yt.Example.FontSize.Pt()
	}
}

func (yt *yamlTheme) applyCaptionTo(t *DocxTheme) {
	if yt.Caption == nil {
		return
	}
	if yt.Caption.FontSize.IsSet() {
		t.Caption.FontSize = yt.Caption.FontSize.Pt()
	}
	if yt.Caption.FontStyle != "" {
		t.Caption.FontStyle = yt.Caption.FontStyle
	}
	if yt.Caption.FontColor.IsSet() {
		t.Caption.FontColor = yt.Caption.FontColor.String()
	}
	if yt.Caption.FontFamily != "" {
		t.Caption.FontFamily = yt.Caption.FontFamily
	}
	if yt.Caption.TextAlign != "" {
		t.Caption.TextAlign = yt.Caption.TextAlign
	}
}

func (yt *yamlTheme) applyDescriptionListTo(t *DocxTheme) {
	if yt.DescriptionList == nil {
		return
	}
	if yt.DescriptionList.TermFontStyle != "" {
		t.DescriptionList.TermFontStyle = yt.DescriptionList.TermFontStyle
	}
	if yt.DescriptionList.TermFontColor.IsSet() {
		t.DescriptionList.TermFontColor = yt.DescriptionList.TermFontColor.String()
	}
	if yt.DescriptionList.TermFontFamily != "" {
		t.DescriptionList.TermFontFamily = yt.DescriptionList.TermFontFamily
	}
	if yt.DescriptionList.TermFontSize.IsSet() {
		t.DescriptionList.TermFontSize = yt.DescriptionList.TermFontSize.Pt()
	}
}

func applyHeadingLevel(src *yamlHeadingLevel, fontSize *float64, textTransform, fontColor, fontStyle *string, marginTop, marginBottom *float64) {
	if src == nil {
		return
	}
	if src.FontSize.IsSet() {
		*fontSize = src.FontSize.Pt()
	}
	if textTransform != nil && src.TextTransform != "" {
		*textTransform = src.TextTransform
	}
	if src.FontColor.IsSet() {
		*fontColor = src.FontColor.String()
	}
	if src.FontStyle != "" {
		*fontStyle = src.FontStyle
	}
	if src.MarginTop.IsSet() {
		*marginTop = src.MarginTop.Pt()
	}
	if src.MarginBottom.IsSet() {
		*marginBottom = src.MarginBottom.Pt()
	}
}

func applyRunningHF(src *yamlRunningHFTheme, dst *RunningHFTheme) {
	if src == nil {
		return
	}
	if src.Content != "" {
		dst.Content = src.Content
	}
	if src.FontSize.IsSet() {
		dst.FontSize = src.FontSize.Pt()
	}
	if src.FontColor.IsSet() {
		dst.FontColor = src.FontColor.String()
	}
	if src.FontFamily != "" {
		dst.FontFamily = src.FontFamily
	}
	if src.FontStyle != "" {
		dst.FontStyle = src.FontStyle
	}
	if src.Height.IsSet() {
		dst.Height = src.Height.Mm()
	}
	applyRunningHFSide(src.Recto, &dst.Recto)
	applyRunningHFSide(src.Verso, &dst.Verso)
}

func applyRunningHFSide(src *yamlRunningHFSide, dst *RunningHFSide) {
	if src == nil {
		return
	}
	if src.Left.value.IsSet() {
		dst.Left = src.Left.value
	}
	if src.Center.value.IsSet() {
		dst.Center = src.Center.value
	}
	if src.Right.value.IsSet() {
		dst.Right = src.Right.value
	}
}

// parseLengthMM extracts a numeric value in millimeters from a YAML margin entry.
// Handles numeric values (treated as mm) and string values with units ("25mm", "1in", "72pt").
func parseLengthMM(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	case string:
		s := strings.TrimSpace(val)
		if strings.HasSuffix(s, "mm") {
			f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "mm"), 64)
			return f
		}
		if strings.HasSuffix(s, "in") {
			f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "in"), 64)
			return f * 25.4 // inches to mm
		}
		if strings.HasSuffix(s, "pt") {
			f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "pt"), 64)
			return f * 25.4 / 72 // points to mm
		}
		// bare number, treat as mm
		f, _ := strconv.ParseFloat(s, 64)
		return f
	default:
		return 0
	}
}
