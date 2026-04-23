package docx

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
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
	var yt yamlTheme
	if err := yaml.Unmarshal(data, &yt); err != nil {
		return nil, err
	}
	theme := DefaultTheme()
	yt.applyTo(theme)
	return theme, nil
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
}

type yamlPageTheme struct {
	Layout string      `yaml:"layout"`
	Size   string      `yaml:"size"`
	Margin yamlMargins `yaml:"margin"`
}

// yamlMargins handles the [25mm, 20mm, 20mm, 20mm] format.
type yamlMargins [4]float64

func (m *yamlMargins) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw []interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	for i := 0; i < 4 && i < len(raw); i++ {
		m[i] = parseLengthMM(raw[i])
	}
	return nil
}

type yamlBaseTheme struct {
	FontFamily string  `yaml:"font_family"`
	FontColor  string  `yaml:"font_color"`
	FontSize   float64 `yaml:"font_size"`
	FontStyle  string  `yaml:"font_style"`
	TextAlign  string  `yaml:"text_align"`
	LineHeight float64 `yaml:"line_height"`
}

type yamlHeadingTheme struct {
	FontFamily    string            `yaml:"font_family"`
	FontColor     string            `yaml:"font_color"`
	FontStyle     string            `yaml:"font_style"`
	TextTransform string            `yaml:"text_transform"`
	MarginTop     float64           `yaml:"margin_top"`
	MarginBottom  float64           `yaml:"margin_bottom"`
	H1            *yamlHeadingLevel `yaml:"h1"`
	H2            *yamlHeadingLevel `yaml:"h2"`
	H3            *yamlHeadingLevel `yaml:"h3"`
	H4            *yamlHeadingLevel `yaml:"h4"`
	H5            *yamlHeadingLevel `yaml:"h5"`
	H6            *yamlHeadingLevel `yaml:"h6"`
}

type yamlHeadingLevel struct {
	FontSize      float64 `yaml:"font_size"`
	TextTransform string  `yaml:"text_transform"`
	FontColor     string  `yaml:"font_color"`
	FontStyle     string  `yaml:"font_style"`
}

type yamlTitleTheme struct {
	Title    *yamlTitleEntry `yaml:"title"`
	Subtitle *yamlTitleEntry `yaml:"subtitle"`
}

type yamlTitleEntry struct {
	FontSize      float64 `yaml:"font_size"`
	FontStyle     string  `yaml:"font_style"`
	FontColor     string  `yaml:"font_color"`
	FontFamily    string  `yaml:"font_family"`
	TextTransform string  `yaml:"text_transform"`
	TextAlign     string  `yaml:"text_align"`
}

type yamlTableTheme struct {
	Width         string         `yaml:"width"` // "auto", "full", or percentage like "80%"
	FontSize      float64        `yaml:"font_size"`
	BorderColor   string         `yaml:"border_color"`
	BorderWidth   float64        `yaml:"border_width"`
	CellPadding   float64        `yaml:"cell_padding"`
	GridColor     string         `yaml:"grid_color"`
	GridWidth     float64        `yaml:"grid_width"`
	StripeBgColor string         `yaml:"stripe_background_color"`
	Head          *yamlTableHead `yaml:"head"`
	Foot          *yamlTableFoot `yaml:"foot"`
}

type yamlTableHead struct {
	BackgroundColor string `yaml:"background_color"`
	FontStyle       string `yaml:"font_style"`
}

type yamlTableFoot struct {
	BackgroundColor string `yaml:"background_color"`
	FontStyle       string `yaml:"font_style"`
}

type yamlListTheme struct {
	Indent          interface{} `yaml:"indent"` // supports: 24 (pt), "10mm", "0.5in", "24pt"
	ItemSpacing     float64     `yaml:"item_spacing"`
	MarkerFontColor string      `yaml:"marker_font_color"`
}

type yamlCodeTheme struct {
	FontFamily      string  `yaml:"font_family"`
	FontSize        float64 `yaml:"font_size"`
	FontColor       string  `yaml:"font_color"`
	BackgroundColor string  `yaml:"background_color"`
	BorderColor     string  `yaml:"border_color"`
	BorderWidth     float64 `yaml:"border_width"`
	LineHeight      float64 `yaml:"line_height"`
}

