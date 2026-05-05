package docx

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// resolveThemeVariables walks the YAML document tree and replaces any scalar
// containing $variable references or arithmetic expressions with its
// computed value, mirroring Asciidoctor PDF's theme variable interpolation.
//
// Theme keys are addressable as $<path>, where <path> is the parent keys
// joined to the leaf key with underscores: base.font_color → $base_font_color.
// Resolution iterates to a fixed point (a referenced value may itself contain
// $expr). Arithmetic supports + - * / with standard precedence and parentheses.
func resolveThemeVariables(root *yaml.Node) error {
	if root == nil || len(root.Content) == 0 {
		return nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil
	}

	const maxPasses = 16
	for pass := 0; pass < maxPasses; pass++ {
		vars := collectThemeScalars(doc, "")
		changed, err := resolveScalarsInNode(doc, vars)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
	}
	return fmt.Errorf("theme variable resolution did not converge after %d passes", maxPasses)
}

// collectThemeScalars walks the mapping tree and returns a map keyed by
// underscore-joined path → raw scalar value.
func collectThemeScalars(node *yaml.Node, prefix string) map[string]string {
	out := map[string]string{}
	if node.Kind != yaml.MappingNode {
		return out
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]
		path := keyNode.Value
		if prefix != "" {
			path = prefix + "_" + keyNode.Value
		}
		switch valNode.Kind {
		case yaml.ScalarNode:
			out[path] = valNode.Value
		case yaml.MappingNode:
			for k, v := range collectThemeScalars(valNode, path) {
				out[k] = v
			}
		}
	}
	return out
}

// resolveScalarsInNode visits every scalar node and replaces $expr values
// in-place. Returns true if any node changed.
func resolveScalarsInNode(node *yaml.Node, vars map[string]string) (bool, error) {
	changed := false
	var visit func(n *yaml.Node) error
	visit = func(n *yaml.Node) error {
		if n.Kind == yaml.ScalarNode {
			if !strings.Contains(n.Value, "$") {
				return nil
			}
			result, err := evalThemeExpr(n.Value, vars)
			if err != nil {
				return fmt.Errorf("line %d: %w", n.Line, err)
			}
			if result != n.Value {
				n.Value = result
				n.Tag = "" // let yaml re-detect type after substitution
				changed = true
			}
			return nil
		}
		for _, c := range n.Content {
			if err := visit(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(node); err != nil {
		return false, err
	}
	return changed, nil
}

// evalThemeExpr evaluates a theme value that may contain $variable references
// and arithmetic operators. If the input is a single $variable with no
// arithmetic, the variable's raw value is substituted as-is (which may be
// non-numeric, e.g. a colour). Otherwise all operands are coerced to numbers
// and the numeric result is returned.
func evalThemeExpr(s string, vars map[string]string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s, nil
	}

	if isBareVariable(trimmed) {
		v, ok := vars[trimmed[1:]]
		if !ok {
			return "", fmt.Errorf("undefined theme variable %s", trimmed)
		}
		return v, nil
	}

	toks, err := tokenizeExpr(trimmed)
	if err != nil {
		return "", err
	}
	p := &exprParser{toks: toks, vars: vars}
	f, err := p.parseExpr()
	if err != nil {
		return "", err
	}
	if p.pos != len(p.toks) {
		return "", fmt.Errorf("unexpected token %q at end of expression %q", p.toks[p.pos].value, s)
	}
	return strconv.FormatFloat(f, 'f', -1, 64), nil
}

func isBareVariable(s string) bool {
	if !strings.HasPrefix(s, "$") || len(s) < 2 {
		return false
	}
	for _, r := range s[1:] {
		if !isIdentRune(r) {
			return false
		}
	}
	return true
}

// --- Expression tokenizer + parser (operands are numbers, +-*/ with parens) ---

type exprTokenKind int

const (
	tokNum exprTokenKind = iota
	tokVar
	tokOp
	tokLParen
	tokRParen
)

type exprToken struct {
	kind  exprTokenKind
	value string
}

func tokenizeExpr(s string) ([]exprToken, error) {
	var toks []exprToken
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '+' || c == '-' || c == '*' || c == '/':
			toks = append(toks, exprToken{tokOp, string(c)})
			i++
		case c == '(':
			toks = append(toks, exprToken{tokLParen, "("})
			i++
		case c == ')':
			toks = append(toks, exprToken{tokRParen, ")"})
			i++
		case c == '$':
			j := i + 1
			for j < len(s) && isIdentRune(rune(s[j])) {
				j++
			}
			if j == i+1 {
				return nil, fmt.Errorf("empty variable name at position %d in %q", i, s)
			}
			toks = append(toks, exprToken{tokVar, s[i+1 : j]})
			i = j
		case isDigit(c) || c == '.':
			j := i
			for j < len(s) && (isDigit(s[j]) || s[j] == '.') {
				j++
			}
			toks = append(toks, exprToken{tokNum, s[i:j]})
			i = j
		default:
			return nil, fmt.Errorf("unexpected character %q at position %d in %q", c, i, s)
		}
	}
	return toks, nil
}

type exprParser struct {
	toks []exprToken
	pos  int
	vars map[string]string
}

func (p *exprParser) peek() (exprToken, bool) {
	if p.pos >= len(p.toks) {
		return exprToken{}, false
	}
	return p.toks[p.pos], true
}

func (p *exprParser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOp || (t.value != "+" && t.value != "-") {
			return left, nil
		}
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if t.value == "+" {
			left += right
		} else {
			left -= right
		}
	}
}

func (p *exprParser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOp || (t.value != "*" && t.value != "/") {
			return left, nil
		}
		p.pos++
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if t.value == "*" {
			left *= right
		} else {
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		}
	}
}

func (p *exprParser) parseFactor() (float64, error) {
	t, ok := p.peek()
	if !ok {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	switch t.kind {
	case tokNum:
		p.pos++
		f, err := strconv.ParseFloat(t.value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number %q: %w", t.value, err)
		}
		return f, nil
	case tokVar:
		p.pos++
		raw, found := p.vars[t.value]
		if !found {
			return 0, fmt.Errorf("undefined theme variable $%s", t.value)
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return 0, fmt.Errorf("variable $%s = %q is not numeric and cannot be used in arithmetic", t.value, raw)
		}
		return f, nil
	case tokLParen:
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		closing, ok := p.peek()
		if !ok || closing.kind != tokRParen {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return v, nil
	case tokOp:
		// Unary prefix +/-.
		if t.value == "-" || t.value == "+" {
			p.pos++
			v, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			if t.value == "-" {
				return -v, nil
			}
			return v, nil
		}
	}
	return 0, fmt.Errorf("unexpected token %q", t.value)
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
