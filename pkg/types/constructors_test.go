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

var _ = Describe("DocumentHeader", func() {

	Describe("NewDocumentHeader", func() {

		It("should create header with title only", func() {
			title := []interface{}{&types.StringElement{Content: "My Title"}}
			header, err := types.NewDocumentHeader(title, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Title).To(HaveLen(1))
			Expect(header.Elements).To(BeEmpty())
		})

		It("should create header with authors and revision", func() {
			title := []interface{}{&types.StringElement{Content: "Title"}}
			authors := types.DocumentAuthors{
				&types.DocumentAuthor{
					DocumentAuthorFullName: &types.DocumentAuthorFullName{FirstName: "John"},
				},
			}
			revision := &types.DocumentRevision{Revnumber: "1.0"}
			ar := &types.DocumentAuthorsAndRevision{
				Authors:  authors,
				Revision: revision,
			}
			header, err := types.NewDocumentHeader(title, ar, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Title).To(HaveLen(1))
			Expect(header.Elements).To(HaveLen(2)) // authors + revision declarations
		})

		It("should include extra attributes", func() {
			title := []interface{}{&types.StringElement{Content: "Title"}}
			extra := []interface{}{
				&types.AttributeDeclaration{Name: "toc", Value: ""},
			}
			header, err := types.NewDocumentHeader(title, nil, extra)
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Elements).To(HaveLen(1))
		})
	})

	Describe("Authors", func() {

		It("should return authors from AttrAuthors declaration", func() {
			authors := types.DocumentAuthors{
				&types.DocumentAuthor{
					DocumentAuthorFullName: &types.DocumentAuthorFullName{FirstName: "Jane"},
					Email:                 "jane@example.com",
				},
			}
			header := &types.DocumentHeader{
				Elements: []interface{}{
					&types.AttributeDeclaration{
						Name:  types.AttrAuthors,
						Value: authors,
					},
				},
			}
			result := header.Authors()
			Expect(result).To(HaveLen(1))
			Expect(result[0].DocumentAuthorFullName.FirstName).To(Equal("Jane"))
		})

		It("should return authors from individual AttrAuthor and AttrEmail", func() {
			header := &types.DocumentHeader{
				Elements: []interface{}{
					&types.AttributeDeclaration{Name: types.AttrAuthor, Value: "John"},
					&types.AttributeDeclaration{Name: types.AttrEmail, Value: "john@example.com"},
				},
			}
			result := header.Authors()
			Expect(result).To(HaveLen(1))
			Expect(result[0].DocumentAuthorFullName.FirstName).To(Equal("John"))
			Expect(result[0].Email).To(Equal("john@example.com"))
		})

		It("should return nil when no author declarations", func() {
			header := &types.DocumentHeader{
				Elements: []interface{}{
					&types.AttributeDeclaration{Name: "toc", Value: ""},
				},
			}
			Expect(header.Authors()).To(BeNil())
		})
	})

	Describe("Revision", func() {

		It("should return revision when present", func() {
			rev := &types.DocumentRevision{Revnumber: "2.0"}
			header := &types.DocumentHeader{
				Elements: []interface{}{
					&types.AttributeDeclaration{Name: types.AttrRevision, Value: rev},
				},
			}
			result := header.Revision()
			Expect(result).NotTo(BeNil())
			Expect(result.Revnumber).To(Equal("2.0"))
		})

		It("should return nil when no revision", func() {
			header := &types.DocumentHeader{
				Elements: []interface{}{},
			}
			Expect(header.Revision()).To(BeNil())
		})
	})

	Describe("IsEmpty", func() {

		It("should be true for empty header", func() {
			header := &types.DocumentHeader{}
			Expect(header.IsEmpty()).To(BeTrue())
		})

		It("should be false with title", func() {
			header := &types.DocumentHeader{
				Title: []interface{}{&types.StringElement{Content: "Title"}},
			}
			Expect(header.IsEmpty()).To(BeFalse())
		})
	})
})

var _ = Describe("NewAttributeReference", func() {

	It("should return PredefinedAttribute for known names", func() {
		result, err := types.NewAttributeReference("nbsp", "{nbsp}")
		Expect(err).NotTo(HaveOccurred())
		_, ok := result.(*types.PredefinedAttribute)
		Expect(ok).To(BeTrue())
	})

	It("should return AttributeReference for unknown names", func() {
		result, err := types.NewAttributeReference("custom-attr", "{custom-attr}")
		Expect(err).NotTo(HaveOccurred())
		_, ok := result.(*types.AttributeReference)
		Expect(ok).To(BeTrue())
	})
})

