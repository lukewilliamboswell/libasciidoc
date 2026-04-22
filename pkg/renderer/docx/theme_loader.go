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
	Page    *yamlPageTheme    `yaml:"page"`
	Base    *yamlBaseTheme    `yaml:"base"`
	Heading *yamlHeadingTheme `yaml:"heading"`
	Title   *yamlTitleTheme   `yaml:"title_page"`
	Table   *yamlTableTheme   `yaml:"table"`
	List    *yamlListTheme    `yaml:"list"`
	Code    *yamlCodeTheme    `yaml:"code"`
	Link    *yamlLinkTheme    `yaml:"link"`
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
}

type yamlHeadingTheme struct {
	FontFamily    string            `yaml:"font_family"`
	FontColor     string            `yaml:"font_color"`
	FontStyle     string            `yaml:"font_style"`
	TextTransform string            `yaml:"text_transform"`
	H1            *yamlHeadingLevel `yaml:"h1"`
	H2            *yamlHeadingLevel `yaml:"h2"`
	H3            *yamlHeadingLevel `yaml:"h3"`
	H4            *yamlHeadingLevel `yaml:"h4"`
}

type yamlHeadingLevel struct {
	FontSize      float64 `yaml:"font_size"`
	TextTransform string  `yaml:"text_transform"`
}

type yamlTitleTheme struct {
	Title    *yamlTitleEntry `yaml:"title"`
	Subtitle *yamlTitleEntry `yaml:"subtitle"`
}

type yamlTitleEntry struct {
	FontSize  float64 `yaml:"font_size"`
	FontStyle string  `yaml:"font_style"`
	FontColor string  `yaml:"font_color"`
}

type yamlTableTheme struct {
	FontSize    float64        `yaml:"font_size"`
	BorderColor string         `yaml:"border_color"`
	BorderWidth float64        `yaml:"border_width"`
	Head        *yamlTableHead `yaml:"head"`
}

type yamlTableHead struct {
	BackgroundColor string `yaml:"background_color"`
}

type yamlListTheme struct {
	Indent float64 `yaml:"indent"`
}

type yamlCodeTheme struct {
	FontFamily string  `yaml:"font_family"`
	FontSize   float64 `yaml:"font_size"`
}

type yamlLinkTheme struct {
	FontColor string `yaml:"font_color"`
}

// applyTo merges parsed YAML values onto the default theme.
// Only non-zero/non-empty values are applied.
func (yt *yamlTheme) applyTo(t *DocxTheme) {
	if yt.Page != nil {
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
	if yt.Base != nil {
		if yt.Base.FontFamily != "" {
			t.Base.FontFamily = yt.Base.FontFamily
		}
		if yt.Base.FontColor != "" {
			t.Base.FontColor = yt.Base.FontColor
		}
		if yt.Base.FontSize > 0 {
			t.Base.FontSize = yt.Base.FontSize
		}
	}
	if yt.Heading != nil {
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
		if yt.Heading.H1 != nil {
			if yt.Heading.H1.FontSize > 0 {
				t.Heading.H1FontSize = yt.Heading.H1.FontSize
			}
			if yt.Heading.H1.TextTransform != "" {
				t.Heading.H1TextTransform = yt.Heading.H1.TextTransform
			}
		}
		if yt.Heading.H2 != nil {
			if yt.Heading.H2.FontSize > 0 {
				t.Heading.H2FontSize = yt.Heading.H2.FontSize
			}
			if yt.Heading.H2.TextTransform != "" {
				t.Heading.H2TextTransform = yt.Heading.H2.TextTransform
			}
		}
		if yt.Heading.H3 != nil {
			if yt.Heading.H3.FontSize > 0 {
				t.Heading.H3FontSize = yt.Heading.H3.FontSize
			}
			if yt.Heading.H3.TextTransform != "" {
				t.Heading.H3TextTransform = yt.Heading.H3.TextTransform
			}
		}
		if yt.Heading.H4 != nil {
			if yt.Heading.H4.FontSize > 0 {
				t.Heading.H4FontSize = yt.Heading.H4.FontSize
			}
			if yt.Heading.H4.TextTransform != "" {
				t.Heading.H4TextTransform = yt.Heading.H4.TextTransform
			}
		}
	}
	if yt.Title != nil {
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
		}
		if yt.Title.Subtitle != nil {
			if yt.Title.Subtitle.FontSize > 0 {
				t.Title.SubtitleFontSize = yt.Title.Subtitle.FontSize
			}
			if yt.Title.Subtitle.FontColor != "" {
				t.Title.SubtitleFontColor = yt.Title.Subtitle.FontColor
			}
		}
	}
	if yt.Table != nil {
		if yt.Table.FontSize > 0 {
			t.Table.FontSize = yt.Table.FontSize
		}
		if yt.Table.BorderColor != "" {
			t.Table.BorderColor = yt.Table.BorderColor
		}
		if yt.Table.BorderWidth > 0 {
			t.Table.BorderWidth = yt.Table.BorderWidth
		}
		if yt.Table.Head != nil && yt.Table.Head.BackgroundColor != "" {
			t.Table.HeadBgColor = yt.Table.Head.BackgroundColor
		}
	}
	if yt.List != nil && yt.List.Indent > 0 {
		t.List.Indent = yt.List.Indent
	}
	if yt.Code != nil {
		if yt.Code.FontFamily != "" {
			t.Code.FontFamily = yt.Code.FontFamily
		}
		if yt.Code.FontSize > 0 {
			t.Code.FontSize = yt.Code.FontSize
		}
	}
	if yt.Link != nil {
		if yt.Link.FontColor != "" {
			t.Link.FontColor = yt.Link.FontColor
		}
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
