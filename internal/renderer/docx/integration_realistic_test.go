package docx_test

// integration_realistic_test.go — end-to-end integration tests using realistic
// AsciiDoc documents that reflect actual authoring patterns.
//
// Unlike the focused unit tests in other files, each test here renders a
// complete, multi-feature document and validates that several OOXML layers
// interact correctly — styles, numbering, tables, links, and metadata together.
//
// These scenarios were chosen because they exercise common cross-cutting paths
// that unit tests miss:
//   - A heading change can break cross-reference anchor generation.
//   - A list inside a table exercises both the list renderer and cell flushing.
//   - A footnote in a document with hyperlinks exercises the relationship tracker.

import (
	"encoding/xml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lukewilliamboswell/libasciidoc/internal/testsupport"
)

var _ = Describe("realistic integration — technical API documentation", func() {

	// A document representing Go library API docs: structured headings, a
	// parameter reference table, inline code, an external link, and a footnote.
	const apiDocsSource = `= HTTP Client Library
Jane Developer <jane@example.com>
v2.1, 2024-03-15

== Overview

The HTTP Client Library provides a simple interface for making HTTP requests.
Use the ` + "`Client`" + ` type as the entry point for all operations.footnote:[The ` + "`Client`" + ` type is safe for concurrent use by multiple goroutines.]

== Installation

Add the dependency using the standard Go toolchain:

[source,sh]
----
go get github.com/example/httpclient@v2
----

== API Reference

=== NewClient

Creates a new ` + "`Client`" + ` instance with the given base URL.

.Parameters
|===
| Parameter | Type | Description

| baseURL
| ` + "`string`" + `
| The base URL for all requests. Must include the scheme.

| timeout
| ` + "`time.Duration`" + `
| Maximum time to wait for a response.

| retries
| ` + "`int`" + `
| Number of retry attempts on transient failure.
|===

For full documentation see https://pkg.go.dev/example/httpclient[pkg.go.dev].

== Error Handling

All methods return a typed error value. Callers should check errors with
` + "`errors.As`" + ` to distinguish transient failures from permanent ones:

. Check for ` + "`*httpclient.NetworkError`" + ` — indicates a transient failure.
. Check for ` + "`*httpclient.AuthError`" + ` — indicates an authentication failure.
. Otherwise treat the error as permanent.`

	It("should render title with Title style and author with Subtitle style", func() {
		doc := renderDocx(apiDocsSource)

		title := doc.findParagraph("HTTP Client Library")
		Expect(title).ToNot(BeNil())
		Expect(title.Style).To(Equal("Title"))

		author := doc.findParagraph("Jane Developer")
		Expect(author).ToNot(BeNil())
		Expect(author.Style).To(Equal("Subtitle"))
	})

	It("should render section headings at the correct heading levels", func() {
		doc := renderDocx(apiDocsSource)

		overview := doc.findParagraph("Overview")
		Expect(overview).ToNot(BeNil())
		Expect(overview.Style).To(Equal("Heading2"))

		apiRef := doc.findParagraph("API Reference")
		Expect(apiRef).ToNot(BeNil())
		Expect(apiRef.Style).To(Equal("Heading2"))

		newClient := doc.findParagraph("NewClient")
		Expect(newClient).ToNot(BeNil())
		Expect(newClient.Style).To(Equal("Heading3"))
	})

	It("should render the source block with CodeBlock style", func() {
		doc := renderDocx(apiDocsSource)

		codeBlock := doc.findParagraph("go get github.com")
		Expect(codeBlock).ToNot(BeNil())
		Expect(codeBlock.Style).To(Equal("CodeBlock"))
	})

	It("should render the parameter table with header row and correct dimensions", func() {
		doc := renderDocx(apiDocsSource)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1), "expected exactly one table (the parameter reference)")

		table := tables[0]
		Expect(table.Rows).To(HaveLen(4), "1 header row + 3 parameter rows")

		// Header row must be marked as a repeating table header (ECMA-376 §17.4.27)
		Expect(table.Rows[0].IsHeader).To(BeTrue(),
			"the first row of a table with header must carry w:tblHeader")

		// Data rows must not be marked as headers
		for i := 1; i < len(table.Rows); i++ {
			Expect(table.Rows[i].IsHeader).To(BeFalse(),
				"data row %d must not carry w:tblHeader", i)
		}

		// Header cell should be bold
		headerRun := doc.findTableCellRun("Parameter")
		Expect(headerRun).ToNot(BeNil())
		Expect(headerRun.Bold).To(BeTrue())
	})

	It("should render the external link with a hyperlink relationship", func() {
		doc := renderDocx(apiDocsSource)

		rel := doc.findRelationship("https://pkg.go.dev/example/httpclient")
		Expect(rel).ToNot(BeNil(), "expected a relationship for the pkg.go.dev link")
		Expect(rel.TargetMode).To(Equal("External"))
	})

	It("should render footnote reference in document and body in footnotes.xml", func() {
		doc := renderDocx(apiDocsSource)

		Expect(doc.files).To(HaveKey("word/footnotes.xml"))
		Expect(doc.footnotesXML()).To(ContainSubstring("safe for concurrent use"),
			"footnotes.xml must contain the footnote body text")

		// Verify the footnote reference marker appears in the document
		paras := doc.parseParagraphs()
		foundRef := false
		for _, p := range paras {
			for _, r := range p.Runs {
				if r.FootnoteRef {
					foundRef = true
				}
			}
		}
		Expect(foundRef).To(BeTrue(), "document must contain a w:footnoteReference element")
	})

	It("should render the ordered error-handling list with decimal numbering", func() {
		doc := renderDocx(apiDocsSource)

		p := doc.findParagraph("NetworkError")
		Expect(p).ToNot(BeNil())
		Expect(p.NumID).ToNot(BeEmpty())

		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("decimal"))
	})

	It("should produce well-formed XML across all OOXML parts", func() {
		_, result, err := testsupport.RenderDOCXWithMetadata(apiDocsSource)
		Expect(err).ToNot(HaveOccurred())
		doc := openDocx(result)
		for _, partName := range []string{
			"word/document.xml",
			"word/styles.xml",
			"word/numbering.xml",
			"word/footnotes.xml",
		} {
			content := doc.files[partName]
			Expect(content).ToNot(BeEmpty(), "%s must not be empty", partName)
			Expect(xml.Unmarshal(content, new(interface{}))).To(Succeed(),
				"%s must be well-formed XML — check ooxml.go for malformed string literals", partName)
		}
	})
})

