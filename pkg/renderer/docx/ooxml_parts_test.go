package docx_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
)

var _ = Describe("OOXML optional parts", func() {

	Context("word/settings.xml", func() {

		It("should be valid XML with compatibility mode 15", func() {
			doc := renderDocx("hello")
			content := doc.settingsXML()
			Expect(content).ToNot(BeEmpty())

			// Must be well-formed XML
			Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed())

			Expect(content).To(ContainSubstring(`<w:defaultTabStop w:val="720"`))
			Expect(content).To(ContainSubstring(`w:name="compatibilityMode"`))
			Expect(content).To(ContainSubstring(`w:val="15"`))
		})
	})

	Context("word/fontTable.xml", func() {

		It("should be valid XML containing default fonts", func() {
			doc := renderDocx("hello")
			content := doc.fontTableXML()
			Expect(content).ToNot(BeEmpty())

			Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed())

			// Default theme fonts
			Expect(content).To(ContainSubstring(`w:name="Aptos"`))
			Expect(content).To(ContainSubstring(`w:name="Courier New"`))
			Expect(content).To(ContainSubstring(`w:name="Symbol"`))
		})

		It("should include heading font when different from base", func() {
			tmpDir := GinkgoT().TempDir()
			themePath := filepath.Join(tmpDir, "theme.yml")
			Expect(os.WriteFile(themePath, []byte("heading:\n  font_family: Georgia\n"), 0644)).To(Succeed())

			doc := renderDocx("hello", configuration.WithThemePath(themePath))
			content := doc.fontTableXML()
			Expect(content).To(ContainSubstring(`w:name="Georgia"`))
		})
	})

	Context("docProps/core.xml", func() {

		It("should be valid XML with metadata when header is present", func() {
			fixedTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
			doc := renderDocx("= My Document\nJohn Doe\n\nContent here.",
				configuration.WithLastUpdated(fixedTime))

			content := doc.corePropertiesXML()
			Expect(content).ToNot(BeEmpty())

			Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed())

			Expect(content).To(ContainSubstring(`<dc:title>My Document</dc:title>`))
			Expect(content).To(ContainSubstring(`<dc:creator>John Doe</dc:creator>`))
			Expect(content).To(ContainSubstring(`<dcterms:created xsi:type="dcterms:W3CDTF">2024-06-15T10:30:00Z</dcterms:created>`))
			Expect(content).To(ContainSubstring(`<dcterms:modified xsi:type="dcterms:W3CDTF">2024-06-15T10:30:00Z</dcterms:modified>`))
		})

		It("should escape XML special characters in metadata", func() {
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			doc := renderDocx("= Title with <angle> & \"quotes\"\nA & B\n\nContent.",
				configuration.WithLastUpdated(fixedTime))

			content := doc.corePropertiesXML()
			Expect(content).To(ContainSubstring(`&lt;angle&gt;`))
			Expect(content).To(ContainSubstring(`&amp;`))
		})

		It("should omit title and creator when not present", func() {
			doc := renderDocx("Just a paragraph.")
			content := doc.corePropertiesXML()

			Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed())

			Expect(content).ToNot(ContainSubstring(`<dc:title>`))
			Expect(content).ToNot(ContainSubstring(`<dc:creator>`))
		})

		It("should omit timestamps when LastUpdated is zero", func() {
			doc := renderDocx("Just a paragraph.")
			content := doc.corePropertiesXML()

			Expect(content).ToNot(ContainSubstring(`<dcterms:created`))
			Expect(content).ToNot(ContainSubstring(`<dcterms:modified`))
		})

		It("should include title but omit creator when no author", func() {
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			doc := renderDocx("= Solo Title\n\nContent.",
				configuration.WithLastUpdated(fixedTime))

			content := doc.corePropertiesXML()
			Expect(content).To(ContainSubstring(`<dc:title>Solo Title</dc:title>`))
			Expect(content).ToNot(ContainSubstring(`<dc:creator>`))
		})
	})

	Context("docProps/app.xml", func() {

		It("should be valid XML with application name", func() {
			doc := renderDocx("hello")
			content := doc.appPropertiesXML()
			Expect(content).ToNot(BeEmpty())

			Expect(xml.Unmarshal([]byte(content), new(interface{}))).To(Succeed())

			Expect(content).To(ContainSubstring(`<Application>libasciidoc</Application>`))
		})
	})

	Context("content types", func() {

		It("should include overrides for all new parts", func() {
			doc := renderDocx("hello")
			ct := doc.contentTypesXML()

			Expect(ct).To(ContainSubstring(`/word/settings.xml`))
			Expect(ct).To(ContainSubstring(`wordprocessingml.settings+xml`))
			Expect(ct).To(ContainSubstring(`/word/fontTable.xml`))
			Expect(ct).To(ContainSubstring(`wordprocessingml.fontTable+xml`))
			Expect(ct).To(ContainSubstring(`/docProps/core.xml`))
			Expect(ct).To(ContainSubstring(`core-properties+xml`))
			Expect(ct).To(ContainSubstring(`/docProps/app.xml`))
			Expect(ct).To(ContainSubstring(`extended-properties+xml`))
		})
	})

	Context("relationships", func() {

		It("should include package-level relationships for core and app properties", func() {
			doc := renderDocx("hello")
			rels := doc.packageRelsXML()

			Expect(xml.Unmarshal([]byte(rels), new(interface{}))).To(Succeed())

			Expect(rels).To(ContainSubstring(`docProps/core.xml`))
			Expect(rels).To(ContainSubstring(`core-properties`))
			Expect(rels).To(ContainSubstring(`docProps/app.xml`))
			Expect(rels).To(ContainSubstring(`extended-properties`))
		})

		It("should include document-level relationships for settings and fontTable", func() {
			doc := renderDocx("hello")
			rels := doc.relationshipsXML()

			Expect(xml.Unmarshal([]byte(rels), new(interface{}))).To(Succeed())

			Expect(rels).To(ContainSubstring(`settings.xml`))
			Expect(rels).To(ContainSubstring(`fontTable.xml`))
		})
	})

	Context("thematic break", func() {

		It("should render as a paragraph border, not Unicode dashes", func() {
			doc := renderDocx("before\n\n'''\n\nafter")
			docXML := doc.documentXML()

			Expect(docXML).To(ContainSubstring(`<w:pBdr>`))
			Expect(docXML).To(ContainSubstring(`<w:bottom w:val="single"`))
			Expect(docXML).ToNot(ContainSubstring("\u2500"))
		})
	})

	Context("page break", func() {

		It("should render <<< as a DOCX page break", func() {
			doc := renderDocx("before\n\n<<<\n\nafter")
			docXML := doc.documentXML()

			Expect(docXML).To(ContainSubstring(`<w:br w:type="page"/>`))
			// Both surrounding paragraphs should still be present
			Expect(doc.text()).To(ContainSubstring("before"))
			Expect(doc.text()).To(ContainSubstring("after"))
		})

		It("should not treat <<< without blank line after as page break", func() {
			doc := renderDocx("before\n\n<<<\nafter")
			docXML := doc.documentXML()

			// Without the required blank line, <<< should be treated as paragraph text
			Expect(strings.Count(docXML, `<w:br w:type="page"/>`)).To(Equal(0))
		})
	})
})
