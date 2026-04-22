package testsupport

import (
	"bytes"
	"strings"

	"github.com/bytesparadise/libasciidoc"
	"github.com/bytesparadise/libasciidoc/pkg/configuration"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	log "github.com/sirupsen/logrus"
)

// RenderDOCX renders the DOCX output using the given source
func RenderDOCX(actual string, settings ...configuration.Setting) ([]byte, error) {
	_, result, err := RenderDOCXWithMetadata(actual, settings...)
	return result, err
}

// RenderDOCXWithMetadata renders the DOCX output using the given source and returns metadata
func RenderDOCXWithMetadata(actual string, settings ...configuration.Setting) (types.Metadata, []byte, error) {
	allSettings := append([]configuration.Setting{
		configuration.WithFilename("test.adoc"),
		configuration.WithBackEnd("docx"),
	}, settings...)
	config := configuration.NewConfiguration(allSettings...)
	contentReader := strings.NewReader(actual)
	resultWriter := bytes.NewBuffer(nil)
	metadata, err := libasciidoc.Convert(contentReader, resultWriter, config)
	if err != nil {
		log.Error(err)
		return types.Metadata{}, nil, err
	}
	return metadata, resultWriter.Bytes(), nil
}
