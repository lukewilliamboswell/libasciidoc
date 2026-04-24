package testsupport

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/lukewilliamboswell/libasciidoc/configuration"
	"github.com/lukewilliamboswell/libasciidoc/internal/parser"
)

func PreparseDocument(source string, options ...interface{}) (string, error) {
	settings := []configuration.Setting{
		configuration.WithFilename("test.adoc"),
		configuration.WithAttribute("basebackend-html", true), // TODO: still needed?
	}
	opts := []parser.Option{}
	for _, o := range options {
		switch o := o.(type) {
		case configuration.Setting:
			settings = append(settings, o)
		case parser.Option:
			opts = append(opts, o)
		default:
			return "", fmt.Errorf("unexpected type of option: '%T'", o)
		}
	}
	result, err := parser.Preprocess(strings.NewReader(source), configuration.NewConfiguration(settings...), opts...)
	if log.IsLevelEnabled(log.DebugLevel) && err == nil {
		log.Debugf("preparsed document:\n%s", result)
	}
	return result, err

}
