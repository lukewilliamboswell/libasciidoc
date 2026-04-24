package docx_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("special characters and symbols", func() {

	Context("symbols", func() {

		It("should render copyright symbol", func() {
			doc := renderDocx(`(C) 2024 Acme Corp`)

			Expect(doc.text()).To(ContainSubstring("\u00a9"))
		})

		It("should render trademark symbol", func() {
			doc := renderDocx(`Acme(TM) Product`)

			Expect(doc.text()).To(ContainSubstring("\u2122"))
		})

		It("should render registered symbol", func() {
			doc := renderDocx(`Acme(R) Brand`)

			Expect(doc.text()).To(ContainSubstring("\u00ae"))
		})

		It("should render em dash", func() {
			doc := renderDocx(`one -- two`)

			Expect(doc.text()).To(ContainSubstring("\u2014"))
		})

		It("should render ellipsis", func() {
			doc := renderDocx(`and so on...`)

			Expect(doc.text()).To(ContainSubstring("\u2026"))
		})

		It("should render arrows", func() {
			doc := renderDocx(`left <- right -> forward => back`)

			text := doc.text()
			Expect(text).To(ContainSubstring("\u2190"))
			Expect(text).To(ContainSubstring("\u2192"))
			Expect(text).To(ContainSubstring("\u21d2"))
		})

		It("should render curly apostrophe", func() {
			doc := renderDocx(`That's correct.`)

			Expect(doc.text()).To(ContainSubstring("\u2019"))
		})

		It("should render curly double quotes", func() {
			doc := renderDocx("She said, \"`hello`\".")

			text := doc.text()
			Expect(text).To(ContainSubstring("\u201c"))
			Expect(text).To(ContainSubstring("\u201d"))
		})
	})

	Context("special characters", func() {

		It("should escape ampersand in XML but render as text", func() {
			doc := renderDocx(`Tom & Jerry`)

			r := doc.findRun("&")
			Expect(r).ToNot(BeNil())
			// Verify the raw XML has the escaped form
			Expect(doc.documentXML()).To(ContainSubstring("&amp;"))
		})

		It("should escape angle brackets in XML", func() {
			doc := renderDocx(`a < b > c`)

			Expect(doc.documentXML()).To(ContainSubstring("&lt;"))
			Expect(doc.documentXML()).To(ContainSubstring("&gt;"))
		})
	})

	Context("predefined attributes", func() {

		It("should render non-breaking space as Unicode U+00A0 in the raw XML", func() {
			doc := renderDocx(`before{nbsp}after`)

			// doc.text() normalizes whitespace, so check raw XML for the actual char
			Expect(doc.documentXML()).To(ContainSubstring("\u00a0"))
		})

		It("should render C++ shorthand", func() {
			doc := renderDocx(`Written in {cpp}.`)

			r := doc.findRun("C++")
			Expect(r).ToNot(BeNil())
		})

		It("should render degree symbol", func() {
			doc := renderDocx(`The temperature is 100{deg}F.`)

			r := doc.findRun("\u00b0")
			Expect(r).ToNot(BeNil())
		})
	})

	Context("run merging via paragraphBuilder", func() {

		It("should merge adjacent plain text runs into a single w:r", func() {
			// A possessive apostrophe is parsed as a Symbol element interspersed
			// between StringElement runs.  Without deferred run serialisation,
			// "Luke" + "\u2019" + "s guide" would produce three separate <w:r>
			// elements.  paragraphBuilder merges adjacent runs sharing the same
			// runStyle, so the paragraph should contain only ONE non-empty run.
			doc := renderDocx(`Luke's guide`)

			paras := doc.parseParagraphs()
			var targetPara *parsedParagraph
			for i := range paras {
				if strings.Contains(paras[i].text(), "Luke") {
					targetPara = &paras[i]
					break
				}
			}
			Expect(targetPara).ToNot(BeNil())

			nonEmpty := 0
			for _, r := range targetPara.Runs {
				if r.Text != "" {
					nonEmpty++
				}
			}
			Expect(nonEmpty).To(Equal(1), "adjacent plain runs should be merged into a single <w:r>")
		})
	})
})
