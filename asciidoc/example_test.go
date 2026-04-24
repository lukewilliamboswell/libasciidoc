package asciidoc_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"
)

func ExampleConvert() {
	input := `= Hello, AsciiDoc!

This is a simple paragraph.`

	config := configuration.NewConfiguration()
	_, err := asciidoc.Convert(strings.NewReader(input), os.Stdout, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	// Output:
	// <div class="paragraph">
	// <p>This is a simple paragraph.</p>
	// </div>
}

func ExampleConvert_list() {
	input := `* First item
* Second item
* Third item`

	config := configuration.NewConfiguration()
	_, err := asciidoc.Convert(strings.NewReader(input), os.Stdout, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	// Output:
	// <div class="ulist">
	// <ul>
	// <li>
	// <p>First item</p>
	// </li>
	// <li>
	// <p>Second item</p>
	// </li>
	// <li>
	// <p>Third item</p>
	// </li>
	// </ul>
	// </div>
}
