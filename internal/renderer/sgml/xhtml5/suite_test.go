package xhtml5_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"
	"github.com/lukewilliamboswell/libasciidoc/types"

	"testing"

	_ "github.com/lukewilliamboswell/libasciidoc/internal/testsupport"
)

func RenderXHTML(actual string, settings ...configuration.Setting) (string, error) {
	output, _, err := RenderXHTMLWithMetadata(actual, settings...)
	return output, err
}

func RenderXHTMLWithMetadata(actual string, settings ...configuration.Setting) (string, types.Metadata, error) {
	allSettings := append([]configuration.Setting{configuration.WithFilename("test.adoc"), configuration.WithBackEnd("xhtml5")}, settings...)
	config := configuration.NewConfiguration(allSettings...)

	contentReader := strings.NewReader(actual)
	resultWriter := bytes.NewBuffer(nil)
	metadata, err := asciidoc.Convert(contentReader, resultWriter, config)
	if err != nil {
		return "", types.Metadata{}, err
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debug(resultWriter.String())
	}
	return resultWriter.String(), metadata, nil
}

func TestXHtml5(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "XHtml5 Suite")
}
