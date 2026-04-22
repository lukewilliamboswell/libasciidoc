package types_test

import (
	"github.com/lukewilliamboswell/libasciidoc/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("convert to inline elements", func() {

	It("inline content without trailing spaces", func() {
		source := []interface{}{
			&types.StringElement{Content: "hello"},
			&types.StringElement{Content: "world"},
		}
		expected := []interface{}{
			&types.StringElement{Content: "helloworld"},
		}
		Expect(types.NewInlineElements(source...)).To(Equal(expected))
	})

	It("inline content with trailing spaces", func() {
		source := []interface{}{
			&types.StringElement{Content: "hello, "},
			&types.StringElement{Content: "world   "},
		}
		expected := []interface{}{
			&types.StringElement{Content: "hello, world   "},
		}
		Expect(types.NewInlineElements(source...)).To(Equal(expected))
	})
})

var _ = DescribeTable("TrimTrailingSpaces",

	func(source, expected []interface{}) {
		Expect(types.TrimTrailingSpaces(source)).To(Equal(expected))
	},
	Entry("empty slice",
		[]interface{}{},
		[]interface{}{}),

	Entry("single element with trailing spaces",
		[]interface{}{
			&types.StringElement{
				Content: "pasta ", // trailing spaces
			},
		},
		[]interface{}{
			&types.StringElement{
				Content: "pasta", // timmed
			},
		}),

	Entry("multiple elements with trailing spaces - case 1",
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{},
			&types.StringElement{
				Content: "pasta ", // trailing spaces
			},
		},
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{},
			&types.StringElement{
				Content: "pasta", // timmed
			},
		}),
	Entry("multiple elements with trailing spaces - case 2",
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{},
			&types.StringElement{
				Content: " ", // isolated trailing space
			},
		},
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{},
		}),

	Entry("multiple elements without trailing spaces",
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{},
			&types.StringElement{
				Content: "pasta", // no trailing spaces
			},
		},
		[]interface{}{&types.StringElement{
			Content: "cookies",
		},
			&types.InlineLink{},
			&types.StringElement{
				Content: "pasta", // no change
			},
		}),

	Entry("noop",
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{}, // not a StringElement
		},
		[]interface{}{
			&types.StringElement{
				Content: "cookies",
			},
			&types.InlineLink{},
		}),
)

var _ = DescribeTable("split elements per line",
	func(elements []interface{}, expected [][]interface{}) {
		result := types.SplitElementsPerLine(elements)
		Expect(result).To(Equal(expected))

	},
	Entry("empty elements",
		[]interface{}{},
		[][]interface{}{}),

	Entry("single line",
		[]interface{}{
			&types.StringElement{
				Content: "cookie",
			},
			&types.Callout{
				Ref: 1,
			},
		},
		[][]interface{}{
			{
				&types.StringElement{
					Content: "cookie",
				},
				&types.Callout{
					Ref: 1,
				},
			},
		}),

	Entry("2 lines without callouts",
		[]interface{}{
			&types.StringElement{
				Content: "cookie",
			},
			&types.Callout{
				Ref: 1,
			},
			&types.StringElement{
				Content: "\npasta",
			},
			&types.Callout{
				Ref: 2,
			},
		},
		[][]interface{}{
			{
				&types.StringElement{
					Content: "cookie",
				},
				&types.Callout{
					Ref: 1,
				},
			},
			{
				&types.StringElement{
					Content: "pasta",
				},
				&types.Callout{
					Ref: 2,
				},
			},
		}),

	Entry("3 lines without callouts",
		[]interface{}{
			&types.StringElement{
				Content: "cookie\npasta\nchocolate",
			},
		},
		[][]interface{}{
			{
				&types.StringElement{
					Content: "cookie",
				},
			},
			{
				&types.StringElement{
					Content: "pasta",
				},
			},
			{
				&types.StringElement{
					Content: "chocolate",
				},
			},
		}),

	Entry("3 lines without callouts",
		[]interface{}{
			&types.StringElement{
				Content: "cookie",
			},
			&types.Callout{
				Ref: 1,
			},
			&types.StringElement{
				Content: "here\npasta",
			},
			&types.Callout{
				Ref: 2,
			},
			&types.StringElement{
				Content: "also\nchocolate",
			},
			&types.Callout{
				Ref: 3,
			},
		},
		[][]interface{}{
			{
				&types.StringElement{
					Content: "cookie",
				},
				&types.Callout{
					Ref: 1,
				},
				&types.StringElement{
					Content: "here",
				},
			},
			{
				&types.StringElement{
					Content: "pasta",
				},
				&types.Callout{
					Ref: 2,
				},
				&types.StringElement{
					Content: "also",
				},
			},
			{
				&types.StringElement{
					Content: "chocolate",
				},
				&types.Callout{
					Ref: 3,
				},
			},
		}),
)

