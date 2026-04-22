package docx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bytesparadise/libasciidoc/testsupport"
)

var _ = Describe("integration", func() {

	It("should render a business document with correct structure", func() {
		source := `= Service Agreement
Acme Corp <legal@acme.com>
v1.0, 2024-01-15

== Overview

This *Service Agreement* (the "Agreement") is entered into by and between
_Acme Corp_ ("Provider") and the client ("Client").

== Scope of Services

The Provider agrees to deliver the following services:

. Initial consultation and requirements gathering
. System design and architecture
. Implementation and testing
. Deployment and go-live support

== Terms and Conditions

=== Payment Terms

Payment is due within *30 days* of invoice date.

Late payments:: Subject to a 1.5% monthly fee
Early payments:: Eligible for a 2% discount

=== Pricing

|===
| Service | Rate | Unit

| Consultation
| $150
| per hour

| Development
| $200
| per hour

| Support
| $100
| per hour
|===

=== Confidentiality

Both parties agree to maintain the confidentiality of all proprietary
information shared during the engagement.

NOTE: This agreement is governed by the laws of the State of California.

== Contact

For questions about this agreement, please contact us at
https://acme.com/legal[Acme Legal Department].`

		metadata, result, err := testsupport.RenderDOCXWithMetadata(source)
		Expect(err).ToNot(HaveOccurred())
		doc := openDocx(result)

		// Metadata
		Expect(metadata.Title).To(Equal("Service Agreement"))
		Expect(metadata.Authors).To(HaveLen(1))

		// Document structure: title and author with correct styles
		title := doc.findParagraph("Service Agreement")
		Expect(title).ToNot(BeNil())
		Expect(title.Style).To(Equal("Title"))

		author := doc.findParagraph("Acme Corp")
		Expect(author).ToNot(BeNil())
		Expect(author.Style).To(Equal("Subtitle"))

		// Section headings with correct styles
		overview := doc.findParagraph("Overview")
		Expect(overview).ToNot(BeNil())
		Expect(overview.Style).To(Equal("Heading2"))

		payment := doc.findParagraph("Payment Terms")
		Expect(payment).ToNot(BeNil())
		Expect(payment.Style).To(Equal("Heading3"))

		// Inline formatting: "Service Agreement" bold in text, "Acme Corp" italic
		boldRun := doc.findRun("Service Agreement")
		// The first match is the title paragraph; find the one in body text
		paras := doc.parseParagraphs()
		for _, p := range paras {
			if p.Style == "" { // regular paragraph
				for _, r := range p.Runs {
					if r.Text == "Service Agreement" {
						boldRun = &r
					}
				}
			}
		}
		if boldRun != nil {
			Expect(boldRun.Bold).To(BeTrue())
		}

		// Ordered list with numbering
		listItem := doc.findParagraph("Initial consultation")
		Expect(listItem).ToNot(BeNil())
		Expect(listItem.NumID).ToNot(BeEmpty())

		// Labeled list: term bold
		termRun := doc.findRun("Late payments")
		Expect(termRun).ToNot(BeNil())
		Expect(termRun.Bold).To(BeTrue())

		// Table exists with correct structure
		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		Expect(tables[0].Rows).To(HaveLen(4)) // header + 3 data rows

		// Header cell is bold
		headerRun := doc.findTableCellRun("Service")
		Expect(headerRun).ToNot(BeNil())
		Expect(headerRun.Bold).To(BeTrue())

		// Admonition
		notePara := doc.findParagraph("NOTE:")
		Expect(notePara).ToNot(BeNil())
		Expect(notePara.Style).To(Equal("Admonition"))

		// Hyperlink
		rel := doc.findRelationship("https://acme.com/legal")
		Expect(rel).ToNot(BeNil())
		Expect(rel.TargetMode).To(Equal("External"))
	})
})