var _ = Describe("NewDocumentAuthorsAndRevision", func() {

	It("should create with authors and revision", func() {
		authors := types.DocumentAuthors{
			&types.DocumentAuthor{
				DocumentAuthorFullName: &types.DocumentAuthorFullName{FirstName: "Jane"},
			},
		}
		rev := &types.DocumentRevision{Revnumber: "1.0"}
		ar, err := types.NewDocumentAuthorsAndRevision(authors, rev)
		Expect(err).NotTo(HaveOccurred())
		Expect(ar.Authors).To(HaveLen(1))
		Expect(ar.Revision).NotTo(BeNil())
	})

	It("should create with authors only", func() {
		authors := types.DocumentAuthors{}
		ar, err := types.NewDocumentAuthorsAndRevision(authors, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(ar.Revision).To(BeNil())
	})
})

var _ = Describe("Paragraph", func() {

	Describe("NewParagraph", func() {

		It("should append newlines to all but last raw line", func() {
			rl1, _ := types.NewRawLine("line1")
			rl2, _ := types.NewRawLine("line2")
			rl3, _ := types.NewRawLine("line3")
			p, err := types.NewParagraph(nil, rl1, rl2, rl3)
			Expect(err).NotTo(HaveOccurred())
			Expect(p.Elements).To(HaveLen(3))
			Expect(rl1.Content).To(Equal("line1\n"))
			Expect(rl2.Content).To(Equal("line2\n"))
			Expect(rl3.Content).To(Equal("line3"))
		})

		It("should set style attribute", func() {
			rl, _ := types.NewRawLine("content")
			p, err := types.NewParagraph("source", rl)
			Expect(err).NotTo(HaveOccurred())
			Expect(p.Attributes[types.AttrStyle]).To(Equal("source"))
		})

		It("should have nil attributes without style", func() {
			rl, _ := types.NewRawLine("content")
			p, err := types.NewParagraph(nil, rl)
			Expect(err).NotTo(HaveOccurred())
			Expect(p.Attributes).To(BeNil())
		})
	})

	Describe("mapAttributes via AddAttributes", func() {

		It("should map positional-1 to style", func() {
			rl, _ := types.NewRawLine("content")
			p, _ := types.NewParagraph(nil, rl)
			p.AddAttributes(types.Attributes{
				types.AttrPositional1: "source",
			})
			Expect(p.Attributes[types.AttrStyle]).To(Equal("source"))
		})

		It("should map positional-2 to language for source style", func() {
			rl, _ := types.NewRawLine("content")
			p, _ := types.NewParagraph(nil, rl)
			p.AddAttributes(types.Attributes{
				types.AttrPositional1: "source",
				types.AttrPositional2: "go",
			})
			Expect(p.Attributes[types.AttrStyle]).To(Equal("source"))
			Expect(p.Attributes[types.AttrLanguage]).To(Equal("go"))
		})

		It("should map positional-2 to quote author for quote style", func() {
			rl, _ := types.NewRawLine("content")
			p, _ := types.NewParagraph(nil, rl)
			p.AddAttributes(types.Attributes{
				types.AttrPositional1: "quote",
				types.AttrPositional2: "Author Name",
				types.AttrPositional3: "Book Title",
			})
			Expect(p.Attributes[types.AttrStyle]).To(Equal("quote"))
			Expect(p.Attributes[types.AttrQuoteAuthor]).To(Equal("Author Name"))
			Expect(p.Attributes[types.AttrQuoteTitle]).To(Equal("Book Title"))
		})
	})

	Describe("AddElement", func() {

		It("should append newline to last raw line before adding", func() {
			rl1, _ := types.NewRawLine("line1")
			rl2, _ := types.NewRawLine("line2")
			p, _ := types.NewParagraph(nil, rl1)
			err := p.AddElement(rl2)
			Expect(err).NotTo(HaveOccurred())
			Expect(p.Elements).To(HaveLen(2))
			Expect(rl1.Content).To(Equal("line1\n"))
		})
	})
})

var _ = Describe("Section", func() {

	Describe("ResolveID", func() {

		It("should use existing ID attribute", func() {
			s := &types.Section{
				Attributes: types.Attributes{
					types.AttrID: "existing-id",
				},
				Title: []interface{}{&types.StringElement{Content: "My Title"}},
			}
			refs := types.ElementReferences{}
			err := s.ResolveID(types.Attributes{}, refs)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Attributes[types.AttrID]).To(Equal("existing-id"))
		})

		It("should generate ID from title", func() {
			s := &types.Section{
				Title: []interface{}{&types.StringElement{Content: "My Title"}},
			}
			refs := types.ElementReferences{}
			err := s.ResolveID(types.Attributes{}, refs)
			Expect(err).NotTo(HaveOccurred())
			id, ok := s.Attributes[types.AttrID].(string)
			Expect(ok).To(BeTrue())
			Expect(id).To(Equal("_My_Title"))
		})

		It("should add suffix for duplicate IDs", func() {
			s := &types.Section{
				Title: []interface{}{&types.StringElement{Content: "My Title"}},
			}
			refs := types.ElementReferences{
				"_My_Title": "existing",
			}
			err := s.ResolveID(types.Attributes{}, refs)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Attributes[types.AttrID]).To(Equal("_My_Title_2"))
		})
	})

	Describe("SetTitle", func() {

		It("should extract inline ID attribute from trailing element", func() {
			s := &types.Section{}
			s.SetTitle([]interface{}{
				&types.StringElement{Content: "My Title"},
				&types.Attribute{Key: types.AttrID, Value: "custom-id"},
			})
			// Title should not contain the attribute
			Expect(s.Title).To(HaveLen(1))
			Expect(s.Attributes[types.AttrID]).To(Equal("custom-id"))
		})

		It("should trim trailing spaces from title", func() {
			s := &types.Section{}
			s.SetTitle([]interface{}{
				&types.StringElement{Content: "My Title   "},
			})
			se := s.Title[0].(*types.StringElement)
			Expect(se.Content).To(Equal("My Title"))
		})
	})
})