type yamlCodespanTheme struct {
	FontFamily      string  `yaml:"font_family"`
	FontSize        float64 `yaml:"font_size"`
	FontColor       string  `yaml:"font_color"`
	BackgroundColor string  `yaml:"background_color"`
}

type yamlLinkTheme struct {
	FontColor      string `yaml:"font_color"`
	FontStyle      string `yaml:"font_style"`
	TextDecoration string `yaml:"text_decoration"`
}

type yamlProseTheme struct {
	MarginBottom float64 `yaml:"margin_bottom"`
	TextAlign    string  `yaml:"text_align"`
	TextIndent   float64 `yaml:"text_indent"`
}

type yamlQuoteTheme struct {
	FontSize    float64 `yaml:"font_size"`
	FontColor   string  `yaml:"font_color"`
	FontStyle   string  `yaml:"font_style"`
	FontFamily  string  `yaml:"font_family"`
	BorderColor string  `yaml:"border_color"`
	BorderWidth float64 `yaml:"border_width"`
}

type yamlAdmonitionTheme struct {
	FontColor       string               `yaml:"font_color"`
	FontSize        float64              `yaml:"font_size"`
	BackgroundColor string               `yaml:"background_color"`
	BorderColor     string               `yaml:"border_color"`
	BorderWidth     float64              `yaml:"border_width"`
	Label           *yamlAdmonitionLabel `yaml:"label"`
}

type yamlAdmonitionLabel struct {
	FontStyle string `yaml:"font_style"`
	FontColor string `yaml:"font_color"`
}

type yamlSidebarTheme struct {
	BackgroundColor string  `yaml:"background_color"`
	BorderColor     string  `yaml:"border_color"`
	BorderWidth     float64 `yaml:"border_width"`
	FontColor       string  `yaml:"font_color"`
	FontSize        float64 `yaml:"font_size"`
}

type yamlExampleTheme struct {
	BackgroundColor string  `yaml:"background_color"`
	BorderColor     string  `yaml:"border_color"`
	BorderWidth     float64 `yaml:"border_width"`
	FontColor       string  `yaml:"font_color"`
	FontSize        float64 `yaml:"font_size"`
}

type yamlCaptionTheme struct {
	FontSize   float64 `yaml:"font_size"`
	FontStyle  string  `yaml:"font_style"`
	FontColor  string  `yaml:"font_color"`
	FontFamily string  `yaml:"font_family"`
	TextAlign  string  `yaml:"text_align"`
}

type yamlDescriptionListTheme struct {
	TermFontStyle  string  `yaml:"term_font_style"`
	TermFontColor  string  `yaml:"term_font_color"`
	TermFontFamily string  `yaml:"term_font_family"`
	TermFontSize   float64 `yaml:"term_font_size"`
}

