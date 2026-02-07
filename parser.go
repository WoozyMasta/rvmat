package rvmat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Parse parses a RVMAT from bytes.
func Parse(data []byte, opt *ParseOptions) (*Material, error) {
	return Decode(bytes.NewReader(data), opt)
}

// Decode parses a RVMAT from reader.
func Decode(r io.Reader, opt *ParseOptions) (*Material, error) {
	popt := opt.normalize()
	br := bufio.NewReader(r)
	if isBinaryRVMAT(br) {
		return nil, ErrBinaryRVMAT
	}

	p := newParser(br, popt)
	return p.parseMaterial()
}

// DecodeFile parses a RVMAT from a file.
func DecodeFile(path string, opt *ParseOptions) (*Material, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b, opt)
}

// parser represents a parser for the RVMAT file.
type parser struct {
	l   *lexer       // Lexer for the RVMAT file
	buf token        // Buffered token
	has bool         // Has buffered token
	opt ParseOptions // Options for the parser
}

// newParser creates a new parser for the RVMAT file.
func newParser(r io.Reader, opt ParseOptions) *parser {
	return &parser{l: newLexer(r, opt), opt: opt}
}

// next returns the next token from the RVMAT file.
func (p *parser) next() (token, error) {
	if p.has {
		p.has = false
		return p.buf, nil
	}

	return p.l.next()
}

// peek returns the next token from the RVMAT file without consuming it.
func (p *parser) peek() (token, error) {
	if p.has {
		return p.buf, nil
	}

	tok, err := p.l.next()
	if err != nil {
		return tok, err
	}

	p.buf = tok
	p.has = true
	return tok, nil
}