var _ = DescribeTable("insert element in slice",

	func(elements []interface{}, element interface{}, index int, expected []interface{}) {
		result := types.InsertAt(elements, element, index)
		Expect(result).To(Equal(expected))
	},

	Entry("empty elements",
		[]interface{}{},
		&types.TableOfContents{},
		0,
		[]interface{}{
			&types.TableOfContents{},
		}),

	Entry("insert after preamble elements",
		[]interface{}{
			&types.Section{},
			&types.Preamble{},
			&types.Section{},
		},
		&types.TableOfContents{},
		2,
		[]interface{}{
			&types.Section{},
			&types.Preamble{},
			&types.TableOfContents{},
			&types.Section{},
		}),
)

var _ = Describe("merge (via NewInlineElements)", func() {

	It("should merge strings", func() {
		result, err := types.NewInlineElements("hello", " ", "world")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal(&types.StringElement{Content: "hello world"}))
	})

	It("should merge bytes into string", func() {
		result, err := types.NewInlineElements([]byte("he"), []byte("llo"))
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal(&types.StringElement{Content: "hello"}))
	})

	It("should flush buffer on non-string element", func() {
		link := &types.InlineLink{}
		result, err := types.NewInlineElements("before ", link, " after")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(3))
		Expect(result[0]).To(Equal(&types.StringElement{Content: "before "}))
		Expect(result[1]).To(Equal(link))
		Expect(result[2]).To(Equal(&types.StringElement{Content: " after"}))
	})

	It("should handle nested slices", func() {
		result, err := types.NewInlineElements(
			[]interface{}{"nested ", "content"},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal(&types.StringElement{Content: "nested content"}))
	})

	It("should skip nil elements", func() {
		result, err := types.NewInlineElements("a", nil, "b")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal(&types.StringElement{Content: "ab"}))
	})

	It("should trim trailing space before em-dash symbol", func() {
		sym := &types.Symbol{Name: " -- "}
		result, err := types.NewInlineElements("hello ", sym, " world")
		Expect(err).NotTo(HaveOccurred())
		// "hello " should be trimmed to "hello" before the symbol
		Expect(result).To(HaveLen(3))
		Expect(result[0]).To(Equal(&types.StringElement{Content: "hello"}))
		Expect(result[1]).To(Equal(sym))
		Expect(result[2]).To(Equal(&types.StringElement{Content: " world"}))
	})
})

var _ = Describe("Reduce", func() {

	It("should return single string from one StringElement", func() {
		result := types.Reduce([]interface{}{
			&types.StringElement{Content: "hello"},
		})
		Expect(result).To(Equal("hello"))
	})

	It("should return nil for empty slice", func() {
		result := types.Reduce([]interface{}{})
		Expect(result).To(BeNil())
	})

	It("should return nil for nil elements", func() {
		result := types.Reduce([]interface{}{nil, nil})
		Expect(result).To(BeNil())
	})

	It("should return slice when multiple non-string elements remain", func() {
		link := &types.InlineLink{}
		result := types.Reduce([]interface{}{
			&types.StringElement{Content: "text"},
			link,
		})
		slice, ok := result.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(slice).To(HaveLen(2))
	})

	It("should apply options to string result", func() {
		result := types.Reduce(
			[]interface{}{&types.StringElement{Content: "  hello  "}},
			func(s string) string {
				return "trimmed"
			},
		)
		Expect(result).To(Equal("trimmed"))
	})

	It("should handle StringElement input directly", func() {
		result := types.Reduce(&types.StringElement{Content: "hello"})
		Expect(result).To(Equal("hello"))
	})

	It("should handle string input directly", func() {
		result := types.Reduce("hello")
		Expect(result).To(Equal("hello"))
	})

	It("should return nil for empty string", func() {
		result := types.Reduce("")
		Expect(result).To(BeNil())
	})

	It("should return non-slice non-string input unchanged", func() {
		link := &types.InlineLink{}
		result := types.Reduce(link)
		Expect(result).To(Equal(link))
	})
})
