package renderer_test

import (
	"bytes"

	"github.com/lukewilliamboswell/libasciidoc/pkg/configuration"
	"github.com/lukewilliamboswell/libasciidoc/pkg/renderer"
	"github.com/lukewilliamboswell/libasciidoc/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("renderer", func() {

	emptyDoc := func() *types.Document {
		return &types.Document{
			Elements:          []interface{}{},
			ElementReferences: types.ElementReferences{},
			Footnotes:         []*types.Footnote{},
		}
	}

	DescribeTable("should render with valid backends",
		func(backend string) {
			config := configuration.NewConfiguration(configuration.WithBackEnd(backend))
			output := &bytes.Buffer{}
			_, err := renderer.Render(emptyDoc(), config, output)
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("html", "html"),
		Entry("html5", "html5"),
		Entry("xhtml", "xhtml"),
		Entry("xhtml5", "xhtml5"),
		Entry("docx", "docx"),
	)

	It("should use html5 backend by default", func() {
		config := configuration.NewConfiguration()
		output := &bytes.Buffer{}
		_, err := renderer.Render(emptyDoc(), config, output)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return error for unsupported backend", func() {
		config := configuration.NewConfiguration(configuration.WithBackEnd("pdf"))
		output := &bytes.Buffer{}
		_, err := renderer.Render(emptyDoc(), config, output)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("backend 'pdf' not supported"))
	})
})