type yamlRunningHFTheme struct {
	Content    string  `yaml:"content"`
	FontSize   float64 `yaml:"font_size"`
	FontColor  string  `yaml:"font_color"`
	FontFamily string  `yaml:"font_family"`
	FontStyle  string  `yaml:"font_style"`
	Height     float64 `yaml:"height"`
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
	if yt.Base.FontColor != "" {
		t.Base.FontColor = yt.Base.FontColor
	}
	if yt.Base.FontSize > 0 {
		t.Base.FontSize = yt.Base.FontSize
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
	if yt.Heading.FontColor != "" {
		t.Heading.FontColor = yt.Heading.FontColor
	}
	if yt.Heading.FontStyle != "" {
		t.Heading.FontStyle = yt.Heading.FontStyle
	}
	if yt.Heading.TextTransform != "" {
		t.Heading.TextTransform = yt.Heading.TextTransform
	}
	if yt.Heading.MarginTop > 0 {
		t.Heading.MarginTop = yt.Heading.MarginTop
	}
	if yt.Heading.MarginBottom > 0 {
		t.Heading.MarginBottom = yt.Heading.MarginBottom
	}
	applyHeadingLevel(yt.Heading.H1, &t.Heading.H1FontSize, &t.Heading.H1TextTransform, &t.Heading.H1FontColor, &t.Heading.H1FontStyle)
	applyHeadingLevel(yt.Heading.H2, &t.Heading.H2FontSize, &t.Heading.H2TextTransform, &t.Heading.H2FontColor, &t.Heading.H2FontStyle)
	applyHeadingLevel(yt.Heading.H3, &t.Heading.H3FontSize, &t.Heading.H3TextTransform, &t.Heading.H3FontColor, &t.Heading.H3FontStyle)
	applyHeadingLevel(yt.Heading.H4, &t.Heading.H4FontSize, &t.Heading.H4TextTransform, &t.Heading.H4FontColor, &t.Heading.H4FontStyle)
	applyHeadingLevel(yt.Heading.H5, &t.Heading.H5FontSize, nil, &t.Heading.H5FontColor, &t.Heading.H5FontStyle)
	applyHeadingLevel(yt.Heading.H6, &t.Heading.H6FontSize, nil, &t.Heading.H6FontColor, &t.Heading.H6FontStyle)
}

func (yt *yamlTheme) applyTitleTo(t *DocxTheme) {
	if yt.Title == nil {
		return
	}
	if yt.Title.Title != nil {
		if yt.Title.Title.FontSize > 0 {
			t.Title.TitleFontSize = yt.Title.Title.FontSize
		}
		if yt.Title.Title.FontStyle != "" {
			t.Title.TitleFontStyle = yt.Title.Title.FontStyle
		}
		if yt.Title.Title.FontColor != "" {
			t.Title.TitleFontColor = yt.Title.Title.FontColor
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
		if yt.Title.Subtitle.FontSize > 0 {
			t.Title.SubtitleFontSize = yt.Title.Subtitle.FontSize
		}
		if yt.Title.Subtitle.FontColor != "" {
			t.Title.SubtitleFontColor = yt.Title.Subtitle.FontColor
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
	if yt.Table.FontSize > 0 {
		t.Table.FontSize = yt.Table.FontSize
	}
	if yt.Table.BorderColor != "" {
		t.Table.BorderColor = yt.Table.BorderColor
	}
	if yt.Table.BorderWidth > 0 {
		t.Table.BorderWidth = yt.Table.BorderWidth
	}
	if yt.Table.CellPadding > 0 {
		t.Table.CellPadding = yt.Table.CellPadding
	}
	if yt.Table.GridColor != "" {
		t.Table.GridColor = yt.Table.GridColor
	}
	if yt.Table.GridWidth > 0 {
		t.Table.GridWidth = yt.Table.GridWidth
	}
	if yt.Table.StripeBgColor != "" {
		t.Table.StripeBgColor = yt.Table.StripeBgColor
	}
	if yt.Table.Head != nil {
		if yt.Table.Head.BackgroundColor != "" {
			t.Table.HeadBgColor = yt.Table.Head.BackgroundColor
		}
		if yt.Table.Head.FontStyle != "" {
			t.Table.HeadFontStyle = yt.Table.Head.FontStyle
		}
	}
	if yt.Table.Foot != nil {
		if yt.Table.Foot.BackgroundColor != "" {
			t.Table.FootBgColor = yt.Table.Foot.BackgroundColor
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
	if yt.List.Indent != nil {
		if v := parseLengthPt(yt.List.Indent); v > 0 {
			t.List.Indent = v
		}
	}
	if yt.List.ItemSpacing > 0 {
		t.List.ItemSpacing = yt.List.ItemSpacing
	}
	if yt.List.MarkerFontColor != "" {
		t.List.MarkerFontColor = yt.List.MarkerFontColor
	}
}

func (yt *yamlTheme) applyCodeTo(t *DocxTheme) {
	if yt.Code == nil {
		return
	}
	if yt.Code.FontFamily != "" {
		t.Code.FontFamily = yt.Code.FontFamily
	}
	if yt.Code.FontSize > 0 {
		t.Code.FontSize = yt.Code.FontSize
	}
	if yt.Code.FontColor != "" {
		t.Code.FontColor = yt.Code.FontColor
	}
	if yt.Code.BackgroundColor != "" {
		t.Code.BackgroundColor = yt.Code.BackgroundColor
	}
	if yt.Code.BorderColor != "" {
		t.Code.BorderColor = yt.Code.BorderColor
	}
	if yt.Code.BorderWidth > 0 {
		t.Code.BorderWidth = yt.Code.BorderWidth
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
	if yt.Codespan.FontSize > 0 {
		t.Codespan.FontSize = yt.Codespan.FontSize
	}
	if yt.Codespan.FontColor != "" {
		t.Codespan.FontColor = yt.Codespan.FontColor
	}
	if yt.Codespan.BackgroundColor != "" {
		t.Codespan.BackgroundColor = yt.Codespan.BackgroundColor
	}
}

func (yt *yamlTheme) applyLinkTo(t *DocxTheme) {
	if yt.Link == nil {
		return
	}
	if yt.Link.FontColor != "" {
		t.Link.FontColor = yt.Link.FontColor
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
	if yt.Prose.MarginBottom > 0 {
		t.Prose.MarginBottom = yt.Prose.MarginBottom
	}
	if yt.Prose.TextAlign != "" {
		t.Prose.TextAlign = yt.Prose.TextAlign
	}
	if yt.Prose.TextIndent > 0 {
		t.Prose.TextIndent = yt.Prose.TextIndent
	}
}

func (yt *yamlTheme) applyQuoteTo(t *DocxTheme) {
	if yt.Quote == nil {
		return
	}
	if yt.Quote.FontSize > 0 {
		t.Quote.FontSize = yt.Quote.FontSize
	}
	if yt.Quote.FontColor != "" {
		t.Quote.FontColor = yt.Quote.FontColor
	}
	if yt.Quote.FontStyle != "" {
		t.Quote.FontStyle = yt.Quote.FontStyle
	}
	if yt.Quote.FontFamily != "" {
		t.Quote.FontFamily = yt.Quote.FontFamily
	}
	if yt.Quote.BorderColor != "" {
		t.Quote.BorderColor = yt.Quote.BorderColor
	}
	if yt.Quote.BorderWidth > 0 {
		t.Quote.BorderWidth = yt.Quote.BorderWidth
	}
}

func (yt *yamlTheme) applyAdmonitionTo(t *DocxTheme) {
	if yt.Admonition == nil {
		return
	}
	if yt.Admonition.FontColor != "" {
		t.Admonition.FontColor = yt.Admonition.FontColor
	}
	if yt.Admonition.FontSize > 0 {
		t.Admonition.FontSize = yt.Admonition.FontSize
	}
	if yt.Admonition.BackgroundColor != "" {
		t.Admonition.BackgroundColor = yt.Admonition.BackgroundColor
	}
	if yt.Admonition.BorderColor != "" {
		t.Admonition.BorderColor = yt.Admonition.BorderColor
	}
	if yt.Admonition.BorderWidth > 0 {
		t.Admonition.BorderWidth = yt.Admonition.BorderWidth
	}
	if yt.Admonition.Label != nil {
		if yt.Admonition.Label.FontStyle != "" {
			t.Admonition.LabelFontStyle = yt.Admonition.Label.FontStyle
		}
		if yt.Admonition.Label.FontColor != "" {
			t.Admonition.LabelFontColor = yt.Admonition.Label.FontColor
		}
	}
}

func (yt *yamlTheme) applySidebarTo(t *DocxTheme) {
	if yt.Sidebar == nil {
		return
	}
	if yt.Sidebar.BackgroundColor != "" {
		t.Sidebar.BackgroundColor = yt.Sidebar.BackgroundColor
	}
	if yt.Sidebar.BorderColor != "" {
		t.Sidebar.BorderColor = yt.Sidebar.BorderColor
	}
	if yt.Sidebar.BorderWidth > 0 {
		t.Sidebar.BorderWidth = yt.Sidebar.BorderWidth
	}
	if yt.Sidebar.FontColor != "" {
		t.Sidebar.FontColor = yt.Sidebar.FontColor
	}
	if yt.Sidebar.FontSize > 0 {
		t.Sidebar.FontSize = yt.Sidebar.FontSize
	}
}

func (yt *yamlTheme) applyExampleTo(t *DocxTheme) {
	if yt.Example == nil {
		return
	}
	if yt.Example.BackgroundColor != "" {
		t.Example.BackgroundColor = yt.Example.BackgroundColor
	}
	if yt.Example.BorderColor != "" {
		t.Example.BorderColor = yt.Example.BorderColor
	}
	if yt.Example.BorderWidth > 0 {
		t.Example.BorderWidth = yt.Example.BorderWidth
	}
	if yt.Example.FontColor != "" {
		t.Example.FontColor = yt.Example.FontColor
	}
	if yt.Example.FontSize > 0 {
		t.Example.FontSize = yt.Example.FontSize
	}
}

func (yt *yamlTheme) applyCaptionTo(t *DocxTheme) {
	if yt.Caption == nil {
		return
	}
	if yt.Caption.FontSize > 0 {
		t.Caption.FontSize = yt.Caption.FontSize
	}
	if yt.Caption.FontStyle != "" {
		t.Caption.FontStyle = yt.Caption.FontStyle
	}
	if yt.Caption.FontColor != "" {
		t.Caption.FontColor = yt.Caption.FontColor
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
	if yt.DescriptionList.TermFontColor != "" {
		t.DescriptionList.TermFontColor = yt.DescriptionList.TermFontColor
	}
	if yt.DescriptionList.TermFontFamily != "" {
		t.DescriptionList.TermFontFamily = yt.DescriptionList.TermFontFamily
	}
	if yt.DescriptionList.TermFontSize > 0 {
		t.DescriptionList.TermFontSize = yt.DescriptionList.TermFontSize
	}
}

func applyHeadingLevel(src *yamlHeadingLevel, fontSize *float64, textTransform, fontColor, fontStyle *string) {
	if src == nil {
		return
	}
	if src.FontSize > 0 {
		*fontSize = src.FontSize
	}
	if textTransform != nil && src.TextTransform != "" {
		*textTransform = src.TextTransform
	}
	if src.FontColor != "" {
		*fontColor = src.FontColor
	}
	if src.FontStyle != "" {
		*fontStyle = src.FontStyle
	}
}

func applyRunningHF(src *yamlRunningHFTheme, dst *RunningHFTheme) {
	if src == nil {
		return
	}
	if src.Content != "" {
		dst.Content = src.Content
	}
	if src.FontSize > 0 {
		dst.FontSize = src.FontSize
	}
	if src.FontColor != "" {
		dst.FontColor = src.FontColor
	}
	if src.FontFamily != "" {
		dst.FontFamily = src.FontFamily
	}
	if src.FontStyle != "" {
		dst.FontStyle = src.FontStyle
	}
	if src.Height > 0 {
		dst.Height = src.Height
	}
}

// parseLengthMM extracts a numeric value in millimeters from a YAML margin entry.
// Handles numeric values (treated as mm) and string values with units ("25mm", "1in", "72pt").
// parseLengthPt extracts a numeric value in points from a YAML entry.
// Bare numbers are treated as pt. Supports "10mm", "0.5in", "24pt".
func parseLengthPt(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	case string:
		s := strings.TrimSpace(val)
		if strings.HasSuffix(s, "mm") {
			f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "mm"), 64)
			return f / 25.4 * 72 // mm to pt
		}
		if strings.HasSuffix(s, "in") {
			f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "in"), 64)
			return f * 72 // in to pt
		}
		if strings.HasSuffix(s, "pt") {
			f, _ := strconv.ParseFloat(strings.TrimSuffix(s, "pt"), 64)
			return f
		}
		f, _ := strconv.ParseFloat(s, 64)
		return f
	default:
		return 0
	}
}

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
