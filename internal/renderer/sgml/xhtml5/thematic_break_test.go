package xhtml5_test

import (
	. "github.com/lukewilliamboswell/libasciidoc/internal/testsupport"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("thematic breaks", func() {

	It("simple", func() {
		source := `before

'''

after`
		expected := `<div class="paragraph">
<p>before</p>
</div>
<hr/>
<div class="paragraph">
<p>after</p>
</div>
`
		Expect(RenderXHTML(source)).To(MatchHTML(expected))
	})
})