var _ = Describe("realistic integration — meeting minutes", func() {

	const meetingSource = `= Q2 Engineering Planning — Meeting Minutes
Engineering Team
2024-04-10

== Attendees

* Alice Smith (Engineering Lead)
* Bob Jones (Backend)
* Carol Lee (Frontend)
* Dave Park (QA)

== Agenda and Decisions

=== Infrastructure Migration

The team agreed to migrate the remaining services to Kubernetes by end of quarter.

Action items:

* [ ] Bob: audit existing Docker Compose files
* [x] Alice: provision staging cluster (completed)
* [ ] Dave: update smoke-test suite for new endpoints

=== Release Planning

The next release is scheduled for 2024-05-01.

[NOTE]
====
All feature branches must be merged by 2024-04-26 to allow sufficient QA time.
====

Ordered milestones:

. Feature freeze — 2024-04-26
. QA sign-off — 2024-04-28
. Release candidate — 2024-04-30
. Production deployment — 2024-05-01

'''

== Next Meeting

2024-04-17 at 10:00 UTC — same channel.`

	It("should render title and author with correct styles", func() {
		doc := renderDocx(meetingSource)

		title := doc.findParagraph("Q2 Engineering Planning")
		Expect(title).ToNot(BeNil())
		Expect(title.Style).To(Equal("Title"))

		subtitle := doc.findParagraph("Engineering Team")
		Expect(subtitle).ToNot(BeNil())
		Expect(subtitle.Style).To(Equal("Subtitle"))
	})

	It("should render section headings at heading level 2 and 3", func() {
		doc := renderDocx(meetingSource)

		attendees := doc.findParagraph("Attendees")
		Expect(attendees).ToNot(BeNil())
		Expect(attendees.Style).To(Equal("Heading2"))

		infra := doc.findParagraph("Infrastructure Migration")
		Expect(infra).ToNot(BeNil())
		Expect(infra.Style).To(Equal("Heading3"))
	})

	It("should render the attendees list as a bulleted unordered list", func() {
		doc := renderDocx(meetingSource)

		p := doc.findParagraph("Alice Smith")
		Expect(p).ToNot(BeNil())
		Expect(p.NumID).ToNot(BeEmpty(), "attendee list items must have numbering")

		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("bullet"))
	})

	It("should render the checklist items with checkbox characters", func() {
		doc := renderDocx(meetingSource)
		// [x] → ☑ (U+2611), [ ] → ☐ (U+2610)
		Expect(doc.text()).To(ContainSubstring("\u2611"), "completed item must show ☑")
		Expect(doc.text()).To(ContainSubstring("\u2610"), "open item must show ☐")
	})

	It("should render the NOTE admonition with Admonition style", func() {
		doc := renderDocx(meetingSource)

		// Block admonitions produce a label paragraph with the Admonition style
		notePara := doc.findParagraph("NOTE:")
		Expect(notePara).ToNot(BeNil())
		Expect(notePara.Style).To(Equal("Admonition"))
	})

	It("should render the thematic break as a paragraph border, not dashes", func() {
		// ECMA-376 §17.3.1.5 w:pBdr — thematic break renders as bottom border
		doc := renderDocx(meetingSource)
		Expect(doc.documentXML()).To(ContainSubstring(`<w:pBdr>`),
			"''' must produce a w:pBdr element")
		Expect(doc.documentXML()).ToNot(ContainSubstring("———"),
			"''' must not produce literal dash characters")
	})

	It("should render the ordered milestones list with decimal numbering", func() {
		doc := renderDocx(meetingSource)

		freeze := doc.findParagraph("Feature freeze")
		Expect(freeze).ToNot(BeNil())
		Expect(freeze.NumID).ToNot(BeEmpty())

		def := doc.findNumberingDef(freeze.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("decimal"))
	})

	It("should produce all required OOXML parts", func() {
		doc := renderDocx(meetingSource)
		Expect(doc.files).To(HaveKey("word/document.xml"))
		Expect(doc.files).To(HaveKey("word/styles.xml"))
		Expect(doc.files).To(HaveKey("word/numbering.xml"))
	})
})

