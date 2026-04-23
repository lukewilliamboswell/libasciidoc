# libasciidoc

[![CI Build Status](https://github.com/lukewilliamboswell/libasciidoc/workflows/ci-build/badge.svg)](https://github.com/lukewilliamboswell/libasciidoc/actions?query=workflow%3Aci-build)
[![Go Reference](https://pkg.go.dev/badge/github.com/lukewilliamboswell/libasciidoc.svg)](https://pkg.go.dev/github.com/lukewilliamboswell/libasciidoc)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukewilliamboswell/libasciidoc)](https://goreportcard.com/report/github.com/lukewilliamboswell/libasciidoc)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Go library and CLI for converting AsciiDoc to HTML5 and DOCX.

This is a fork of [bytesparadise/libasciidoc](https://github.com/bytesparadise/libasciidoc) with DOCX output support, static site generation, and modernised build tooling.

**[Read the full documentation](https://lukewilliamboswell.github.io/libasciidoc/)**

## Quick Start

Install the CLI tools:

```sh
go install github.com/lukewilliamboswell/libasciidoc/cmd/ascii2html@latest
go install github.com/lukewilliamboswell/libasciidoc/cmd/ascii2doc@latest
```

Convert a file to HTML:

```sh
ascii2html myfile.adoc
```

Convert a file to DOCX:

```sh
ascii2doc myfile.adoc
```

Build a static site:

```sh
ascii2html --static-site -o _site/ www/
```

## Using as a Library

```go
package main

import (
	"os"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"
)

func main() {
	input := `= My Document

Hello, AsciiDoc!`

	config := configuration.NewConfiguration()
	_, err := asciidoc.Convert(strings.NewReader(input), os.Stdout, config)
	if err != nil {
		panic(err)
	}
}
```

To convert a file:

```go
config := configuration.NewConfiguration(
	configuration.WithFilename("myfile.adoc"),
	configuration.WithHeaderFooter(true),
)
_, err := asciidoc.ConvertFile(os.Stdout, config)
```

To convert to DOCX:

```go
config := configuration.NewConfiguration(
	configuration.WithFilename("myfile.adoc"),
	configuration.WithBackEnd("docx"),
)
_, err := asciidoc.ConvertFile(output, config)
```

## Contributing

See the [contributing guide](https://lukewilliamboswell.github.io/libasciidoc/contributing/) for development setup, architecture overview, and how to submit changes.

## License

Apache License 2.0. See [LICENSE](LICENSE).