// parseMaterial parses the material from the RVMAT file.
func (p *parser) parseMaterial() (*Material, error) {
	m := &Material{}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Type == tokEOF {
			break
		}

		if tok.Type == tokClass {
			// Top-level classes are either StageX/TexGenX or unknown blocks.
			if err := p.parseTopClass(m); err != nil {
				return nil, err
			}
			continue
		}

		// Parse top-level assignments.
		if err := p.parseTopAssign(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// parseTopClass parses a top-level class.
func (p *parser) parseTopClass(m *Material) error {
	if _, err := p.expect(tokClass); err != nil {
		return err
	}

	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return err
	}

	// Check if class has base
	base := ""
	if tok, _ := p.peek(); tok.Type == tokColon {
		_, _ = p.next()
		btok, err := p.expect(tokIdent)
		if err != nil {
			return err
		}

		base = btok.Lit
	}

	// Check if class is a stage or texture generator
	name := nameTok.Lit
	if isStageName(name, p.opt) && base == "" {
		st, err := p.parseStageBody(name)
		if err != nil {
			return err
		}

		m.Stages = append(m.Stages, st)
		return nil
	}

	// Check if class is a texture generator
	if isTexGenName(name, p.opt) {
		tg, err := p.parseTexGenBody(name, base)
		if err != nil {
			return err
		}

		m.TexGens = append(m.TexGens, tg)
		return nil
	}

	// Parse class body
	cn, err := p.parseClassBody(name, base)
	if err != nil {
		return err
	}

	m.extras = append(m.extras, cn)
	return nil
}

// parseClassBody parses the body of a class.
func (p *parser) parseClassBody(name, base string) (classNode, error) {
	// Expect left brace.
	if _, err := p.expect(tokLBrace); err != nil {
		return classNode{}, err
	}

	// Parse nodes in the body
	var body []node
	for {
		tok, err := p.peek()
		if err != nil {
			return classNode{}, err
		}

		// Check if reached end of class body
		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		// Parse a node in the body
		n, err := p.parseNode()
		if err != nil {
			return classNode{}, err
		}

		body = append(body, n)
	}

	if _, err := p.expect(tokSemicolon); err != nil {
		return classNode{}, err
	}

	return classNode{Name: name, Base: base, Body: body}, nil
}

// parseStageBody parses the body of a stage.
func (p *parser) parseStageBody(name string) (Stage, error) {
	// Expect left brace
	if _, err := p.expect(tokLBrace); err != nil {
		return Stage{}, err
	}

	// Parse stage body
	st := Stage{Name: name}
	for {
		tok, err := p.peek()
		if err != nil {
			return Stage{}, err
		}

		// Check if reached end of stage body
		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		// Check if class is a stage class
		if tok.Type == tokClass {
			if err := p.parseStageClass(&st); err != nil {
				return Stage{}, err
			}
			continue
		}

		if err := p.parseStageAssign(&st); err != nil {
			return Stage{}, err
		}
	}

	if _, err := p.expect(tokSemicolon); err != nil {
		return Stage{}, err
	}

	return st, nil
}

// parseTexGenBody parses the body of a texture generator.
func (p *parser) parseTexGenBody(name, base string) (TexGen, error) {
	// Expect left brace
	if _, err := p.expect(tokLBrace); err != nil {
		return TexGen{}, err
	}

	// Parse texture generator body
	tg := TexGen{Name: name, Base: base}
	for {
		tok, err := p.peek()
		if err != nil {
			return TexGen{}, err
		}

		// Check if reached end of texture generator body
		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		// Check if class is a texture generator class
		if tok.Type == tokClass {
			if err := p.parseTexGenClass(&tg); err != nil {
				return TexGen{}, err
			}
			continue
		}

		if err := p.parseTexGenAssign(&tg); err != nil {
			return TexGen{}, err
		}
	}

	if _, err := p.expect(tokSemicolon); err != nil {
		return TexGen{}, err
	}

	return tg, nil
}

// parseStageClass parses the body of a stage class.
func (p *parser) parseStageClass(st *Stage) error {
	if _, err := p.expect(tokClass); err != nil {
		return err
	}

	// Expect identifier
	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return err
	}

	// Check if base is empty
	base := ""
	if tok, _ := p.peek(); tok.Type == tokColon {
		_, _ = p.next()
		btok, err := p.expect(tokIdent)
		if err != nil {
			return err
		}

		base = btok.Lit
	}

	// Check if name is uvTransform and base is empty
	if equalFold(nameTok.Lit, "uvTransform", p.opt) && base == "" {
		uv, err := p.parseUVTransformBody()
		if err != nil {
			return err
		}

		st.UVTransform = uv
		return nil
	}

	// Parse class body
	cn, err := p.parseClassBody(nameTok.Lit, base)
	if err != nil {
		return err
	}

	st.extras = append(st.extras, cn)
	return nil
}

// parseTexGenClass parses the body of a texture generator class.
func (p *parser) parseTexGenClass(tg *TexGen) error {
	if _, err := p.expect(tokClass); err != nil {
		return err
	}

	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return err
	}

	// Check if base is empty
	base := ""
	if tok, _ := p.peek(); tok.Type == tokColon {
		_, _ = p.next()
		btok, err := p.expect(tokIdent)
		if err != nil {
			return err
		}

		base = btok.Lit
	}

	// Check if name is uvTransform and base is empty
	if equalFold(nameTok.Lit, "uvTransform", p.opt) && base == "" {
		uv, err := p.parseUVTransformBody()
		if err != nil {
			return err
		}

		tg.UVTransform = uv
		return nil
	}

	cn, err := p.parseClassBody(nameTok.Lit, base)
	if err != nil {
		return err
	}

	tg.extras = append(tg.extras, cn)
	return nil
}