var _ = Describe("InternalCrossReference", func() {

	Describe("ResolveID", func() {

		It("should resolve string ID with spaces", func() {
			xref, _ := types.NewInternalCrossReference("my section", nil)
			err := xref.ResolveID(types.Attributes{})
			Expect(err).NotTo(HaveOccurred())
			Expect(xref.ID).To(Equal("_my_section")) // spaces replaced with separators
		})

		It("should not modify string ID without spaces", func() {
			xref, _ := types.NewInternalCrossReference("my_section", nil)
			err := xref.ResolveID(types.Attributes{})
			Expect(err).NotTo(HaveOccurred())
			Expect(xref.ID).To(Equal("my_section"))
		})

		It("should resolve slice ID", func() {
			xref, _ := types.NewInternalCrossReference(
				[]interface{}{&types.StringElement{Content: "My Section"}},
				nil,
			)
			err := xref.ResolveID(types.Attributes{})
			Expect(err).NotTo(HaveOccurred())
			Expect(xref.ID).To(Equal("_My_Section"))
		})
	})
})

var _ = Describe("YamlFrontMatter", func() {

	It("should parse valid YAML", func() {
		fm, err := types.NewYamlFrontMatter("title: My Doc\nauthor: Jane\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(fm.Attributes).To(HaveKeyWithValue("title", "My Doc"))
		Expect(fm.Attributes).To(HaveKeyWithValue("author", "Jane"))
	})

	It("should return nil attributes for empty YAML", func() {
		fm, err := types.NewYamlFrontMatter("")
		Expect(err).NotTo(HaveOccurred())
		Expect(fm.Attributes).To(BeNil())
	})

	It("should return error for invalid YAML", func() {
		_, err := types.NewYamlFrontMatter(":\n  - :\n    :\n  bad: [")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("CounterSubstitution", func() {

	It("should convert string value to rune", func() {
		cs, err := types.NewCounterSubstitution("mycount", false, "A", "{counter:mycount:A}")
		Expect(err).NotTo(HaveOccurred())
		Expect(cs.Name).To(Equal("mycount"))
		Expect(cs.Value).To(Equal('A'))
		Expect(cs.Hidden).To(BeFalse())
	})

	It("should keep non-string value as-is", func() {
		cs, err := types.NewCounterSubstitution("mycount", true, nil, "{counter2:mycount}")
		Expect(err).NotTo(HaveOccurred())
		Expect(cs.Value).To(BeNil())
		Expect(cs.Hidden).To(BeTrue())
	})
})

var _ = Describe("Preamble", func() {

	It("should have content with a paragraph", func() {
		p := &types.Preamble{
			Elements: []interface{}{
				&types.Paragraph{},
			},
		}
		Expect(p.HasContent()).To(BeTrue())
	})

	It("should not have content with only blank lines", func() {
		p := &types.Preamble{
			Elements: []interface{}{
				&types.BlankLine{},
				&types.AttributeDeclaration{Name: "foo"},
			},
		}
		Expect(p.HasContent()).To(BeFalse())
	})

	It("should not have content when empty", func() {
		p := &types.Preamble{}
		Expect(p.HasContent()).To(BeFalse())
	})
})

var _ = Describe("DelimitedBlock", func() {

	It("should append newlines to all but last raw line", func() {
		rl1, _ := types.NewRawLine("line1")
		rl2, _ := types.NewRawLine("line2")
		b, err := types.NewDelimitedBlock(types.Listing, []interface{}{rl1, rl2})
		Expect(err).NotTo(HaveOccurred())
		Expect(b.Kind).To(Equal(types.Listing))
		Expect(rl1.Content).To(Equal("line1\n"))
		Expect(rl2.Content).To(Equal("line2"))
	})
})

var _ = Describe("EscapedQuotedText", func() {

	It("should strip matching backslashes", func() {
		result, err := types.NewEscapedQuotedText(`\`, "*", &types.StringElement{Content: "bold"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(3))
		// backslash stripped, marker preserved
		Expect(result[0]).To(Equal(&types.StringElement{Content: "*"}))
		Expect(result[2]).To(Equal(&types.StringElement{Content: "*"}))
	})

	It("should keep extra backslashes", func() {
		result, err := types.NewEscapedQuotedText(`\\`, "*", &types.StringElement{Content: "bold"})
		Expect(err).NotTo(HaveOccurred())
		// 2 backslashes - 1 marker length = 1 extra backslash
		Expect(result[0]).To(Equal(&types.StringElement{Content: `\*`}))
	})
})

var _ = Describe("Footnotes", func() {

	It("should add a new footnote and return reference", func() {
		footnotes := types.NewFootnotes()
		note := &types.Footnote{
			Ref:      "fn1",
			Elements: []interface{}{&types.StringElement{Content: "footnote text"}},
		}
		ref := footnotes.Reference(note)
		Expect(ref.ID).To(Equal(1))
		Expect(ref.Ref).To(Equal("fn1"))
		Expect(ref.Duplicate).To(BeFalse())
		Expect(footnotes.Notes).To(HaveLen(1))
	})

	It("should return duplicate reference for same ref", func() {
		footnotes := types.NewFootnotes()
		note1 := &types.Footnote{
			Ref:      "fn1",
			Elements: []interface{}{&types.StringElement{Content: "text"}},
		}
		footnotes.Reference(note1)
		// Second reference with same ref but no elements (duplicate reference)
		note2 := &types.Footnote{Ref: "fn1"}
		ref := footnotes.Reference(note2)
		Expect(ref.ID).To(Equal(1))
		Expect(ref.Duplicate).To(BeTrue())
	})

	It("should return invalid reference for unknown ref without elements", func() {
		footnotes := types.NewFootnotes()
		note := &types.Footnote{Ref: "unknown"}
		ref := footnotes.Reference(note)
		Expect(ref.ID).To(Equal(types.InvalidFootnoteReference))
	})
})

var _ = Describe("RawLine", func() {

	It("should create from string", func() {
		rl, err := types.NewRawLine("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(rl.Content).To(Equal("hello"))
	})

	It("should append content", func() {
		rl, _ := types.NewRawLine("hello")
		rl.Append(" world")
		Expect(rl.Content).To(Equal("hello world"))
	})

	It("should check Contains", func() {
		rl, _ := types.NewRawLine("hello world")
		Expect(rl.Contains("world")).To(BeTrue())
		Expect(rl.Contains("foo")).To(BeFalse())
	})

	It("should check HasSuffix", func() {
		rl, _ := types.NewRawLine("hello world")
		Expect(rl.HasSuffix("world")).To(BeTrue())
		Expect(rl.HasSuffix("hello")).To(BeFalse())
	})
})
