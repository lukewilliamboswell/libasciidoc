package types_test

import (
	"github.com/lukewilliamboswell/libasciidoc/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("attributes", func() {

	Describe("NewAttributes", func() {

		It("should return nil for empty input", func() {
			result, err := types.NewAttributes()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle a single Attribute", func() {
			attr := &types.Attribute{Key: "foo", Value: "bar"}
			result, err := types.NewAttributes(attr)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKeyWithValue("foo", "bar"))
		})

		It("should handle a single PositionalAttribute", func() {
			pa, err := types.NewPositionalAttribute("source")
			Expect(err).NotTo(HaveOccurred())
			result, err := types.NewAttributes(pa)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKeyWithValue("@positional-1", "source"))
		})

		It("should handle multiple positional attributes", func() {
			pa1, err := types.NewPositionalAttribute("source")
			Expect(err).NotTo(HaveOccurred())
			pa2, err := types.NewPositionalAttribute("go")
			Expect(err).NotTo(HaveOccurred())
			result, err := types.NewAttributes(pa1, pa2)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKeyWithValue("@positional-1", "source"))
			Expect(result).To(HaveKeyWithValue("@positional-2", "go"))
		})

		It("should return error for unsupported type", func() {
			_, err := types.NewAttributes("not-an-attribute")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected type of attribute"))
		})
	})

	Describe("MergeAttributes", func() {

		It("should return nil for empty input", func() {
			result, err := types.MergeAttributes()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle a single Attribute", func() {
			attr := &types.Attribute{Key: "style", Value: "source"}
			result, err := types.MergeAttributes(attr)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKeyWithValue("style", "source"))
		})

		It("should handle an Attributes map", func() {
			attrs := types.Attributes{"key1": "val1", "key2": "val2"}
			result, err := types.MergeAttributes(attrs)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKeyWithValue("key1", "val1"))
			Expect(result).To(HaveKeyWithValue("key2", "val2"))
		})

		It("should merge mixed types", func() {
			attr := &types.Attribute{Key: "style", Value: "quote"}
			attrs := types.Attributes{"author": "John", "title": "My Quote"}
			result, err := types.MergeAttributes(attr, attrs)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKeyWithValue("style", "quote"))
			Expect(result).To(HaveKeyWithValue("author", "John"))
			Expect(result).To(HaveKeyWithValue("title", "My Quote"))
		})

		It("should return error for unsupported type", func() {
			_, err := types.MergeAttributes(42)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected type of attribute"))
		})
	})

	Describe("Set", func() {

		It("should set a regular key", func() {
			a := types.Attributes{}
			a.Set("foo", "bar")
			Expect(a).To(HaveKeyWithValue("foo", "bar"))
		})

		It("should aggregate role into roles", func() {
			a := types.Attributes{}
			a.Set("role", "primary")
			Expect(a).To(HaveKey("roles"))
			Expect(a["roles"]).To(Equal(types.Roles{"primary"}))
		})

		It("should append to existing roles", func() {
			a := types.Attributes{
				"roles": types.Roles{"primary"},
			}
			a.Set("role", "secondary")
			Expect(a["roles"]).To(Equal(types.Roles{"primary", "secondary"}))
		})

		It("should set roles from Roles value", func() {
			a := types.Attributes{}
			a.Set("roles", types.Roles{"a", "b"})
			Expect(a["roles"]).To(Equal(types.Roles{"a", "b"}))
		})

		It("should append roles from Roles value to existing", func() {
			a := types.Attributes{
				"roles": types.Roles{"a"},
			}
			a.Set("roles", types.Roles{"b", "c"})
			Expect(a["roles"]).To(Equal(types.Roles{"a", "b", "c"}))
		})

		It("should aggregate option into options", func() {
			a := types.Attributes{}
			a.Set("option", "header")
			Expect(a).To(HaveKey("options"))
			Expect(a["options"]).To(Equal(types.Options{"header"}))
		})

		It("should append to existing options", func() {
			a := types.Attributes{
				"options": types.Options{"header"},
			}
			a.Set("option", "footer")
			Expect(a["options"]).To(Equal(types.Options{"header", "footer"}))
		})

		It("should split comma-separated options string", func() {
			a := types.Attributes{}
			a.Set("options", "header,footer,autowidth")
			opts := a["options"].(types.Options)
			Expect(opts).To(HaveLen(3))
			Expect(opts[0]).To(Equal("header"))
			Expect(opts[1]).To(Equal("footer"))
			Expect(opts[2]).To(Equal("autowidth"))
		})

		It("should set Options value directly", func() {
			a := types.Attributes{}
			a.Set("options", types.Options{"header"})
			Expect(a["options"]).To(Equal(types.Options{"header"}))
		})

		It("should wrap non-Options non-string value in Options", func() {
			a := types.Attributes{}
			a.Set("options", 42)
			Expect(a["options"]).To(Equal(types.Options{42}))
		})

		It("should create map on nil receiver", func() {
			var a types.Attributes
			a = a.Set("foo", "bar")
			Expect(a).To(HaveKeyWithValue("foo", "bar"))
		})
	})

	Describe("Unset", func() {

		It("should remove a key", func() {
			a := types.Attributes{"foo": "bar", "baz": "qux"}
			a.Unset("foo")
			Expect(a).NotTo(HaveKey("foo"))
			Expect(a).To(HaveKey("baz"))
		})
	})

	Describe("Has", func() {

		It("should return true for existing key", func() {
			a := types.Attributes{"foo": "bar"}
			Expect(a.Has("foo")).To(BeTrue())
		})

		It("should return false for missing key", func() {
			a := types.Attributes{"foo": "bar"}
			Expect(a.Has("missing")).To(BeFalse())
		})
	})

	Describe("HasOption", func() {

		It("should find option in Options slice", func() {
			a := types.Attributes{
				"options": types.Options{"header", "footer"},
			}
			Expect(a.HasOption("header")).To(BeTrue())
			Expect(a.HasOption("footer")).To(BeTrue())
			Expect(a.HasOption("missing")).To(BeFalse())
		})

		It("should find option as direct attribute", func() {
			a := types.Attributes{
				"interactive": true,
			}
			Expect(a.HasOption("interactive")).To(BeTrue())
		})

		It("should return false when no options", func() {
			a := types.Attributes{}
			Expect(a.HasOption("anything")).To(BeFalse())
		})
	})

	Describe("GetAsString", func() {

		It("should return string value", func() {
			a := types.Attributes{"key": "value"}
			val, ok := a.GetAsString("key")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("value"))
		})

		It("should return false for non-string value", func() {
			a := types.Attributes{"key": 42}
			val, ok := a.GetAsString("key")
			Expect(ok).To(BeFalse())
			Expect(val).To(BeEmpty())
		})

		It("should return false for missing key", func() {
			a := types.Attributes{}
			val, ok := a.GetAsString("missing")
			Expect(ok).To(BeFalse())
			Expect(val).To(BeEmpty())
		})
	})

	Describe("GetAsIntWithDefault", func() {

		It("should return int value", func() {
			a := types.Attributes{"key": 42}
			Expect(a.GetAsIntWithDefault("key", 0)).To(Equal(42))
		})

		It("should parse string int value", func() {
			a := types.Attributes{"key": "42"}
			Expect(a.GetAsIntWithDefault("key", 0)).To(Equal(42))
		})

		It("should return default for non-int string", func() {
			a := types.Attributes{"key": "not-a-number"}
			Expect(a.GetAsIntWithDefault("key", 99)).To(Equal(99))
		})

		It("should return default for missing key", func() {
			a := types.Attributes{}
			Expect(a.GetAsIntWithDefault("missing", 5)).To(Equal(5))
		})
	})

	Describe("GetAsBoolWithDefault", func() {

		It("should return bool value", func() {
			a := types.Attributes{"key": true}
			Expect(a.GetAsBoolWithDefault("key", false)).To(BeTrue())
		})

		It("should parse string bool value", func() {
			a := types.Attributes{"key": "true"}
			Expect(a.GetAsBoolWithDefault("key", false)).To(BeTrue())
		})

		It("should return default for non-bool string", func() {
			a := types.Attributes{"key": "not-a-bool"}
			Expect(a.GetAsBoolWithDefault("key", true)).To(BeTrue())
		})

		It("should return default for missing key", func() {
			a := types.Attributes{}
			Expect(a.GetAsBoolWithDefault("missing", true)).To(BeTrue())
		})
	})

	Describe("GetAsStringWithDefault", func() {

		It("should return string value", func() {
			a := types.Attributes{"key": "hello"}
			Expect(a.GetAsStringWithDefault("key", "default")).To(Equal("hello"))
		})

		It("should return empty string for nil value", func() {
			a := types.Attributes{"key": nil}
			Expect(a.GetAsStringWithDefault("key", "default")).To(BeEmpty())
		})

		It("should format non-string value", func() {
			a := types.Attributes{"key": 42}
			Expect(a.GetAsStringWithDefault("key", "default")).To(Equal("42"))
		})

		It("should return default for missing key", func() {
			a := types.Attributes{}
			Expect(a.GetAsStringWithDefault("missing", "default")).To(Equal("default"))
		})
	})

	Describe("Clone", func() {

		It("should return independent copy", func() {
			a := types.Attributes{"key": "value"}
			b := a.Clone()
			b["key"] = "changed"
			Expect(a["key"]).To(Equal("value"))
			Expect(b["key"]).To(Equal("changed"))
		})
	})

	Describe("AddAll", func() {

		It("should add all from other attributes", func() {
			a := types.Attributes{"key1": "val1"}
			a = a.AddAll(types.Attributes{"key2": "val2"})
			Expect(a).To(HaveKeyWithValue("key1", "val1"))
			Expect(a).To(HaveKeyWithValue("key2", "val2"))
		})

		It("should return self when other is nil", func() {
			a := types.Attributes{"key1": "val1"}
			result := a.AddAll(nil)
			Expect(result).To(HaveKeyWithValue("key1", "val1"))
		})

		It("should create new map on nil receiver", func() {
			var a types.Attributes
			a = a.AddAll(types.Attributes{"key": "val"})
			Expect(a).To(HaveKeyWithValue("key", "val"))
		})
	})

	Describe("SetAll", func() {

		It("should set from Attribute pointer", func() {
			a := types.Attributes{}
			a = a.SetAll(&types.Attribute{Key: "foo", Value: "bar"})
			Expect(a).To(HaveKeyWithValue("foo", "bar"))
		})

		It("should set from Attributes map", func() {
			a := types.Attributes{}
			a = a.SetAll(types.Attributes{"k1": "v1", "k2": "v2"})
			Expect(a).To(HaveKeyWithValue("k1", "v1"))
			Expect(a).To(HaveKeyWithValue("k2", "v2"))
		})

		It("should set from slice of interfaces", func() {
			a := types.Attributes{}
			a = a.SetAll([]interface{}{
				&types.Attribute{Key: "a", Value: "1"},
				types.Attributes{"b": "2"},
			})
			Expect(a).To(HaveKeyWithValue("a", "1"))
			Expect(a).To(HaveKeyWithValue("b", "2"))
		})

		It("should return nil for empty Attributes input on nil receiver", func() {
			var a types.Attributes
			a = a.SetAll(types.Attributes{})
			Expect(a).To(BeNil())
		})

		It("should create new map on nil receiver when needed", func() {
			var a types.Attributes
			a = a.SetAll(&types.Attribute{Key: "foo", Value: "bar"})
			Expect(a).To(HaveKeyWithValue("foo", "bar"))
		})
	})

	Describe("HasAttributeWithValue", func() {

		It("should match Attribute pointer", func() {
			attr := &types.Attribute{Key: "style", Value: "source"}
			Expect(types.HasAttributeWithValue(attr, "style", "source")).To(BeTrue())
			Expect(types.HasAttributeWithValue(attr, "style", "quote")).To(BeFalse())
		})

		It("should match Attributes map", func() {
			attrs := types.Attributes{"style": "source"}
			Expect(types.HasAttributeWithValue(attrs, "style", "source")).To(BeTrue())
			Expect(types.HasAttributeWithValue(attrs, "style", "quote")).To(BeFalse())
		})

		It("should return false for unsupported type", func() {
			Expect(types.HasAttributeWithValue("not-attrs", "key", "val")).To(BeFalse())
		})
	})

	Describe("HasNotAttribute", func() {

		It("should return true when key is absent", func() {
			attrs := types.Attributes{"foo": "bar"}
			Expect(types.HasNotAttribute(attrs, "missing")).To(BeTrue())
		})

		It("should return false when key is present", func() {
			attrs := types.Attributes{"foo": "bar"}
			Expect(types.HasNotAttribute(attrs, "foo")).To(BeFalse())
		})
	})

	Describe("NewPositionalAttribute", func() {

		It("should create a positional attribute", func() {
			pa, err := types.NewPositionalAttribute("source")
			Expect(err).NotTo(HaveOccurred())
			Expect(pa.Value).To(Equal("source"))
		})

		It("should trim spaces from value", func() {
			pa, err := types.NewPositionalAttribute("  source  ")
			Expect(err).NotTo(HaveOccurred())
			Expect(pa.Value).To(Equal("source"))
		})
	})

	Describe("NewNamedAttribute", func() {

		It("should create a named attribute", func() {
			attr, err := types.NewNamedAttribute("style", "source")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.Key).To(Equal("style"))
			Expect(attr.Value).To(Equal("source"))
		})

		It("should trim key spaces", func() {
			attr, err := types.NewNamedAttribute("  style  ", "source")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.Key).To(Equal("style"))
		})

		It("should alias opts to options", func() {
			attr, err := types.NewNamedAttribute("opts", "header")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.Key).To(Equal("options"))
		})
	})

	Describe("NewOptionAttribute", func() {

		It("should create an option attribute", func() {
			attr, err := types.NewOptionAttribute("header")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.Key).To(Equal("option"))
			Expect(attr.Value).To(Equal("header"))
		})
	})
})