// parseStageAssign parses a stage assign.
func (p *parser) parseStageAssign(st *Stage) error {
	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return err
	}

	// Check if array
	isArray := false
	if tok, _ := p.peek(); tok.Type == tokLBracket {
		_, _ = p.next()
		if _, err := p.expect(tokRBracket); err != nil {
			return err
		}

		isArray = true
	}

	// Expect equal
	if _, err := p.expect(tokEqual); err != nil {
		return err
	}

	if !isArray {
		switch {
		case matchKey(nameTok.Lit, "texture", !p.opt.DisableCaseInsensitive):
			s, err := p.parseStringValue()
			if err != nil {
				return err
			}
			st.Texture = ParseTextureRef(s)
			return p.expectSemicolon()

		case matchKey(nameTok.Lit, "uvsource", !p.opt.DisableCaseInsensitive):
			s, err := p.parseStringValue()
			if err != nil {
				return err
			}
			st.UVSource = s
			return p.expectSemicolon()

		case matchKey(nameTok.Lit, "texgen", !p.opt.DisableCaseInsensitive):
			s, err := p.parseStringOrNumberValue(!p.opt.DisableRelaxedNumbers)
			if err != nil {
				return err
			}
			st.TexGen = s
			return p.expectSemicolon()
		}
	}

	// Parse value
	val, err := p.parseValue()
	if err != nil {
		return err
	}
	if err := p.expectSemicolon(); err != nil {
		return err
	}

	st.extras = append(st.extras, assignNode{Name: nameTok.Lit, IsArray: isArray, Value: val})
	return nil
}

// parseTexGenAssign parses a texture generator assign.
func (p *parser) parseTexGenAssign(tg *TexGen) error {
	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return err
	}

	isArray := false
	if tok, _ := p.peek(); tok.Type == tokLBracket {
		_, _ = p.next()
		if _, err := p.expect(tokRBracket); err != nil {
			return err
		}

		isArray = true
	}

	// Expect equal
	if _, err := p.expect(tokEqual); err != nil {
		return err
	}

	if !isArray {
		switch {
		case matchKey(nameTok.Lit, "uvsource", !p.opt.DisableCaseInsensitive):
			s, err := p.parseStringValue()
			if err != nil {
				return err
			}

			tg.UVSource = s
			return p.expectSemicolon()
		}
	}

	val, err := p.parseValue()
	if err != nil {
		return err
	}
	if err := p.expectSemicolon(); err != nil {
		return err
	}

	tg.extras = append(tg.extras, assignNode{Name: nameTok.Lit, IsArray: isArray, Value: val})
	return nil
}

// parseTopAssign parses a top-level assign.
func (p *parser) parseTopAssign(m *Material) error {
	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return err
	}

	isArray := false
	if tok, _ := p.peek(); tok.Type == tokLBracket {
		_, _ = p.next()
		if _, err := p.expect(tokRBracket); err != nil {
			return err
		}

		isArray = true
	}

	if _, err := p.expect(tokEqual); err != nil {
		return err
	}

	if isArray {
		// Hot path: arrays for top-level color fields.
		switch {
		case matchKey(nameTok.Lit, "ambient", !p.opt.DisableCaseInsensitive),
			matchKey(nameTok.Lit, "diffuse", !p.opt.DisableCaseInsensitive),
			matchKey(nameTok.Lit, "forceddiffuse", !p.opt.DisableCaseInsensitive),
			matchKey(nameTok.Lit, "emmisive", !p.opt.DisableCaseInsensitive),
			matchKey(nameTok.Lit, "specular", !p.opt.DisableCaseInsensitive):
			vals, err := p.parseNumberArray()
			if err != nil {
				return err
			}

			switch {
			case matchKey(nameTok.Lit, "ambient", !p.opt.DisableCaseInsensitive):
				m.Ambient = vals
			case matchKey(nameTok.Lit, "diffuse", !p.opt.DisableCaseInsensitive):
				m.Diffuse = vals
			case matchKey(nameTok.Lit, "forceddiffuse", !p.opt.DisableCaseInsensitive):
				m.ForcedDiffuse = vals
			case matchKey(nameTok.Lit, "emmisive", !p.opt.DisableCaseInsensitive):
				m.Emmisive = vals
			case matchKey(nameTok.Lit, "specular", !p.opt.DisableCaseInsensitive):
				m.Specular = vals
			}

			return p.expectSemicolon()
		}
	}

	if !isArray {
		// Hot path: scalar fields at top-level.
		switch {
		case matchKey(nameTok.Lit, "specularpower", !p.opt.DisableCaseInsensitive):
			num, err := p.parseNumberValue()
			if err != nil {
				return err
			}
			m.SpecularPower = &num
			return p.expectSemicolon()

		case matchKey(nameTok.Lit, "pixelshaderid", !p.opt.DisableCaseInsensitive):
			s, err := p.parseStringValue()
			if err != nil {
				return err
			}
			m.PixelShaderID = s
			return p.expectSemicolon()

		case matchKey(nameTok.Lit, "vertexshaderid", !p.opt.DisableCaseInsensitive):
			s, err := p.parseStringValue()
			if err != nil {
				return err
			}
			m.VertexShaderID = s
			return p.expectSemicolon()
		}
	}

	// Parse value
	val, err := p.parseValue()
	if err != nil {
		return err
	}
	if err := p.expectSemicolon(); err != nil {
		return err
	}

	m.extras = append(m.extras, assignNode{Name: nameTok.Lit, IsArray: isArray, Value: val})
	return nil
}

