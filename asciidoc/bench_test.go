package asciidoc_test

import (
	"strings"
	"testing"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"

	log "github.com/sirupsen/logrus"
)

func BenchmarkRealDocumentProcessing(b *testing.B) {
	log.SetLevel(log.ErrorLevel)
	b.Run("demo.adoc", processDocument("../test/compat/demo.adoc"))
	b.Run("vertx-examples.adoc", processDocument("../test/bench/vertx-examples.adoc"))
	b.Run("mocking.adoc", processDocument("../test/bench/mocking.adoc"))
}

func processDocument(filename string) func(b *testing.B) {
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out := &strings.Builder{}
			_, err := asciidoc.ConvertFile(out,
				configuration.NewConfiguration(
					configuration.WithFilename(filename),
					configuration.WithCSS([]string{"path/to/style.css"}),
					configuration.WithHeaderFooter(true)))
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
