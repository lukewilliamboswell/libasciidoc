package testsupport

import (
	"bytes"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"
	"github.com/lukewilliamboswell/libasciidoc/types"
)

// RenderHTML renders the HTML body using the given source
func RenderHTML(actual string, settings ...configuration.Setting) (string, error) {
	output, _, err := RenderHTMLWithMetadata(actual, settings...)
	return output, err
}

// RenderHTML renders the HTML body using the given source
func RenderHTMLWithMetadata(actual string, settings ...configuration.Setting) (string, types.Metadata, error) {
	allSettings := append([]configuration.Setting{configuration.WithFilename("test.adoc"), configuration.WithBackEnd("html5")}, settings...)
	config := configuration.NewConfiguration(allSettings...)
	contentReader := strings.NewReader(actual)
	resultWriter := bytes.NewBuffer(nil)
	metadata, err := asciidoc.Convert(contentReader, resultWriter, config)
	if err != nil {
		log.Error(err)
		return "", types.Metadata{}, err
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debug(resultWriter.String())
	}
	return resultWriter.String(), metadata, nil
}

// RenderHTML renders the HTML body using the given source
func RenderHTMLFromFile(filename string, settings ...configuration.Setting) (string, types.Metadata, error) {
	info, err := os.Stat(filename)
	if err != nil {
		log.Error(err)
		return "", types.Metadata{}, err
	}

	allSettings := append([]configuration.Setting{
		configuration.WithLastUpdated(info.ModTime()),
		configuration.WithFilename(filename),
		configuration.WithBackEnd("html5")},
		settings...)
	config := configuration.NewConfiguration(allSettings...)
	f, err := os.Open(filename)
	if err != nil {
		log.Error(err)
		return "", types.Metadata{}, err
	}
	defer func() { f.Close() }()
	resultWriter := bytes.NewBuffer(nil)
	metadata, err := asciidoc.Convert(f, resultWriter, config)
	if err != nil {
		log.Error(err)
		return "", types.Metadata{}, err
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debug(resultWriter.String())
	}
	return resultWriter.String(), metadata, nil
}