// parseUVTransformBody parses the body of a uvTransform.
func (p *parser) parseUVTransformBody() (*UVTransform, error) {
	// Expect left brace
	if _, err := p.expect(tokLBrace); err != nil {
		return nil, err
	}

	uv := &UVTransform{}
	// uvTransform is a fixed set of numeric arrays; we parse directly.
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		nameTok, err := p.expect(tokIdent)
		if err != nil {
			return nil, err
		}

		// Check if array
		if tok, _ := p.peek(); tok.Type == tokLBracket {
			_, _ = p.next()
			if _, err := p.expect(tokRBracket); err != nil {
				return nil, err
			}
		}
		if _, err := p.expect(tokEqual); err != nil {
			return nil, err
		}

		// Parse number array
		vals, err := p.parseNumberArray()
		if err != nil {
			return nil, err
		}

		// Parse value
		switch {
		case matchKey(nameTok.Lit, "aside", !p.opt.DisableCaseInsensitive):
			uv.Aside = vals
		case matchKey(nameTok.Lit, "up", !p.opt.DisableCaseInsensitive):
			uv.Up = vals
		case matchKey(nameTok.Lit, "dir", !p.opt.DisableCaseInsensitive):
			uv.Dir = vals
		case matchKey(nameTok.Lit, "pos", !p.opt.DisableCaseInsensitive):
			uv.Pos = vals
		}

		if err := p.expectSemicolon(); err != nil {
			return nil, err
		}
	}

	if err := p.expectSemicolon(); err != nil {
		return nil, err
	}

	return uv, nil
}

// parseNode parses a node.
func (p *parser) parseNode() (node, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}

	if tok.Type == tokClass {
		return p.parseClass()
	}

	return p.parseAssign()
}

// parseClass parses a class.
func (p *parser) parseClass() (node, error) {
	if _, err := p.expect(tokClass); err != nil {
		return nil, err
	}

	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return nil, err
	}

	base := ""
	if tok, _ := p.peek(); tok.Type == tokColon {
		_, _ = p.next()
		btok, err := p.expect(tokIdent)
		if err != nil {
			return nil, err
		}
		base = btok.Lit
	}

	if _, err := p.expect(tokLBrace); err != nil {
		return nil, err
	}

	// Parse class body
	var body []node
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}

		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		n, err := p.parseNode()
		if err != nil {
			return nil, err
		}

		body = append(body, n)
	}

	if _, err := p.expect(tokSemicolon); err != nil {
		return nil, err
	}

	return classNode{Name: nameTok.Lit, Base: base, Body: body}, nil
}

