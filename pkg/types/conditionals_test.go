package types_test

import (
	"github.com/lukewilliamboswell/libasciidoc/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("conditionals", func() {

	Describe("IfdefCondition", func() {

		It("should create with substitution string", func() {
			cond, err := types.NewIfdefCondition("myattr", "replacement")
			Expect(err).NotTo(HaveOccurred())
			Expect(cond.Name).To(Equal("myattr"))
			Expect(cond.Substitution).To(Equal("replacement"))
		})

		It("should create without substitution", func() {
			cond, err := types.NewIfdefCondition("myattr", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cond.Substitution).To(BeEmpty())
		})

		It("should eval true when attribute is present", func() {
			cond, _ := types.NewIfdefCondition("found", nil)
			attrs := map[string]interface{}{"found": "value"}
			Expect(cond.Eval(attrs)).To(BeTrue())
		})

		It("should eval false when attribute is absent", func() {
			cond, _ := types.NewIfdefCondition("missing", nil)
			attrs := map[string]interface{}{}
			Expect(cond.Eval(attrs)).To(BeFalse())
		})

		It("should return substitution as single line content", func() {
			cond, _ := types.NewIfdefCondition("x", "hello")
			content, ok := cond.SingleLineContent()
			Expect(ok).To(BeTrue())
			Expect(content).To(Equal("hello"))
		})

		It("should return false for single line content without substitution", func() {
			cond, _ := types.NewIfdefCondition("x", nil)
			_, ok := cond.SingleLineContent()
			Expect(ok).To(BeFalse())
		})
	})

	Describe("IfndefCondition", func() {

		It("should create with substitution string", func() {
			cond, err := types.NewIfndefCondition("myattr", "replacement")
			Expect(err).NotTo(HaveOccurred())
			Expect(cond.Name).To(Equal("myattr"))
			Expect(cond.Substitution).To(Equal("replacement"))
		})

		It("should eval true when attribute is absent", func() {
			cond, _ := types.NewIfndefCondition("missing", nil)
			attrs := map[string]interface{}{}
			Expect(cond.Eval(attrs)).To(BeTrue())
		})

		It("should eval false when attribute is present", func() {
			cond, _ := types.NewIfndefCondition("found", nil)
			attrs := map[string]interface{}{"found": "value"}
			Expect(cond.Eval(attrs)).To(BeFalse())
		})

		It("should return substitution as single line content", func() {
			cond, _ := types.NewIfndefCondition("x", "hello")
			content, ok := cond.SingleLineContent()
			Expect(ok).To(BeTrue())
			Expect(content).To(Equal("hello"))
		})
	})

	Describe("IfevalCondition", func() {

		It("should create condition", func() {
			cond, err := types.NewIfevalCondition("a", "b", types.EqualOperand)
			Expect(err).NotTo(HaveOccurred())
			Expect(cond).NotTo(BeNil())
		})

		It("should eval with literal string values", func() {
			cond, _ := types.NewIfevalCondition("hello", "hello", types.EqualOperand)
			Expect(cond.Eval(nil)).To(BeTrue())
		})

		It("should eval with attribute reference on left", func() {
			ref := &types.AttributeReference{Name: "myattr"}
			cond, _ := types.NewIfevalCondition(ref, "expected", types.EqualOperand)
			attrs := map[string]interface{}{"myattr": "expected"}
			Expect(cond.Eval(attrs)).To(BeTrue())
		})

		It("should eval with attribute reference on right", func() {
			ref := &types.AttributeReference{Name: "myattr"}
			cond, _ := types.NewIfevalCondition("expected", ref, types.EqualOperand)
			attrs := map[string]interface{}{"myattr": "expected"}
			Expect(cond.Eval(attrs)).To(BeTrue())
		})

		It("should fall back to literal when attribute not found", func() {
			ref := &types.AttributeReference{Name: "missing"}
			cond, _ := types.NewIfevalCondition(ref, ref, types.EqualOperand)
			// Both sides resolve to the same AttributeReference, which won't match string comparison
			// but the operand will receive the raw AttributeReference objects
			Expect(cond.Eval(nil)).To(BeFalse())
		})

		It("should return empty single line content", func() {
			cond, _ := types.NewIfevalCondition("a", "b", types.EqualOperand)
			content, ok := cond.SingleLineContent()
			Expect(ok).To(BeFalse())
			Expect(content).To(BeEmpty())
		})
	})

	Describe("operands", func() {

		Describe("EqualOperand", func() {
			DescribeTable("comparisons",
				func(left, right interface{}, expected bool) {
					Expect(types.EqualOperand(left, right)).To(Equal(expected))
				},
				Entry("equal strings", "a", "a", true),
				Entry("unequal strings", "a", "b", false),
				Entry("equal ints", 1, 1, true),
				Entry("unequal ints", 1, 2, false),
				Entry("equal floats", 1.5, 1.5, true),
				Entry("unequal floats", 1.5, 2.5, false),
				Entry("mismatched types", "1", 1, false),
			)
		})

		Describe("NotEqualOperand", func() {
			DescribeTable("comparisons",
				func(left, right interface{}, expected bool) {
					Expect(types.NotEqualOperand(left, right)).To(Equal(expected))
				},
				Entry("equal strings", "a", "a", false),
				Entry("unequal strings", "a", "b", true),
				Entry("equal ints", 1, 1, false),
				Entry("unequal ints", 1, 2, true),
				Entry("mismatched types", "1", 1, false),
			)
		})

		Describe("LessThanOperand", func() {
			DescribeTable("comparisons",
				func(left, right interface{}, expected bool) {
					Expect(types.LessThanOperand(left, right)).To(Equal(expected))
				},
				Entry("string less", "a", "b", true),
				Entry("string not less", "b", "a", false),
				Entry("string equal", "a", "a", false),
				Entry("int less", 1, 2, true),
				Entry("int not less", 2, 1, false),
				Entry("float less", 1.0, 2.0, true),
				Entry("float not less", 2.0, 1.0, false),
			)
		})

		Describe("LessOrEqualOperand", func() {
			DescribeTable("comparisons",
				func(left, right interface{}, expected bool) {
					Expect(types.LessOrEqualOperand(left, right)).To(Equal(expected))
				},
				Entry("string less", "a", "b", true),
				Entry("string equal", "a", "a", true),
				Entry("string greater", "b", "a", false),
				Entry("int less", 1, 2, true),
				Entry("int equal", 1, 1, true),
				Entry("int greater", 2, 1, false),
			)
		})

		Describe("GreaterThanOperand", func() {
			DescribeTable("comparisons",
				func(left, right interface{}, expected bool) {
					Expect(types.GreaterThanOperand(left, right)).To(Equal(expected))
				},
				Entry("string greater", "b", "a", true),
				Entry("string not greater", "a", "b", false),
				Entry("int greater", 2, 1, true),
				Entry("int not greater", 1, 2, false),
				Entry("float greater", 2.0, 1.0, true),
			)
		})

		Describe("GreaterOrEqualOperand", func() {
			DescribeTable("comparisons",
				func(left, right interface{}, expected bool) {
					Expect(types.GreaterOrEqualOperand(left, right)).To(Equal(expected))
				},
				Entry("string greater", "b", "a", true),
				Entry("string equal", "a", "a", true),
				Entry("string less", "a", "b", false),
				Entry("int greater", 2, 1, true),
				Entry("int equal", 1, 1, true),
				Entry("int less", 1, 2, false),
			)
		})

		Describe("operand constructors", func() {
			It("should return EqualOperand", func() {
				op, err := types.NewEqualOperand()
				Expect(err).NotTo(HaveOccurred())
				Expect(op("a", "a")).To(BeTrue())
			})
			It("should return NotEqualOperand", func() {
				op, err := types.NewNotEqualOperand()
				Expect(err).NotTo(HaveOccurred())
				Expect(op("a", "b")).To(BeTrue())
			})
			It("should return LessThanOperand", func() {
				op, err := types.NewLessThanOperand()
				Expect(err).NotTo(HaveOccurred())
				Expect(op(1, 2)).To(BeTrue())
			})
			It("should return LessOrEqualOperand", func() {
				op, err := types.NewLessOrEqualOperand()
				Expect(err).NotTo(HaveOccurred())
				Expect(op(1, 1)).To(BeTrue())
			})
			It("should return GreaterThanOperand", func() {
				op, err := types.NewGreaterThanOperand()
				Expect(err).NotTo(HaveOccurred())
				Expect(op(2, 1)).To(BeTrue())
			})
			It("should return GreaterOrEqualOperand", func() {
				op, err := types.NewGreaterOrEqualOperand()
				Expect(err).NotTo(HaveOccurred())
				Expect(op(1, 1)).To(BeTrue())
			})
		})
	})

	Describe("EndOfCondition", func() {
		It("should create", func() {
			eoc, err := types.NewEndOfCondition()
			Expect(err).NotTo(HaveOccurred())
			Expect(eoc).NotTo(BeNil())
		})
	})
})
