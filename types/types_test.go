package types_test

import (
	"errors"

	"github.com/lukewilliamboswell/libasciidoc/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DocumentFragment", func() {

	Describe("NewDocumentFragment", func() {

		It("should create a fragment with position and no elements", func() {
			p := types.Position{Start: 0, End: 10}
			f := types.NewDocumentFragment(p)
			Expect(f.Position.Start).To(Equal(0))
			Expect(f.Position.End).To(Equal(10))
			Expect(f.Error).To(BeNil())
			Expect(f.Elements).To(BeEmpty())
		})

		It("should create a fragment with multiple elements", func() {
			p := types.Position{Start: 5, End: 20}
			e1 := &types.StringElement{Content: "hello"}
			e2 := &types.StringElement{Content: "world"}
			f := types.NewDocumentFragment(p, e1, e2)
			Expect(f.Elements).To(HaveLen(2))
			Expect(f.Elements[0]).To(Equal(e1))
			Expect(f.Elements[1]).To(Equal(e2))
			Expect(f.Error).To(BeNil())
		})
	})

	Describe("NewErrorFragment", func() {

		It("should create a fragment carrying an error", func() {
			p := types.Position{Start: 0, End: 5}
			err := errors.New("parse error")
			f := types.NewErrorFragment(p, err)
			Expect(f.Position.Start).To(Equal(0))
			Expect(f.Position.End).To(Equal(5))
			Expect(f.Error).To(MatchError("parse error"))
			Expect(f.Elements).To(BeNil())
		})
	})
})

var _ = Describe("AttributeDeclaration", func() {

	Describe("NewAttributeDeclaration", func() {

		It("should create a declaration with name and string value", func() {
			decl, err := types.NewAttributeDeclaration("toc", "left", ":toc: left")
			Expect(err).NotTo(HaveOccurred())
			Expect(decl.Name).To(Equal("toc"))
			Expect(decl.Value).To(Equal("left"))
		})

		It("should trim spaces from string value", func() {
			decl, err := types.NewAttributeDeclaration("foo", "  bar  ", ":foo:   bar  ")
			Expect(err).NotTo(HaveOccurred())
			Expect(decl.Value).To(Equal("bar"))
		})

		It("should create a declaration with nil value", func() {
			decl, err := types.NewAttributeDeclaration("myattr", nil, ":myattr:")
			Expect(err).NotTo(HaveOccurred())
			Expect(decl.Name).To(Equal("myattr"))
			Expect(decl.Value).To(BeNil())
		})

		It("should expose RawText", func() {
			rawText := ":myattr: somevalue"
			decl, err := types.NewAttributeDeclaration("myattr", "somevalue", rawText)
			Expect(err).NotTo(HaveOccurred())
			Expect(decl.RawText()).To(Equal(rawText))
		})
	})
})

var _ = Describe("AttributeReset", func() {

	Describe("NewAttributeReset", func() {

		It("should create a reset with name", func() {
			reset, err := types.NewAttributeReset("toc", ":toc!:")
			Expect(err).NotTo(HaveOccurred())
			Expect(reset.Name).To(Equal("toc"))
		})

		It("should expose RawText", func() {
			rawText := ":myattr!:"
			reset, err := types.NewAttributeReset("myattr", rawText)
			Expect(err).NotTo(HaveOccurred())
			Expect(reset.RawText()).To(Equal(rawText))
		})
	})
})