// parseAssign parses an assign.
func (p *parser) parseAssign() (node, error) {
	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return nil, err
	}

	isArray := false
	if tok, _ := p.peek(); tok.Type == tokLBracket {
		_, _ = p.next()
		if _, err := p.expect(tokRBracket); err != nil {
			return nil, err
		}
		isArray = true
	}

	if _, err := p.expect(tokEqual); err != nil {
		return nil, err
	}

	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(tokSemicolon); err != nil {
		return nil, err
	}

	return assignNode{Name: nameTok.Lit, IsArray: isArray, Value: val}, nil
}

// parseValue parses a value.
func (p *parser) parseValue() (value, error) {
	tok, err := p.next()
	if err != nil {
		return value{}, err
	}

	// Parse value
	switch tok.Type {
	case tokNumber:
		f, err := strconv.ParseFloat(tok.Lit, 64)
		if err != nil {
			return value{}, p.errorf(tok, "invalid number")
		}
		return value{Kind: valueNumber, Num: f}, nil

	case tokString:
		return value{Kind: valueString, Str: tok.Lit}, nil

	case tokIdent:
		return value{Kind: valueIdent, Str: tok.Lit}, nil

	case tokLBrace:
		arr, err := p.parseArray()
		return value{Kind: valueArray, Array: arr}, err

	default:
		return value{}, p.errorf(tok, "unexpected token")
	}
}

// parseArray parses an array.
func (p *parser) parseArray() ([]value, error) {
	var arr []value
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}

		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
		tok, err = p.peek()
		if err != nil {
			return nil, err
		}

		if tok.Type == tokComma {
			_, _ = p.next()
			continue
		}

		// Check if reached end of array
		if tok.Type == tokRBrace {
			continue
		}

		return nil, p.errorf(tok, "expected ',' or '}' in array")
	}

	return arr, nil
}

// parseNumberArray parses a number array.
func (p *parser) parseNumberArray() ([]float64, error) {
	return p.parseNumberArrayWithRelax(!p.opt.DisableRelaxedNumbers)
}

// parseNumberArrayWithRelax parses a number array with relaxed parsing.
func (p *parser) parseNumberArrayWithRelax(relaxed bool) ([]float64, error) {
	if _, err := p.expect(tokLBrace); err != nil {
		return nil, err
	}

	// Fast path for numeric arrays used in colors and transforms.
	var arr []float64
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}

		if tok.Type == tokRBrace {
			_, _ = p.next()
			break
		}

		numTok, err := p.next()
		if err != nil {
			return nil, err
		}

		f, ok := parseNumberToken(numTok)
		if !ok {
			if !relaxed {
				return nil, p.errorf(numTok, "expected number")
			}
			f = 0
		}

		arr = append(arr, f)
		tok, err = p.peek()
		if err != nil {
			return nil, err
		}

		// If comma parse next number
		if tok.Type == tokComma {
			_, _ = p.next()
			continue
		}

		// Check if reached end of array
		if tok.Type == tokRBrace {
			continue
		}

		return nil, p.errorf(tok, "expected ',' or '}' in array")
	}

	return arr, nil
}

// parseNumberValue parses a number value.
func (p *parser) parseNumberValue() (float64, error) {
	tok, err := p.expect(tokNumber)
	if err != nil {
		return 0, err
	}

	if f, ok := parseNumberToken(tok); ok {
		return f, nil
	}

	return 0, p.errorf(tok, "invalid number")
}

// parseStringValue parses a string value.
func (p *parser) parseStringValue() (string, error) {
	tok, err := p.next()
	if err != nil {
		return "", err
	}

	switch tok.Type {
	case tokString, tokIdent:
		return tok.Lit, nil
	default:
		return "", p.errorf(tok, "expected string")
	}
}

