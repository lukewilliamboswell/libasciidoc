package types_test

import (
	"github.com/lukewilliamboswell/libasciidoc/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("document", func() {

	Describe("FrontMatter", func() {

		It("should return front matter when present", func() {
			fm := &types.FrontMatter{Attributes: types.Attributes{"title": "My Doc"}}
			doc := &types.Document{
				Elements: []interface{}{fm},
			}
			Expect(doc.FrontMatter()).To(Equal(fm))
		})

		It("should return nil when first element is not front matter", func() {
			doc := &types.Document{
				Elements: []interface{}{
					&types.Paragraph{},
				},
			}
			Expect(doc.FrontMatter()).To(BeNil())
		})
	})

	Describe("Header", func() {

		It("should return header when first element", func() {
			header := &types.DocumentHeader{
				Title: []interface{}{&types.StringElement{Content: "Title"}},
			}
			doc := &types.Document{
				Elements: []interface{}{header},
			}
			h, idx := doc.Header()
			Expect(h).To(Equal(header))
			Expect(idx).To(Equal(0))
		})

		It("should return header after front matter and blank lines", func() {
			header := &types.DocumentHeader{
				Title: []interface{}{&types.StringElement{Content: "Title"}},
			}
			doc := &types.Document{
				Elements: []interface{}{
					&types.FrontMatter{},
					&types.BlankLine{},
					header,
				},
			}
			h, idx := doc.Header()
			Expect(h).To(Equal(header))
			Expect(idx).To(Equal(2))
		})

		It("should return nil when no header", func() {
			doc := &types.Document{
				Elements: []interface{}{
					&types.Paragraph{},
				},
			}
			h, idx := doc.Header()
			Expect(h).To(BeNil())
			Expect(idx).To(Equal(-1))
		})

		It("should return nil for empty document", func() {
			doc := &types.Document{
				Elements: []interface{}{},
			}
			h, idx := doc.Header()
			Expect(h).To(BeNil())
			Expect(idx).To(Equal(-1))
		})
	})

	Describe("BodyElements", func() {

		It("should return nil for empty document", func() {
			doc := &types.Document{
				Elements: []interface{}{},
			}
			Expect(doc.BodyElements()).To(BeNil())
		})

		It("should return all elements when no header", func() {
			p := &types.Paragraph{}
			doc := &types.Document{
				Elements: []interface{}{p},
			}
			Expect(doc.BodyElements()).To(Equal([]interface{}{p}))
		})

		It("should return elements after header", func() {
			header := &types.DocumentHeader{
				Title: []interface{}{&types.StringElement{Content: "Title"}},
			}
			p := &types.Paragraph{}
			doc := &types.Document{
				Elements: []interface{}{header, p},
			}
			Expect(doc.BodyElements()).To(Equal([]interface{}{p}))
		})

		It("should include front matter and blank lines before header in body", func() {
			fm := &types.FrontMatter{}
			bl := &types.BlankLine{}
			header := &types.DocumentHeader{
				Title: []interface{}{&types.StringElement{Content: "Title"}},
			}
			p := &types.Paragraph{}
			doc := &types.Document{
				Elements: []interface{}{fm, bl, header, p},
			}
			body := doc.BodyElements()
			Expect(body).To(HaveLen(3))
			Expect(body[0]).To(Equal(fm))
			Expect(body[1]).To(Equal(bl))
			Expect(body[2]).To(Equal(p))
		})
	})

	Describe("AddElement", func() {

		It("should append element", func() {
			doc := &types.Document{
				Elements: []interface{}{},
			}
			p := &types.Paragraph{}
			err := doc.AddElement(p)
			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Elements).To(HaveLen(1))
			Expect(doc.Elements[0]).To(Equal(p))
		})
	})
})

