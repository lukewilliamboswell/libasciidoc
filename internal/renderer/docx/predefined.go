package docx

import (
	"github.com/lukewilliamboswell/libasciidoc/types"
)

var predefinedAttributes = map[string]string{
	"sp":             " ",
	"blank":          "",
	"empty":          "",
	"nbsp":           "\u00a0",
	"zwsp":           "\u200b",
	"wj":             "\u2060",
	"apos":           "'",
	"quot":           "\"",
	"lsquo":          "\u2018",
	"rsquo":          "\u2019",
	"ldquo":          "\u201c",
	"rdquo":          "\u201d",
	"deg":            "\u00b0",
	"plus":           "+",
	"brvbar":         "\u00a6",
	"vbar":           "|",
	"amp":            "&",
	"lt":             "<",
	"gt":             ">",
	"startsb":        "[",
	"endsb":          "]",
	"caret":          "^",
	"asterisk":       "*",
	"tilde":          "~",
	"backslash":      `\`,
	"backtick":       "`",
	"two-colons":     "::",
	"two-semicolons": ";",
	"cpp":            "C++",
}

var specialCharacters = map[string]string{
	"&":  "&",
	"<":  "<",
	">":  ">",
	"\"": "\"",
}

var symbols = map[string]string{
	"(C)":  "\u00a9",
	"(TM)": "\u2122",
	"(R)":  "\u00ae",
	"...":  "\u2026",
	"->":   "\u2192",
	"<-":   "\u2190",
	"=>":   "\u21d2",
	"<=":   "\u21d0",
	"--":   "\u2014",
	" -- ": "\u2009\u2014\u2009", // thin space + em dash + thin space
	"'":    "\u2019",             // right single quotation mark (curly apostrophe)
	"'`":   "\u2018",             // left single quotation mark
	"`'":   "\u2019",             // right single quotation mark
	"\"`":  "\u201c",             // left double quotation mark
	"`\"":  "\u201d",             // right double quotation mark
}

func (r *docxRenderer) renderSpecialCharacter(para *paragraphBuilder, e *types.SpecialCharacter, style runStyle) error {
	text := e.Name
	if mapped, ok := specialCharacters[text]; ok {
		text = mapped
	}
	r.writeTextRun(para, text, style)
	return nil
}

func (r *docxRenderer) renderSymbol(para *paragraphBuilder, e *types.Symbol, style runStyle) error {
	text := e.Name
	if mapped, ok := symbols[text]; ok {
		text = mapped
	}
	r.writeTextRun(para, text, style)
	return nil
}

func (r *docxRenderer) renderPredefinedAttribute(para *paragraphBuilder, e *types.PredefinedAttribute, style runStyle) error {
	text := e.Name
	if mapped, ok := predefinedAttributes[text]; ok {
		text = mapped
	}
	r.writeTextRun(para, text, style)
	return nil
}