// parseStringOrNumberValue parses a string or number value.
func (p *parser) parseStringOrNumberValue(relaxed bool) (string, error) {
	tok, err := p.next()
	if err != nil {
		return "", err
	}

	// Parse string or number value
	switch tok.Type {
	case tokString, tokIdent:
		return tok.Lit, nil
	case tokNumber:
		return tok.Lit, nil

	default:
		if relaxed {
			return tok.Lit, nil
		}
		return "", p.errorf(tok, "expected string")
	}
}

// expect expects a token.
func (p *parser) expect(tt tokenType) (token, error) {
	tok, err := p.next()
	if err != nil {
		return tok, err
	}

	if tok.Type != tt {
		return tok, p.errorf(tok, "expected %s", tokenName(tt))
	}

	return tok, nil
}

// expectSemicolon expects a semicolon.
func (p *parser) expectSemicolon() error {
	_, err := p.expect(tokSemicolon)
	return err
}

// errorf formats an error.
func (p *parser) errorf(tok token, format string, args ...any) error {
	return fmt.Errorf("%w at %d:%d: %s", ErrParse, tok.Line, tok.Col, fmt.Sprintf(format, args...))
}

// tokenName returns the name of a token.
func tokenName(tt tokenType) string {
	switch tt {
	case tokEOF:
		return "EOF"
	case tokIdent:
		return "identifier"
	case tokNumber:
		return "number"
	case tokString:
		return "string"
	case tokLBrace:
		return "{"
	case tokRBrace:
		return "}"
	case tokLBracket:
		return "["
	case tokRBracket:
		return "]"
	case tokEqual:
		return "="
	case tokSemicolon:
		return ";"
	case tokColon:
		return ":"
	case tokComma:
		return ","
	case tokClass:
		return "class"
	default:
		return "token"
	}
}

// isBinaryRVMAT checks if the RVMAT is binary.
func isBinaryRVMAT(r *bufio.Reader) bool {
	// Binary RVMATs contain zero bytes early; text files do not.
	peek, err := r.Peek(4096)
	if err != nil && len(peek) == 0 {
		return false
	}

	// Check if binary (rapP) RVMAT
	for _, b := range peek {
		if b == 0x00 {
			return true
		}
	}

	return false
}

// isStageName checks if the name is a stage name.
func isStageName(name string, opt ParseOptions) bool {
	return hasPrefixKey(name, "stage", !opt.DisableCaseInsensitive)
}

// isTexGenName checks if the name is a texture generator name.
func isTexGenName(name string, opt ParseOptions) bool {
	return hasPrefixKey(name, "texgen", !opt.DisableCaseInsensitive)
}

// equalFold checks if the two strings are equal.
func equalFold(a, b string, opt ParseOptions) bool {
	if !opt.DisableCaseInsensitive {
		return strings.EqualFold(a, b)
	}

	return a == b
}

// matchKey checks if the two strings are equal.
func matchKey(a, b string, ci bool) bool {
	// ASCII-only case folding to avoid allocations.
	if ci {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// hasPrefixKey checks if the string has a prefix.
func hasPrefixKey(s, prefix string, ci bool) bool {
	// ASCII-only prefix check with optional case-insensitivity.
	if !ci {
		return strings.HasPrefix(s, prefix)
	}

	if len(s) < len(prefix) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		if asciiLower(s[i]) != prefix[i] {
			return false
		}
	}

	return true
}

// asciiLower converts a byte to lowercase.
func asciiLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

// parseNumberToken parses a number token.
func parseNumberToken(tok token) (float64, bool) {
	switch tok.Type {
	case tokNumber:
		f, err := strconv.ParseFloat(tok.Lit, 64)
		return f, err == nil

	case tokString, tokIdent:
		s := strings.TrimSpace(tok.Lit)
		for len(s) > 0 && s[len(s)-1] == '.' {
			s = s[:len(s)-1]
		}
		if s == "" {
			return 0, false
		}

		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil

	default:
		return 0, false
	}
}