var _ = Describe("realistic integration — product comparison spec", func() {

	const specSource = `= Widget Pro vs Widget Lite — Feature Comparison
Product Team

[#overview]
== Overview

This document compares the two product tiers.
See <<features,the comparison table>> for the detailed feature matrix.

[#features]
== Feature Matrix

[cols="2,1,1,3"]
|===
| Feature | Pro | Lite | Notes

| API rate limit
| 10 000/min
| 1 000/min
| Lite tier is suitable for small applications.

| Storage
| 1 TB
| 10 GB
| Pro storage is expandable.

| SLA uptime
| 99.99%
| 99.9%
| Both tiers include 24/7 monitoring.

| Support
| Dedicated
| Community
| Pro includes a named account manager.
|===

== Key Advantages of Pro

. Higher API throughput — suitable for production workloads
. Dedicated support with guaranteed response times
. Expandable storage without service interruption
. Advanced analytics and audit logs`

	It("should render the feature table with correct row and column counts", func() {
		doc := renderDocx(specSource)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))

		table := tables[0]
		Expect(table.Rows).To(HaveLen(5), "1 header row + 4 feature rows")
		for i, row := range table.Rows {
			Expect(row.Cells).To(HaveLen(4),
				"every row must have exactly 4 cells (row %d)", i)
		}
	})

	It("should mark the header row with w:tblHeader and data rows without", func() {
		doc := renderDocx(specSource)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))

		Expect(tables[0].Rows[0].IsHeader).To(BeTrue())
		for i := 1; i < len(tables[0].Rows); i++ {
			Expect(tables[0].Rows[i].IsHeader).To(BeFalse(),
				"data row %d must not carry w:tblHeader", i)
		}
	})

	It("should produce proportional column widths for cols=2,1,1,3", func() {
		// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/tables/add-columns/
		// ECMA-376 §17.4.13 w:gridCol — column widths derived from page text width
		doc := renderDocx(specSource)

		tables := doc.parseTables()
		Expect(tables).To(HaveLen(1))
		widths := tables[0].GridColWidths
		Expect(widths).To(HaveLen(4))

		total := widths[0] + widths[1] + widths[2] + widths[3]
		Expect(total).To(BeNumerically(">", 0))

		// Ratios: 2:1:1:3 → col0 = 2/7, col1 = col2 = 1/7, col3 = 3/7.
		// First three columns floor-divide; last column absorbs the residual,
		// so allow it a wider tolerance equal to the column count in twips.
		unit := total / 7
		Expect(widths[0]).To(BeNumerically("~", 2*unit, 2),
			"first column (ratio 2) width should be ~2/7 of total")
		Expect(widths[1]).To(BeNumerically("~", unit, 2),
			"second column (ratio 1) width should be ~1/7 of total")
		Expect(widths[2]).To(BeNumerically("~", unit, 2),
			"third column (ratio 1) width should be ~1/7 of total")
		Expect(widths[3]).To(BeNumerically("~", 3*unit, 6),
			"fourth column (ratio 3) absorbs the rounding residual")
	})

	It("should render cross-references as internal hyperlinks anchored to section bookmarks", func() {
		// AsciiDoc spec: https://docs.asciidoctor.org/asciidoc/latest/macros/xref/
		// ECMA-376 §17.18.1 w:hyperlink with w:anchor
		doc := renderDocx(specSource)

		// The "Feature Matrix" section must have a bookmark
		featuresPara := doc.findParagraph("Feature Matrix")
		Expect(featuresPara).ToNot(BeNil())
		Expect(featuresPara.Bookmarks).ToNot(BeEmpty(),
			"[#features] section must produce a w:bookmarkStart")
		featuresAnchor := featuresPara.Bookmarks[0]

		// The <<features,the comparison table>> cross-reference must link to that bookmark
		overviewPara := doc.findParagraph("comparison table")
		Expect(overviewPara).ToNot(BeNil())
		Expect(overviewPara.Links).ToNot(BeEmpty(),
			"<<features,...>> must produce a w:hyperlink")
		Expect(overviewPara.Links[0].Anchor).To(Equal(featuresAnchor),
			"the xref anchor must exactly match the section bookmark name")
	})

	It("should render the Pro advantages list with decimal numbering", func() {
		doc := renderDocx(specSource)

		p := doc.findParagraph("Higher API throughput")
		Expect(p).ToNot(BeNil())
		Expect(p.NumID).ToNot(BeEmpty())

		def := doc.findNumberingDef(p.NumID)
		Expect(def).ToNot(BeNil())
		Expect(def.Levels[0].Format).To(Equal("decimal"))
	})

	It("should produce well-formed XML for all OOXML parts", func() {
		doc := renderDocx(specSource)
		for _, partName := range []string{
			"word/document.xml",
			"word/styles.xml",
			"word/numbering.xml",
		} {
			content := doc.files[partName]
			Expect(content).ToNot(BeEmpty(), "%s must not be empty", partName)
			Expect(xml.Unmarshal(content, new(interface{}))).To(Succeed(),
				"%s must be well-formed XML", partName)
		}
	})
})
