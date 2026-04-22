package configuration_test

import (
	"io"
	"time"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// mockMacroTemplate implements configuration.MacroTemplate for testing
type mockMacroTemplate struct{}

func (m *mockMacroTemplate) Execute(wr io.Writer, data interface{}) error {
	return nil
}

var _ = Describe("configuration", func() {

	Describe("NewConfiguration", func() {

		It("should have correct defaults", func() {
			config := configuration.NewConfiguration()
			Expect(config.BackEnd).To(Equal("html5"))
			Expect(config.Attributes["basebackend-html"]).To(BeTrue())
			Expect(config.Attributes["backend"]).To(Equal("html5"))
			Expect(config.Macros).NotTo(BeNil())
			Expect(config.Macros).To(BeEmpty())
			Expect(config.Filename).To(BeEmpty())
			Expect(config.WrapInHTMLBodyElement).To(BeFalse())
			Expect(config.CSS).To(BeNil())
		})
	})

	Describe("WithBackEnd", func() {

		DescribeTable("should set backend and basebackend-html attribute",
			func(backend string, expectBaseBackendHTML bool) {
				config := configuration.NewConfiguration(configuration.WithBackEnd(backend))
				Expect(config.BackEnd).To(Equal(backend))
				Expect(config.Attributes["backend"]).To(Equal(backend))
				if expectBaseBackendHTML {
					Expect(config.Attributes["basebackend-html"]).To(BeTrue())
				} else {
					Expect(config.Attributes).NotTo(HaveKey("basebackend-html"))
				}
			},
			Entry("html", "html", true),
			Entry("html5", "html5", true),
			Entry("xhtml", "xhtml", true),
			Entry("xhtml5", "xhtml5", true),
			Entry("docx", "docx", false),
			Entry("empty string", "", false),
		)
	})

	Describe("WithAttribute", func() {

		It("should set a single attribute", func() {
			config := configuration.NewConfiguration(
				configuration.WithAttribute("foo", "bar"),
			)
			Expect(config.Attributes["foo"]).To(Equal("bar"))
		})
	})

	Describe("WithAttributes", func() {

		It("should replace the attribute map", func() {
			attrs := map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
			}
			config := configuration.NewConfiguration(
				configuration.WithAttributes(attrs),
			)
			Expect(config.Attributes).To(HaveKeyWithValue("key1", "val1"))
			Expect(config.Attributes).To(HaveKeyWithValue("key2", "val2"))
			// default attributes should be replaced
			Expect(config.Attributes).NotTo(HaveKey("basebackend-html"))
		})
	})

	Describe("WithFilename", func() {

		It("should set the filename", func() {
			config := configuration.NewConfiguration(
				configuration.WithFilename("test.adoc"),
			)
			Expect(config.Filename).To(Equal("test.adoc"))
		})
	})

	Describe("WithCSS", func() {

		It("should set CSS hrefs", func() {
			hrefs := []string{"style.css", "custom.css"}
			config := configuration.NewConfiguration(
				configuration.WithCSS(hrefs),
			)
			Expect(config.CSS).To(Equal(hrefs))
		})
	})

	Describe("WithHeaderFooter", func() {

		It("should set WrapInHTMLBodyElement to true", func() {
			config := configuration.NewConfiguration(
				configuration.WithHeaderFooter(true),
			)
			Expect(config.WrapInHTMLBodyElement).To(BeTrue())
		})

		It("should set WrapInHTMLBodyElement to false", func() {
			config := configuration.NewConfiguration(
				configuration.WithHeaderFooter(false),
			)
			Expect(config.WrapInHTMLBodyElement).To(BeFalse())
		})
	})

	Describe("WithLastUpdated", func() {

		It("should set the last updated time", func() {
			now := time.Now()
			config := configuration.NewConfiguration(
				configuration.WithLastUpdated(now),
			)
			Expect(config.LastUpdated).To(Equal(now))
		})
	})

	Describe("WithFigureCaption", func() {

		It("should set the figure-caption attribute", func() {
			config := configuration.NewConfiguration(
				configuration.WithFigureCaption("Figure"),
			)
			Expect(config.Attributes["figure-caption"]).To(Equal("Figure"))
		})
	})

	Describe("WithMacroTemplate", func() {

		It("should register a macro template", func() {
			tmpl := &mockMacroTemplate{}
			config := configuration.NewConfiguration(
				configuration.WithMacroTemplate("mymacro", tmpl),
			)
			Expect(config.Macros).To(HaveKey("mymacro"))
			Expect(config.Macros["mymacro"]).To(Equal(tmpl))
		})
	})

	Describe("composed settings", func() {

		It("should apply multiple settings correctly", func() {
			now := time.Now()
			config := configuration.NewConfiguration(
				configuration.WithBackEnd("xhtml5"),
				configuration.WithFilename("doc.adoc"),
				configuration.WithAttribute("author", "Test"),
				configuration.WithLastUpdated(now),
				configuration.WithHeaderFooter(true),
				configuration.WithCSS([]string{"main.css"}),
			)
			Expect(config.BackEnd).To(Equal("xhtml5"))
			Expect(config.Filename).To(Equal("doc.adoc"))
			Expect(config.Attributes["author"]).To(Equal("Test"))
			Expect(config.LastUpdated).To(Equal(now))
			Expect(config.WrapInHTMLBodyElement).To(BeTrue())
			Expect(config.CSS).To(Equal([]string{"main.css"}))
			// basebackend-html should still be set for xhtml5
			Expect(config.Attributes["basebackend-html"]).To(BeTrue())
		})
	})
})