var _ = Describe("document author", func() {

	Describe("NewDocumentAuthorFullName", func() {

		It("should handle first name only", func() {
			name, err := types.NewDocumentAuthorFullName("John", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(name.FirstName).To(Equal("John"))
			Expect(name.MiddleName).To(BeEmpty())
			Expect(name.LastName).To(BeEmpty())
		})

		It("should handle first and last name", func() {
			name, err := types.NewDocumentAuthorFullName("John", "Doe", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(name.FirstName).To(Equal("John"))
			Expect(name.MiddleName).To(BeEmpty())
			Expect(name.LastName).To(Equal("Doe"))
		})

		It("should handle first, middle, and last name", func() {
			name, err := types.NewDocumentAuthorFullName("John", "M", "Doe")
			Expect(err).NotTo(HaveOccurred())
			Expect(name.FirstName).To(Equal("John"))
			Expect(name.MiddleName).To(Equal("M"))
			Expect(name.LastName).To(Equal("Doe"))
		})

		It("should replace underscores with spaces", func() {
			name, err := types.NewDocumentAuthorFullName("John_Paul", "Van_Der", "Berg")
			Expect(err).NotTo(HaveOccurred())
			Expect(name.FirstName).To(Equal("John Paul"))
			Expect(name.MiddleName).To(Equal("Van Der"))
			Expect(name.LastName).To(Equal("Berg"))
		})
	})

	Describe("FullName", func() {

		It("should return first name only", func() {
			name := &types.DocumentAuthorFullName{FirstName: "John"}
			Expect(name.FullName()).To(Equal("John"))
		})

		It("should return first and last name", func() {
			name := &types.DocumentAuthorFullName{FirstName: "John", LastName: "Doe"}
			Expect(name.FullName()).To(Equal("John Doe"))
		})

		It("should return full name with middle", func() {
			name := &types.DocumentAuthorFullName{FirstName: "John", MiddleName: "M", LastName: "Doe"}
			Expect(name.FullName()).To(Equal("John M Doe"))
		})
	})

	Describe("Initials", func() {

		It("should return initials for first name only", func() {
			name := &types.DocumentAuthorFullName{FirstName: "John"}
			Expect(name.Initials()).To(Equal("J"))
		})

		It("should return initials for full name", func() {
			name := &types.DocumentAuthorFullName{FirstName: "John", MiddleName: "Michael", LastName: "Doe"}
			Expect(name.Initials()).To(Equal("JMD"))
		})
	})

	Describe("NewDocumentAuthor", func() {

		It("should create author with full name and email", func() {
			fullName, _ := types.NewDocumentAuthorFullName("John", "Doe", nil)
			author, err := types.NewDocumentAuthor(fullName, "john@example.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(author.FirstName).To(Equal("John"))
			Expect(author.LastName).To(Equal("Doe"))
			Expect(author.Email).To(Equal("john@example.com"))
		})

		It("should trim email spaces", func() {
			fullName, _ := types.NewDocumentAuthorFullName("John", nil, nil)
			author, err := types.NewDocumentAuthor(fullName, "  john@example.com  ")
			Expect(err).NotTo(HaveOccurred())
			Expect(author.Email).To(Equal("john@example.com"))
		})

		It("should handle nil full name and email", func() {
			author, err := types.NewDocumentAuthor(nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(author.DocumentAuthorFullName).To(BeNil())
			Expect(author.Email).To(BeEmpty())
		})
	})

	Describe("NewDocumentAuthors", func() {

		It("should create authors array", func() {
			fullName, _ := types.NewDocumentAuthorFullName("John", "Doe", nil)
			author, _ := types.NewDocumentAuthor(fullName, "john@example.com")
			authors, err := types.NewDocumentAuthors(author)
			Expect(err).NotTo(HaveOccurred())
			Expect(authors).To(HaveLen(1))
			Expect(authors[0]).To(Equal(author))
		})

		It("should return error for non-author type", func() {
			_, err := types.NewDocumentAuthors("not-an-author")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DocumentAuthors.Expand", func() {

		It("should expand single author", func() {
			fullName, _ := types.NewDocumentAuthorFullName("John", "Michael", "Doe")
			author, _ := types.NewDocumentAuthor(fullName, "john@example.com")
			authors := types.DocumentAuthors{author}
			attrs := authors.Expand()
			Expect(attrs["author"]).To(Equal("John Michael Doe"))
			Expect(attrs["authorinitials"]).To(Equal("JMD"))
			Expect(attrs["firstname"]).To(Equal("John"))
			Expect(attrs["middlename"]).To(Equal("Michael"))
			Expect(attrs["lastname"]).To(Equal("Doe"))
			Expect(attrs["email"]).To(Equal("john@example.com"))
		})

		It("should use numbered keys for subsequent authors", func() {
			name1, _ := types.NewDocumentAuthorFullName("John", "Doe", nil)
			author1, _ := types.NewDocumentAuthor(name1, "")
			name2, _ := types.NewDocumentAuthorFullName("Jane", "Smith", nil)
			author2, _ := types.NewDocumentAuthor(name2, "")
			authors := types.DocumentAuthors{author1, author2}
			attrs := authors.Expand()
			// first author uses bare keys
			Expect(attrs["author"]).To(Equal("John Doe"))
			Expect(attrs["firstname"]).To(Equal("John"))
			// second author uses numbered keys
			Expect(attrs["author_2"]).To(Equal("Jane Smith"))
			Expect(attrs["firstname_2"]).To(Equal("Jane"))
		})
	})
})

var _ = Describe("document revision", func() {

	Describe("NewDocumentRevision", func() {

		It("should strip v prefix from revnumber", func() {
			rev, err := types.NewDocumentRevision("v1.0", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rev.Revnumber).To(Equal("1.0"))
		})

		It("should strip V prefix from revnumber", func() {
			rev, err := types.NewDocumentRevision("V2.0", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rev.Revnumber).To(Equal("2.0"))
		})

		It("should trim spaces from revdate", func() {
			rev, err := types.NewDocumentRevision(nil, "  2024-01-01  ", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rev.Revdate).To(Equal("2024-01-01"))
		})

		It("should strip colon prefix from revremark", func() {
			rev, err := types.NewDocumentRevision(nil, nil, ": some remark")
			Expect(err).NotTo(HaveOccurred())
			Expect(rev.Revremark).To(Equal("some remark"))
		})

		It("should handle all nil inputs", func() {
			rev, err := types.NewDocumentRevision(nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rev.Revnumber).To(BeEmpty())
			Expect(rev.Revdate).To(BeEmpty())
			Expect(rev.Revremark).To(BeEmpty())
		})

		It("should handle all fields", func() {
			rev, err := types.NewDocumentRevision("v3.0", "2024-06-15", ": Initial release")
			Expect(err).NotTo(HaveOccurred())
			Expect(rev.Revnumber).To(Equal("3.0"))
			Expect(rev.Revdate).To(Equal("2024-06-15"))
			Expect(rev.Revremark).To(Equal("Initial release"))
		})
	})

	Describe("Expand", func() {

		It("should expand all fields", func() {
			rev := &types.DocumentRevision{
				Revnumber: "1.0",
				Revdate:   "2024-01-01",
				Revremark: "First release",
			}
			attrs := rev.Expand()
			Expect(attrs["revnumber"]).To(Equal("1.0"))
			Expect(attrs["revdate"]).To(Equal("2024-01-01"))
			Expect(attrs["revremark"]).To(Equal("First release"))
			Expect(attrs[types.AttrRevision]).To(Equal(rev))
		})

		It("should skip empty fields", func() {
			rev := &types.DocumentRevision{
				Revdate: "2024-01-01",
			}
			attrs := rev.Expand()
			Expect(attrs).NotTo(HaveKey("revnumber"))
			Expect(attrs["revdate"]).To(Equal("2024-01-01"))
			Expect(attrs).NotTo(HaveKey("revremark"))
		})
	})
})

var _ = Describe("location", func() {

	Describe("ToString", func() {

		It("should return empty string for nil location", func() {
			var l *types.Location
			Expect(l.ToString()).To(BeEmpty())
		})

		It("should return path without scheme", func() {
			l := &types.Location{Path: "images/photo.png"}
			Expect(l.ToString()).To(Equal("images/photo.png"))
		})

		It("should return scheme and path", func() {
			l := &types.Location{Scheme: "https://", Path: "example.com"}
			Expect(l.ToString()).To(Equal("https://example.com"))
		})
	})

	Describe("ToDisplayString", func() {

		It("should return empty string for nil location", func() {
			var l *types.Location
			Expect(l.ToDisplayString()).To(BeEmpty())
		})

		It("should omit mailto scheme", func() {
			l := &types.Location{Scheme: "mailto:", Path: "user@example.com"}
			Expect(l.ToDisplayString()).To(Equal("user@example.com"))
		})

		It("should include non-mailto scheme", func() {
			l := &types.Location{Scheme: "https://", Path: "example.com"}
			Expect(l.ToDisplayString()).To(Equal("https://example.com"))
		})
	})

	Describe("TrimAngleBracketSuffix", func() {

		It("should trim > from string path", func() {
			l := &types.Location{Path: "https://example.com>"}
			trimmed, err := l.TrimAngleBracketSuffix()
			Expect(err).NotTo(HaveOccurred())
			Expect(trimmed).To(BeTrue())
			Expect(l.Path).To(Equal("https://example.com"))
		})

		It("should return false when no > suffix on string path", func() {
			l := &types.Location{Path: "https://example.com"}
			trimmed, err := l.TrimAngleBracketSuffix()
			Expect(err).NotTo(HaveOccurred())
			Expect(trimmed).To(BeFalse())
		})

		It("should trim SpecialCharacter > from slice path", func() {
			l := &types.Location{
				Path: []interface{}{
					&types.StringElement{Content: "example.com"},
					&types.SpecialCharacter{Name: ">"},
				},
			}
			trimmed, err := l.TrimAngleBracketSuffix()
			Expect(err).NotTo(HaveOccurred())
			Expect(trimmed).To(BeTrue())
		})
	})
})

var _ = Describe("table of contents", func() {

	Describe("Add", func() {

		It("should add top-level section", func() {
			toc := &types.TableOfContents{MaxDepth: 3}
			s := &types.Section{
				Level:      1,
				Attributes: types.Attributes{types.AttrID: "section1"},
			}
			toc.Add(s)
			Expect(toc.Sections).To(HaveLen(1))
			Expect(toc.Sections[0].ID).To(Equal("section1"))
			Expect(toc.Sections[0].Level).To(Equal(1))
		})

		It("should skip section exceeding max depth", func() {
			toc := &types.TableOfContents{MaxDepth: 1}
			s := &types.Section{
				Level:      2,
				Attributes: types.Attributes{types.AttrID: "deep"},
			}
			toc.Add(s)
			Expect(toc.Sections).To(BeEmpty())
		})

		It("should add sibling sections at same level", func() {
			toc := &types.TableOfContents{MaxDepth: 3}
			s1 := &types.Section{
				Level:      1,
				Attributes: types.Attributes{types.AttrID: "s1"},
			}
			s2 := &types.Section{
				Level:      1,
				Attributes: types.Attributes{types.AttrID: "s2"},
			}
			toc.Add(s1)
			toc.Add(s2)
			Expect(toc.Sections).To(HaveLen(2))
			Expect(toc.Sections[0].ID).To(Equal("s1"))
			Expect(toc.Sections[1].ID).To(Equal("s2"))
		})

		It("should nest child section under parent", func() {
			toc := &types.TableOfContents{MaxDepth: 3}
			parent := &types.Section{
				Level:      1,
				Attributes: types.Attributes{types.AttrID: "parent"},
			}
			child := &types.Section{
				Level:      2,
				Attributes: types.Attributes{types.AttrID: "child"},
			}
			toc.Add(parent)
			toc.Add(child)
			Expect(toc.Sections).To(HaveLen(1))
			Expect(toc.Sections[0].Children).To(HaveLen(1))
			Expect(toc.Sections[0].Children[0].ID).To(Equal("child"))
		})

		It("should nest grandchild section", func() {
			toc := &types.TableOfContents{MaxDepth: 3}
			s1 := &types.Section{Level: 1, Attributes: types.Attributes{types.AttrID: "s1"}}
			s2 := &types.Section{Level: 2, Attributes: types.Attributes{types.AttrID: "s2"}}
			s3 := &types.Section{Level: 3, Attributes: types.Attributes{types.AttrID: "s3"}}
			toc.Add(s1)
			toc.Add(s2)
			toc.Add(s3)
			Expect(toc.Sections).To(HaveLen(1))
			Expect(toc.Sections[0].Children).To(HaveLen(1))
			Expect(toc.Sections[0].Children[0].Children).To(HaveLen(1))
			Expect(toc.Sections[0].Children[0].Children[0].ID).To(Equal("s3"))
		})
	})
})

var _ = Describe("utilities", func() {

	Describe("Reduce", func() {

		It("should return nil for nil input", func() {
			Expect(types.Reduce(nil)).To(BeNil())
		})

		It("should return nil for empty slice", func() {
			Expect(types.Reduce([]interface{}{})).To(BeNil())
		})

		It("should return string content from single StringElement slice", func() {
			result := types.Reduce([]interface{}{
				&types.StringElement{Content: "hello"},
			})
			Expect(result).To(Equal("hello"))
		})

		It("should apply options to string result", func() {
			result := types.Reduce("  hello  ", func(s string) string {
				return "trimmed"
			})
			Expect(result).To(Equal("trimmed"))
		})

		It("should return nil for empty string", func() {
			Expect(types.Reduce("")).To(BeNil())
		})

		It("should return non-empty string as-is", func() {
			Expect(types.Reduce("hello")).To(Equal("hello"))
		})

		It("should return non-slice non-string input as-is", func() {
			Expect(types.Reduce(42)).To(Equal(42))
		})
	})

	Describe("Flatten", func() {

		It("should flatten nested slices", func() {
			result := types.Flatten([]interface{}{
				[]interface{}{"a", "b"},
				"c",
			})
			Expect(result).To(Equal([]interface{}{"a", "b", "c"}))
		})

		It("should handle empty input", func() {
			result := types.Flatten([]interface{}{})
			Expect(result).To(BeEmpty())
		})
	})

	Describe("AllNilEntries", func() {

		It("should return true for all nil entries", func() {
			Expect(types.AllNilEntries([]interface{}{nil, nil})).To(BeTrue())
		})

		It("should return false when non-nil entry exists", func() {
			Expect(types.AllNilEntries([]interface{}{nil, "value"})).To(BeFalse())
		})

		It("should handle nested nil slices", func() {
			Expect(types.AllNilEntries([]interface{}{
				[]interface{}{nil},
				nil,
			})).To(BeTrue())
		})

		It("should return true for empty slice", func() {
			Expect(types.AllNilEntries([]interface{}{})).To(BeTrue())
		})
	})

	Describe("Append", func() {

		It("should append scalars and slices", func() {
			result, err := types.Append("a", []interface{}{"b", "c"}, "d")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]interface{}{"a", "b", "c", "d"}))
		})

		It("should skip nil elements", func() {
			result, err := types.Append("a", nil, "b")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]interface{}{"a", "b"}))
		})

		It("should handle empty input", func() {
			result, err := types.Append()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	Describe("Apply", func() {

		It("should chain transformations", func() {
			result := types.Apply("  hello  ",
				func(s string) string { return s + "!" },
				func(s string) string { return s + "?" },
			)
			Expect(result).To(Equal("  hello  !?"))
		})

		It("should return source when no funcs", func() {
			Expect(types.Apply("hello")).To(Equal("hello"))
		})
	})
})
