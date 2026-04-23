package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("delimited blocks", func() {

	Context("source and listing blocks", func() {

		It("should render a source block with CodeBlock style", func() {
			doc := renderDocx(`[source,go]
----
func main() {
    fmt.Println("hello")
}
----`)

			p := doc.findParagraph("func main()")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("CodeBlock"))
		})

		It("should render a listing block with CodeBlock style", func() {
			doc := renderDocx(`----
some listing content
with multiple lines
----`)

			p := doc.findParagraph("some listing content")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("CodeBlock"))

			p2 := doc.findParagraph("with multiple lines")
			Expect(p2).ToNot(BeNil())
			Expect(p2.Style).To(Equal("CodeBlock"))
		})

		It("should render a fenced code block with CodeBlock style", func() {
			doc := renderDocx("```\nfenced code\n```")

			p := doc.findParagraph("fenced code")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("CodeBlock"))
		})

		It("should render a source block title as Caption style", func() {
			doc := renderDocx(`.My Code
[source,go]
----
x := 1
----`)

			caption := doc.findParagraph("My Code")
			Expect(caption).ToNot(BeNil())
			Expect(caption.Style).To(Equal("Caption"))

			code := doc.findParagraph("x := 1")
			Expect(code).ToNot(BeNil())
			Expect(code.Style).To(Equal("CodeBlock"))
		})
	})

	Context("literal blocks", func() {

		It("should render a literal block with CodeBlock style", func() {
			doc := renderDocx(`....
literal content
    indented
....`)

			p := doc.findParagraph("literal content")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("CodeBlock"))
		})
	})

	Context("quote blocks", func() {

		It("should render a quote block with Quote style on each paragraph", func() {
			doc := renderDocx(`[quote]
____
This is a quoted paragraph.
____`)

			p := doc.findParagraph("This is a quoted paragraph.")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Quote"))
		})

		It("should render a quote block with attribution as a Caption paragraph", func() {
			doc := renderDocx(`[quote,Albert Einstein,On Relativity]
____
Imagination is more important than knowledge.
____`)

			quotePara := doc.findParagraph("Imagination is more important")
			Expect(quotePara).ToNot(BeNil())
			Expect(quotePara.Style).To(Equal("Quote"))

			attrPara := doc.findParagraph("Albert Einstein")
			Expect(attrPara).ToNot(BeNil())
			Expect(attrPara.Style).To(Equal("Caption"))
			Expect(attrPara.text()).To(ContainSubstring("On Relativity"))
		})

		It("should render a quote with author only, attribution without title portion", func() {
			doc := renderDocx(`[quote,Mark Twain]
____
The reports of my death are greatly exaggerated.
____`)

			attrPara := doc.findParagraph("Mark Twain")
			Expect(attrPara).ToNot(BeNil())
			Expect(attrPara.Style).To(Equal("Caption"))
			// Attribution should be "— Mark Twain" with no trailing comma/title
			Expect(attrPara.text()).To(HavePrefix("\u2014 Mark Twain"))
			Expect(attrPara.text()).ToNot(ContainSubstring(", "))
		})

		It("should render a markdown-style quote block with Quote style", func() {
			doc := renderDocx(`> This is a markdown quote.
> Second line.`)

			p := doc.findParagraph("markdown quote")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Quote"))
		})
	})

	Context("verse blocks", func() {

		It("should render a verse block with Quote style and attribution", func() {
			doc := renderDocx(`[verse,Carl Sandburg,Fog]
____
The fog comes
on little cat feet.
____`)

			p := doc.findParagraph("fog comes")
			Expect(p).ToNot(BeNil())
			Expect(p.Style).To(Equal("Quote"))

			attrPara := doc.findParagraph("Carl Sandburg")
			Expect(attrPara).ToNot(BeNil())
			Expect(attrPara.text()).To(ContainSubstring("Fog"))
		})
	})

	Context("example blocks", func() {

		It("should render an example block with its content", func() {
			doc := renderDocx(`====
This is an example block.
====`)

			Expect(doc.text()).To(ContainSubstring("This is an example block."))
			// Plain example blocks should NOT have Admonition style
			for _, p := range doc.parseParagraphs() {
				Expect(p.Style).ToNot(Equal("Admonition"))
			}
		})
	})

	Context("sidebar blocks", func() {

		It("should render a sidebar block with its content", func() {
			doc := renderDocx(`****
This is a sidebar.
****`)

			Expect(doc.text()).To(ContainSubstring("This is a sidebar."))
		})
	})

	Context("comment blocks", func() {

		It("should not render comment block content", func() {
			doc := renderDocx(`before

////
This is a comment and should not appear.
////

after`)

			Expect(doc.text()).To(ContainSubstring("before"))
			Expect(doc.text()).To(ContainSubstring("after"))
			Expect(doc.text()).ToNot(ContainSubstring("comment and should not appear"))
		})
	})
})
