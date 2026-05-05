package docx_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/internal/renderer/docx"
)

var _ = Describe("custom role styling", func() {

	Context("YAML loader", func() {

		It("should ingest a single role with normalised colour", func() {
			theme, err := loadInlineTheme(`
role:
  foo:
    font_color: "#1A2B3C"
    font_size: 9
    font_style: italic
`)
			Expect(err).ToNot(HaveOccurred())
			role, ok := theme.Roles.Get("foo")
			Expect(ok).To(BeTrue())
			Expect(role.FontColor).To(Equal("1A2B3C"))
			Expect(role.FontSize).To(Equal(9.0))
			Expect(role.FontStyle).To(Equal("italic"))
			Expect(role.StyleID).To(Equal("RoleFoo"))
		})

		It("should round-trip multiple roles in source order", func() {
			theme, err := loadInlineTheme(`
role:
  alpha:
    font_color: "AAAAAA"
  beta:
    font_color: "BBBBBB"
  gamma:
    font_color: "CCCCCC"
`)
			Expect(err).ToNot(HaveOccurred())
			Expect(theme.Roles.Len()).To(Equal(3))
			var names []string
			theme.Roles.Each(func(r docx.RoleTheme) { names = append(names, r.Name) })
			Expect(names).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("should accept padding as a scalar, 2-element, or 4-element sequence", func() {
			theme, err := loadInlineTheme(`
role:
  one:
    padding: 4
  two:
    padding: [4, 6]
  four:
    padding: [1, 2, 3, 4]
`)
			Expect(err).ToNot(HaveOccurred())
			one, _ := theme.Roles.Get("one")
			Expect(one.PaddingTop).To(Equal(4.0))
			Expect(one.PaddingLeft).To(Equal(4.0))
			two, _ := theme.Roles.Get("two")
			Expect(two.PaddingTop).To(Equal(4.0))
			Expect(two.PaddingLeft).To(Equal(6.0))
			four, _ := theme.Roles.Get("four")
			Expect(four.PaddingTop).To(Equal(1.0))
			Expect(four.PaddingRight).To(Equal(2.0))
			Expect(four.PaddingBottom).To(Equal(3.0))
			Expect(four.PaddingLeft).To(Equal(4.0))
		})
	})

	Context("style emission", func() {

		It("should emit a paragraph style per role with the expected colour and font", func() {
			doc := renderDocxWithTheme(`Hello`, `
role:
  foo:
    font_color: "1A2B3C"
    font_size: 9
    font_style: italic
    background_color: "F4F4F2"
`)
			styles := string(doc.files["word/styles.xml"])
			Expect(styles).To(ContainSubstring(`w:styleId="RoleFoo"`))

			role := doc.findStyle("RoleFoo")
			Expect(role).ToNot(BeNil())
			Expect(role.Color).To(Equal("1A2B3C"))
			Expect(role.Italic).To(BeTrue())
			Expect(role.Size).To(Equal("18")) // 9pt = 18 half-points
			Expect(role.Shading).To(Equal("F4F4F2"))
		})

		It("should base role styles on Normal so unset properties inherit", func() {
			doc := renderDocxWithTheme(`Hello`, `
role:
  bar:
    font_color: "112233"
`)
			styles := string(doc.files["word/styles.xml"])
			// The role style block must declare basedOn="Normal".
			startIdx := strings.Index(styles, `w:styleId="RoleBar"`)
			Expect(startIdx).ToNot(Equal(-1))
			endIdx := strings.Index(styles[startIdx:], `</w:style>`)
			Expect(endIdx).ToNot(Equal(-1))
			roleBlock := styles[startIdx : startIdx+endIdx]
			Expect(roleBlock).To(ContainSubstring(`<w:basedOn w:val="Normal"/>`))
		})
	})

	Context("renderer mapping", func() {

		It("should attach the role's style id to a paragraph carrying the role", func() {
			doc := renderDocxWithTheme(`[.foo]
This is the body.
`, `
role:
  foo:
    font_color: "112233"
`)
			p := doc.findParagraph("This is the body.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("RoleFoo"))
		})

		It("should leave paragraphs unstyled when the role is not defined", func() {
			doc := renderDocxWithTheme(`[.unknown]
Plain prose here.
`, `
role:
  other:
    font_color: "112233"
`)
			p := doc.findParagraph("Plain prose here.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(BeEmpty())
		})

		It("should pick the first defined role when multiple roles are present", func() {
			doc := renderDocxWithTheme(`[.alpha.beta]
Two-role body.
`, `
role:
  alpha:
    font_color: "AAAAAA"
  beta:
    font_color: "BBBBBB"
`)
			p := doc.findParagraph("Two-role body.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("RoleAlpha"))
		})

		It("should skip undefined roles when finding the first defined one", func() {
			doc := renderDocxWithTheme(`[.missing.beta]
Mixed-role body.
`, `
role:
  beta:
    font_color: "BBBBBB"
`)
			p := doc.findParagraph("Mixed-role body.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("RoleBeta"))
		})

		It("should look up role names case-insensitively", func() {
			doc := renderDocxWithTheme(`[.Foo]
Capital role.
`, `
role:
  foo:
    font_color: "112233"
`)
			p := doc.findParagraph("Capital role.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("RoleFoo"))
		})
	})

	Context("end-to-end placeholder example", func() {

		It("should render the placeholder template guidance with its theme style", func() {
			doc := renderDocxWithTheme(`[.placeholder]
Edit this part of the SOW with the actual scope.
`, `
role:
  placeholder:
    font_color: "6B6B6B"
    font_style: italic
    background_color: "F4F4F2"
    padding: [4, 6]
`)
			p := doc.findParagraph("Edit this part of the SOW")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("RolePlaceholder"))

			style := doc.findStyle("RolePlaceholder")
			Expect(style).ToNot(BeNil())
			Expect(style.Color).To(Equal("6B6B6B"))
			Expect(style.Italic).To(BeTrue())
			Expect(style.Shading).To(Equal("F4F4F2"))
			Expect(style.IndentLeft).To(Equal("120")) // 6pt = 120 twips
		})
	})
})

// loadInlineTheme writes the YAML to a temp file and returns the loaded
// DocxTheme so loader behaviour can be asserted directly without going via
// a full DOCX render.
func loadInlineTheme(themeYAML string) (*docx.DocxTheme, error) {
	tmpDir := GinkgoT().TempDir()
	themePath := tmpDir + "/test-theme.yml"
	if err := os.WriteFile(themePath, []byte(strings.TrimSpace(themeYAML)), 0644); err != nil {
		return nil, err
	}
	return docx.LoadTheme(themePath)
}