var _ = Describe("List", func() {

	makeList := func(kind types.ListKind, elements ...types.ListElement) *types.List {
		return &types.List{
			Kind:     kind,
			Elements: elements,
		}
	}

	makeUnorderedElement := func(content string) *types.UnorderedListElement {
		p := &types.Paragraph{
			Elements: []interface{}{&types.StringElement{Content: content}},
		}
		prefix := types.UnorderedListElementPrefix{BulletStyle: types.OneAsterisk}
		elem, err := types.NewUnorderedListElement(prefix, nil, p)
		Expect(err).NotTo(HaveOccurred())
		return elem
	}

	Describe("GetAttributes", func() {

		It("should return nil when no attributes", func() {
			l := makeList(types.UnorderedListKind)
			Expect(l.GetAttributes()).To(BeNil())
		})

		It("should return set attributes", func() {
			l := makeList(types.UnorderedListKind)
			l.Attributes = types.Attributes{"id": "mylist"}
			Expect(l.GetAttributes()).To(HaveKeyWithValue("id", "mylist"))
		})
	})

	Describe("AddAttributes", func() {

		It("should add attributes to empty list", func() {
			l := makeList(types.UnorderedListKind)
			l.AddAttributes(types.Attributes{"id": "mylist"})
			Expect(l.GetAttributes()).To(HaveKeyWithValue("id", "mylist"))
		})

		It("should merge new attributes with existing", func() {
			l := makeList(types.UnorderedListKind)
			l.Attributes = types.Attributes{"existing": "val"}
			l.AddAttributes(types.Attributes{"new": "attr"})
			Expect(l.GetAttributes()).To(HaveKeyWithValue("existing", "val"))
			Expect(l.GetAttributes()).To(HaveKeyWithValue("new", "attr"))
		})
	})

	Describe("SetAttributes", func() {

		It("should replace attributes", func() {
			l := makeList(types.UnorderedListKind)
			l.Attributes = types.Attributes{"old": "gone"}
			l.SetAttributes(types.Attributes{"new": "kept"})
			Expect(l.GetAttributes()).To(HaveKeyWithValue("new", "kept"))
			Expect(l.GetAttributes()).NotTo(HaveKey("old"))
		})
	})

	Describe("GetElements", func() {

		It("should return elements as interface slice", func() {
			e := makeUnorderedElement("item")
			l := makeList(types.UnorderedListKind, e)
			elems := l.GetElements()
			Expect(elems).To(HaveLen(1))
		})

		It("should return empty slice for list with no elements", func() {
			l := makeList(types.UnorderedListKind)
			Expect(l.GetElements()).To(BeEmpty())
		})
	})

	Describe("SetElements", func() {

		It("should not error (not implemented)", func() {
			l := makeList(types.UnorderedListKind)
			err := l.SetElements([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("CanAddElement", func() {

		It("should accept list element of the same kind", func() {
			e1 := makeUnorderedElement("first")
			e2 := makeUnorderedElement("second")
			l := makeList(types.UnorderedListKind, e1)
			Expect(l.CanAddElement(e2)).To(BeTrue())
		})

		It("should reject element of different kind", func() {
			prefix := types.OrderedListElementPrefix{Style: types.Arabic}
			ordered, err := types.NewOrderedListElement(prefix, &types.Paragraph{
				Elements: []interface{}{&types.StringElement{Content: "ordered"}},
			})
			Expect(err).NotTo(HaveOccurred())
			e := makeUnorderedElement("unordered")
			l := makeList(types.UnorderedListKind, e)
			Expect(l.CanAddElement(ordered)).To(BeFalse())
		})

		It("should accept ListContinuation", func() {
			e := makeUnorderedElement("first")
			l := makeList(types.UnorderedListKind, e)
			cont := &types.ListContinuation{Offset: 0, Element: &types.BlankLine{}}
			Expect(l.CanAddElement(cont)).To(BeTrue())
		})

		It("should reject unrelated type", func() {
			e := makeUnorderedElement("first")
			l := makeList(types.UnorderedListKind, e)
			Expect(l.CanAddElement(&types.BlankLine{})).To(BeFalse())
		})
	})

	Describe("AddElement", func() {

		It("should add a matching list element", func() {
			e1 := makeUnorderedElement("first")
			l := makeList(types.UnorderedListKind, e1)
			e2 := makeUnorderedElement("second")
			err := l.AddElement(e2)
			Expect(err).NotTo(HaveOccurred())
			Expect(l.Elements).To(HaveLen(2))
		})

		It("should return error for non-matching element", func() {
			prefix := types.OrderedListElementPrefix{Style: types.Arabic}
			ordered, _ := types.NewOrderedListElement(prefix, &types.Paragraph{
				Elements: []interface{}{&types.StringElement{Content: "ordered"}},
			})
			e := makeUnorderedElement("first")
			l := makeList(types.UnorderedListKind, e)
			err := l.AddElement(ordered)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Reference", func() {

		It("should add to refs when id and title are set", func() {
			e := makeUnorderedElement("item")
			l := makeList(types.UnorderedListKind, e)
			l.Attributes = types.Attributes{
				types.AttrID:    "mylist",
				types.AttrTitle: "My List",
			}
			refs := types.ElementReferences{}
			l.Reference(refs)
			Expect(refs).To(HaveKeyWithValue("mylist", "My List"))
		})

		It("should not add to refs when no id", func() {
			e := makeUnorderedElement("item")
			l := makeList(types.UnorderedListKind, e)
			l.Attributes = types.Attributes{types.AttrTitle: "My List"}
			refs := types.ElementReferences{}
			l.Reference(refs)
			Expect(refs).To(BeEmpty())
		})
	})

	Describe("LastElement", func() {

		It("should return last list element", func() {
			e1 := makeUnorderedElement("first")
			e2 := makeUnorderedElement("second")
			l := makeList(types.UnorderedListKind, e1, e2)
			Expect(l.LastElement()).To(Equal(e2))
		})

		It("should return nil when empty", func() {
			l := makeList(types.UnorderedListKind)
			Expect(l.LastElement()).To(BeNil())
		})
	})
})

var _ = Describe("NewListElements", func() {

	It("should collect plain elements", func() {
		p := &types.Paragraph{
			Elements: []interface{}{&types.StringElement{Content: "hello"}},
		}
		le, err := types.NewListElements([]interface{}{p})
		Expect(err).NotTo(HaveOccurred())
		Expect(le.Elements).To(HaveLen(1))
	})

	It("should attach preceding attributes to following WithAttributes element", func() {
		attrs := types.Attributes{"id": "para1"}
		p := &types.Paragraph{
			Elements: []interface{}{&types.StringElement{Content: "hello"}},
		}
		le, err := types.NewListElements([]interface{}{attrs, p})
		Expect(err).NotTo(HaveOccurred())
		// The paragraph should have received the attributes
		Expect(le.Elements).To(HaveLen(1))
		Expect(p.GetAttributes()).To(HaveKeyWithValue("id", "para1"))
	})
})

var _ = Describe("CalloutListElement", func() {

	makeCalloutElement := func(content string) *types.CalloutListElement {
		return &types.CalloutListElement{
			Ref: 1,
			Elements: []interface{}{
				&types.Paragraph{
					Elements: []interface{}{&types.StringElement{Content: content}},
				},
			},
		}
	}

	Describe("GetAttributes", func() {

		It("should return attributes", func() {
			e := makeCalloutElement("text")
			e.Attributes = types.Attributes{"id": "callout1"}
			Expect(e.GetAttributes()).To(HaveKeyWithValue("id", "callout1"))
		})
	})

	Describe("AddAttributes", func() {

		It("should add attributes", func() {
			e := makeCalloutElement("text")
			e.AddAttributes(types.Attributes{"id": "c1"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("id", "c1"))
		})
	})

	Describe("SetAttributes", func() {

		It("should set attributes", func() {
			e := makeCalloutElement("text")
			e.Attributes = types.Attributes{"old": "gone"}
			e.SetAttributes(types.Attributes{"new": "kept"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("new", "kept"))
			Expect(e.GetAttributes()).NotTo(HaveKey("old"))
		})
	})

	Describe("GetElements", func() {

		It("should return elements", func() {
			e := makeCalloutElement("text")
			Expect(e.GetElements()).To(HaveLen(1))
		})
	})

	Describe("SetElements", func() {

		It("should set elements", func() {
			e := makeCalloutElement("text")
			newElems := []interface{}{&types.BlankLine{}}
			err := e.SetElements(newElems)
			Expect(err).NotTo(HaveOccurred())
			Expect(e.GetElements()).To(HaveLen(1))
		})
	})

	Describe("AddElement", func() {

		It("should add a raw line element to existing paragraph", func() {
			e := makeCalloutElement("existing")
			rl, err := types.NewRawLine("new line")
			Expect(err).NotTo(HaveOccurred())
			err = e.AddElement(rl)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("LastElement", func() {

		It("should return last element", func() {
			p := &types.Paragraph{
				Elements: []interface{}{&types.StringElement{Content: "text"}},
			}
			e := &types.CalloutListElement{
				Ref:      1,
				Elements: []interface{}{p},
			}
			Expect(e.LastElement()).To(Equal(p))
		})

		It("should return nil when empty", func() {
			e := &types.CalloutListElement{Ref: 1}
			Expect(e.LastElement()).To(BeNil())
		})
	})
})

var _ = Describe("OrderedListElement", func() {

	makeOrderedElement := func(content string) *types.OrderedListElement {
		prefix := types.OrderedListElementPrefix{Style: types.Arabic}
		p := &types.Paragraph{
			Elements: []interface{}{&types.StringElement{Content: content}},
		}
		elem, err := types.NewOrderedListElement(prefix, p)
		Expect(err).NotTo(HaveOccurred())
		return elem
	}

	Describe("AdjustStyle", func() {

		It("should not panic on AdjustStyle (no-op)", func() {
			e := makeOrderedElement("item")
			l := &types.List{Kind: types.OrderedListKind, Elements: []types.ListElement{e}}
			Expect(func() { e.AdjustStyle(l) }).NotTo(Panic())
		})
	})

	Describe("ListKind", func() {

		It("should return OrderedListKind", func() {
			e := makeOrderedElement("item")
			Expect(e.ListKind()).To(Equal(types.OrderedListKind))
		})
	})

	Describe("GetElements / SetElements", func() {

		It("should set and get elements", func() {
			e := makeOrderedElement("item")
			Expect(e.GetElements()).To(HaveLen(1))
			err := e.SetElements([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
			Expect(e.GetElements()).To(BeEmpty())
		})
	})

	Describe("LastElement", func() {

		It("should return last element", func() {
			e := makeOrderedElement("item")
			p := e.GetElements()[0]
			Expect(e.LastElement()).To(Equal(p))
		})

		It("should return nil when empty", func() {
			e := &types.OrderedListElement{}
			Expect(e.LastElement()).To(BeNil())
		})
	})

	Describe("AddElement", func() {

		It("should add a raw line to the paragraph", func() {
			e := makeOrderedElement("existing")
			rl, _ := types.NewRawLine("appended")
			err := e.AddElement(rl)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetAttributes / AddAttributes / SetAttributes", func() {

		It("should get, add and set attributes", func() {
			e := makeOrderedElement("item")
			Expect(e.GetAttributes()).To(BeNil())

			e.AddAttributes(types.Attributes{"id": "item1"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("id", "item1"))

			e.SetAttributes(types.Attributes{"new": "val"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("new", "val"))
			Expect(e.GetAttributes()).NotTo(HaveKey("id"))
		})
	})
})

var _ = Describe("LabeledListElement", func() {

	makeLabeled := func(term, content string) *types.LabeledListElement {
		p := &types.Paragraph{
			Elements: []interface{}{&types.StringElement{Content: content}},
		}
		elem, err := types.NewLabeledListElement(1, &types.StringElement{Content: term}, p)
		Expect(err).NotTo(HaveOccurred())
		return elem
	}

	Describe("AdjustStyle", func() {

		It("should not panic (no-op)", func() {
			e := makeLabeled("term", "desc")
			l := &types.List{Kind: types.LabeledListKind, Elements: []types.ListElement{e}}
			Expect(func() { e.AdjustStyle(l) }).NotTo(Panic())
		})
	})

	Describe("ListKind", func() {

		It("should return LabeledListKind", func() {
			e := makeLabeled("term", "desc")
			Expect(e.ListKind()).To(Equal(types.LabeledListKind))
		})
	})

	Describe("GetElements / SetElements", func() {

		It("should set and get elements", func() {
			e := makeLabeled("term", "desc")
			Expect(e.GetElements()).To(HaveLen(1))
			err := e.SetElements([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
			Expect(e.GetElements()).To(BeEmpty())
		})
	})

	Describe("LastElement", func() {

		It("should return last element", func() {
			e := makeLabeled("term", "desc")
			p := e.GetElements()[0]
			Expect(e.LastElement()).To(Equal(p))
		})

		It("should return nil when no elements", func() {
			e := &types.LabeledListElement{}
			Expect(e.LastElement()).To(BeNil())
		})
	})

	Describe("AddElement", func() {

		It("should add a raw line to the paragraph", func() {
			e := makeLabeled("term", "desc")
			rl, _ := types.NewRawLine("extra")
			err := e.AddElement(rl)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetAttributes / AddAttributes / SetAttributes", func() {

		It("should get, add, and set attributes", func() {
			e := makeLabeled("term", "desc")
			Expect(e.GetAttributes()).To(BeNil())

			e.AddAttributes(types.Attributes{"id": "label1"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("id", "label1"))

			e.SetAttributes(types.Attributes{"new": "attr"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("new", "attr"))
			Expect(e.GetAttributes()).NotTo(HaveKey("id"))
		})
	})
})

var _ = Describe("UnorderedListElement", func() {

	makeUnordered := func(content string) *types.UnorderedListElement {
		prefix := types.UnorderedListElementPrefix{BulletStyle: types.OneAsterisk}
		p := &types.Paragraph{
			Elements: []interface{}{&types.StringElement{Content: content}},
		}
		elem, err := types.NewUnorderedListElement(prefix, nil, p)
		Expect(err).NotTo(HaveOccurred())
		return elem
	}

	Describe("ListKind", func() {

		It("should return UnorderedListKind", func() {
			e := makeUnordered("item")
			Expect(e.ListKind()).To(Equal(types.UnorderedListKind))
		})
	})

	Describe("GetElements / SetElements", func() {

		It("should set and get elements", func() {
			e := makeUnordered("item")
			Expect(e.GetElements()).To(HaveLen(1))
			err := e.SetElements([]interface{}{})
			Expect(err).NotTo(HaveOccurred())
			Expect(e.GetElements()).To(BeEmpty())
		})
	})

	Describe("LastElement", func() {

		It("should return last element", func() {
			e := makeUnordered("item")
			p := e.GetElements()[0]
			Expect(e.LastElement()).To(Equal(p))
		})

		It("should return nil when empty", func() {
			e := &types.UnorderedListElement{}
			Expect(e.LastElement()).To(BeNil())
		})
	})

	Describe("AddElement", func() {

		It("should add a raw line to the paragraph", func() {
			e := makeUnordered("first")
			rl, _ := types.NewRawLine("second")
			err := e.AddElement(rl)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetAttributes / AddAttributes / SetAttributes", func() {

		It("should get, add, and set attributes", func() {
			e := makeUnordered("item")
			Expect(e.GetAttributes()).To(BeNil())

			e.AddAttributes(types.Attributes{"id": "ul1"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("id", "ul1"))

			e.SetAttributes(types.Attributes{"new": "val"})
			Expect(e.GetAttributes()).To(HaveKeyWithValue("new", "val"))
			Expect(e.GetAttributes()).NotTo(HaveKey("id"))
		})
	})

	Describe("Reference", func() {

		It("should add to refs when id and title set", func() {
			e := makeUnordered("item")
			e.Attributes = types.Attributes{
				types.AttrID:    "ul-id",
				types.AttrTitle: "UL Title",
			}
			refs := types.ElementReferences{}
			e.Reference(refs)
			Expect(refs).To(HaveKeyWithValue("ul-id", "UL Title"))
		})
	})
})

var _ = Describe("ImageBlock", func() {

	makeLocation := func(path string) *types.Location {
		return &types.Location{Path: path}
	}

	Describe("NewImageBlock", func() {

		It("should create an image block with location", func() {
			loc := makeLocation("images/photo.png")
			img, err := types.NewImageBlock(loc, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(img.GetLocation()).To(Equal(loc))
		})

		It("should map positional attributes to alt/width/height", func() {
			loc := makeLocation("images/photo.png")
			attrs := types.Attributes{
				types.AttrPositional1: "My Alt",
				types.AttrPositional2: "300",
				types.AttrPositional3: "200",
			}
			img, err := types.NewImageBlock(loc, attrs)
			Expect(err).NotTo(HaveOccurred())
			Expect(img.GetAttributes()).To(HaveKeyWithValue(types.AttrImageAlt, "My Alt"))
			Expect(img.GetAttributes()).To(HaveKeyWithValue(types.AttrWidth, "300"))
			Expect(img.GetAttributes()).To(HaveKeyWithValue(types.AttrHeight, "200"))
		})
	})

	Describe("SetAttributes", func() {

		It("should replace attributes", func() {
			loc := makeLocation("images/photo.png")
			img, _ := types.NewImageBlock(loc, nil)
			img.SetAttributes(types.Attributes{"id": "img1"})
			Expect(img.GetAttributes()).To(HaveKeyWithValue("id", "img1"))
		})
	})

	Describe("AddAttributes", func() {

		It("should add attributes", func() {
			loc := makeLocation("images/photo.png")
			img, _ := types.NewImageBlock(loc, nil)
			img.AddAttributes(types.Attributes{"id": "img1"})
			Expect(img.GetAttributes()).To(HaveKeyWithValue("id", "img1"))
		})
	})

	Describe("Reference", func() {

		It("should add to refs when id and title set", func() {
			loc := makeLocation("images/photo.png")
			img, _ := types.NewImageBlock(loc, nil)
			img.SetAttributes(types.Attributes{
				types.AttrID:    "img-id",
				types.AttrTitle: "Image Title",
			})
			refs := types.ElementReferences{}
			img.Reference(refs)
			Expect(refs).To(HaveKeyWithValue("img-id", "Image Title"))
		})

		It("should not add to refs when no id", func() {
			loc := makeLocation("images/photo.png")
			img, _ := types.NewImageBlock(loc, nil)
			img.SetAttributes(types.Attributes{types.AttrTitle: "Title Only"})
			refs := types.ElementReferences{}
			img.Reference(refs)
			Expect(refs).To(BeEmpty())
		})
	})
})

var _ = Describe("InlineImage", func() {

	makeLocation := func(path string) *types.Location {
		return &types.Location{Path: path}
	}

	Describe("NewInlineImage", func() {

		It("should create an inline image", func() {
			loc := makeLocation("icons/warning.png")
			img, err := types.NewInlineImage(loc, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(img.GetLocation()).To(Equal(loc))
		})

		It("should map positional attributes", func() {
			loc := makeLocation("icons/warning.png")
			attrs := types.Attributes{
				types.AttrPositional1: "Warning Icon",
			}
			img, err := types.NewInlineImage(loc, attrs)
			Expect(err).NotTo(HaveOccurred())
			Expect(img.GetAttributes()).To(HaveKeyWithValue(types.AttrImageAlt, "Warning Icon"))
		})
	})

	Describe("AddAttributes", func() {

		It("should add attributes", func() {
			loc := makeLocation("icons/warning.png")
			img, _ := types.NewInlineImage(loc, nil)
			img.AddAttributes(types.Attributes{"class": "icon"})
			Expect(img.GetAttributes()).To(HaveKeyWithValue("class", "icon"))
		})
	})

	Describe("SetAttributes", func() {

		It("should replace attributes", func() {
			loc := makeLocation("icons/warning.png")
			img, _ := types.NewInlineImage(loc, nil)
			img.SetAttributes(types.Attributes{"class": "icon"})
			Expect(img.GetAttributes()).To(HaveKeyWithValue("class", "icon"))
		})
	})
})

var _ = Describe("ExternalCrossReference", func() {

	Describe("AddAttributes / SetAttributes / GetAttributes", func() {

		It("should add and set attributes", func() {
			xref, err := types.NewExternalCrossReference(
				&types.Location{Path: "https://example.com"},
				nil,
			)
			Expect(err).NotTo(HaveOccurred())

			xref.AddAttributes(types.Attributes{"label": "Example"})
			Expect(xref.GetAttributes()).To(HaveKeyWithValue("label", "Example"))

			xref.SetAttributes(types.Attributes{"new": "val"})
			Expect(xref.GetAttributes()).NotTo(BeNil())
		})
	})
})
