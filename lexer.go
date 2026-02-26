// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/rvmat

package rvmat

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// tokenType represents a type of a token.
type tokenType int

// token types.
const (
	tokEOF       tokenType = iota // End of file
	tokIdent                      // Identifier
	tokNumber                     // Number
	tokString                     // String
	tokLBrace                     // Left brace
	tokRBrace                     // Right brace
	tokLBracket                   // Left bracket
	tokRBracket                   // Right bracket
	tokEqual                      // Equal
	tokSemicolon                  // Semicolon
	tokColon                      // Colon
	tokComma                      // Comma
	tokClass                      // Class
)

// token represents a token in the RVMAT file.
type token struct {
	Lit  string    // Literal value of the token
	Type tokenType // Type of the token
	Line int       // Line number of the token
	Col  int       // Column number of the token
}

// runeReader reads runes with one-rune pushback.
type runeReader interface {
	ReadRune() (rune, int, error)
	UnreadRune() error
}

// lexer represents a lexer for the RVMAT file.
type lexer struct {
	r   runeReader   // Reader for the input
	pos position     // Position of the current token
	ch  rune         // Current character
	opt ParseOptions // Options for the lexer
	eof bool         // End of file
}

// position represents a position in the input.
type position struct {
	line int // Line number
	col  int // Column number
}

// newLexer creates a new lexer for the RVMAT file.
func newLexer(r io.Reader, opt ParseOptions) *lexer {
	l := &lexer{r: toRuneReader(r), opt: opt, pos: position{line: 1, col: 0}}
	l.read()
	if l.ch == 0xFEFF {
		// Skip UTF-8 BOM if present.
		l.read()
	}

	return l
}

// toRuneReader reuses rune-capable readers and falls back to bufio.
func toRuneReader(r io.Reader) runeReader {
	if rr, ok := r.(runeReader); ok {
		return rr
	}

	return toBufferedReader(r)
}

// toBufferedReader reuses a buffered reader when available.
func toBufferedReader(r io.Reader) *bufio.Reader {
	if br, ok := r.(*bufio.Reader); ok {
		return br
	}

	return bufio.NewReader(r)
}

// next returns the next token from the RVMAT file.
func (l *lexer) next() (token, error) {
	// Tokenization is single-pass; skip whitespace/comments first.
	l.skipWhitespace()
	if l.eof {
		return token{Type: tokEOF, Line: l.pos.line, Col: l.pos.col}, nil
	}

	startLine, startCol := l.pos.line, l.pos.col

	// Tokenize the current character.
	switch l.ch {
	case '{':
		l.read()
		return token{Type: tokLBrace, Lit: "{", Line: startLine, Col: startCol}, nil
	case '}':
		l.read()
		return token{Type: tokRBrace, Lit: "}", Line: startLine, Col: startCol}, nil
	case '[':
		l.read()
		return token{Type: tokLBracket, Lit: "[", Line: startLine, Col: startCol}, nil
	case ']':
		l.read()
		return token{Type: tokRBracket, Lit: "]", Line: startLine, Col: startCol}, nil
	case '=':
		l.read()
		return token{Type: tokEqual, Lit: "=", Line: startLine, Col: startCol}, nil
	case ';':
		l.read()
		return token{Type: tokSemicolon, Lit: ";", Line: startLine, Col: startCol}, nil
	case ':':
		l.read()
		return token{Type: tokColon, Lit: ":", Line: startLine, Col: startCol}, nil
	case ',':
		l.read()
		return token{Type: tokComma, Lit: ",", Line: startLine, Col: startCol}, nil
	case '"':
		lit, err := l.readString()
		return token{Type: tokString, Lit: lit, Line: startLine, Col: startCol}, err

	default:
		if isIdentStart(l.ch) {
			lit := l.readIdent()
			if isClassKeyword(lit) {
				return token{Type: tokClass, Lit: lit, Line: startLine, Col: startCol}, nil
			}

			return token{Type: tokIdent, Lit: lit, Line: startLine, Col: startCol}, nil
		}

		if isNumberStart(l.ch) {
			// Some real-world files contain identifiers starting with digits (e.g. "1specular").
			// We read as a word, then decide whether it's a number or identifier.
			lit := l.readNumberOrIdent()
			if isValidNumber(lit) {
				return token{Type: tokNumber, Lit: lit, Line: startLine, Col: startCol}, nil
			}

			return token{Type: tokIdent, Lit: lit, Line: startLine, Col: startCol}, nil
		}

		return token{}, l.errorf("unexpected character '%c'", l.ch)
	}
}

// read reads the next character from the RVMAT file.
func (l *lexer) read() {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		l.eof = true
		l.ch = 0
		return
	}

	if ch == '\n' {
		l.pos.line++
		l.pos.col = 0
	} else {
		l.pos.col++
	}

	l.ch = ch
}

// peek returns the next character from the RVMAT file without consuming it.
func (l *lexer) peek() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return 0
	}

	_ = l.r.UnreadRune()
	return ch
}

// skipWhitespace skips whitespace characters.
func (l *lexer) skipWhitespace() {
	for {
		for unicode.IsSpace(l.ch) {
			l.read()
			if l.eof {
				return
			}
		}

		if !l.opt.DisableComments && l.ch == '/' {
			// Support // comments.
			next := l.peek()
			if next == '/' {
				l.read()
				l.read()
				for l.ch != '\n' && !l.eof {
					l.read()
				}
				continue
			}

			// Support /* */ comments.
			if next == '*' {
				l.read()
				l.read()
				for {
					if l.eof {
						return
					}
					if l.ch == '*' && l.peek() == '/' {
						l.read()
						l.read()
						break
					}
					l.read()
				}
				continue
			}
		}
		break
	}
}

// readIdent reads an identifier from the RVMAT file.
func (l *lexer) readIdent() string {
	b := make([]byte, 0, 16)
	for isIdentPart(l.ch) {
		b = appendRuneByteSlice(b, l.ch)
		l.read()
		if l.eof {
			break
		}
	}

	return string(b)
}

// readNumberOrIdent reads a number or identifier from the RVMAT file.
func (l *lexer) readNumberOrIdent() string {
	b := make([]byte, 0, 16)
	for isWordPart(l.ch) {
		b = appendRuneByteSlice(b, l.ch)
		l.read()
		if l.eof {
			break
		}
	}

	return string(b)
}

// readString reads a string from the RVMAT file.
func (l *lexer) readString() (string, error) {
	l.read() // consume opening quote
	b := make([]byte, 0, 16)
	for {
		if l.eof {
			return "", l.errorf("unterminated string")
		}

		// Handle quoted strings.
		if l.ch == '"' {
			if l.peek() == '"' {
				// Treat doubled quotes as an escaped quote (CSV-style).
				l.read()
				l.read()
				b = append(b, '"')
				continue
			}
			l.read()
			break
		}

		// Handle escaped characters.
		if l.ch == '\\' {
			next := l.peek()
			if next == '\\' || next == '"' {
				l.read()
				b = appendRuneByteSlice(b, l.ch)
				l.read()
				continue
			}
		}
		b = appendRuneByteSlice(b, l.ch)
		l.read()
	}

	return string(b), nil
}

// errorf formats an error message and returns an error.
func (l *lexer) errorf(format string, args ...any) error {
	return fmt.Errorf("%w at %d:%d: %s", ErrLex, l.pos.line, l.pos.col, fmt.Sprintf(format, args...))
}

// isIdentStart checks if a character is a valid start of an identifier.
func isIdentStart(r rune) bool {
	if isASCII(r) {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '$'
	}

	return unicode.IsLetter(r) || r == '_' || r == '$'
}

// isIdentPart checks if a character is a valid part of an identifier.
func isIdentPart(r rune) bool {
	if isASCII(r) {
		return (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' ||
			r == '$'
	}

	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

// isNumberStart checks if a character is a valid start of a number.
func isNumberStart(r rune) bool {
	if isASCII(r) {
		return (r >= '0' && r <= '9') || r == '-'
	}

	return unicode.IsDigit(r) || r == '-'
}

// isWordPart checks if a character is a valid part of a word.
func isWordPart(r rune) bool {
	return isIdentPart(r) || r == '.' || r == '+' || r == '-'
}

// isValidNumber checks if a string is a valid number.
func isValidNumber(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		b := s[i]
		if (b >= '0' && b <= '9') || b == '.' || b == '+' || b == '-' || b == 'e' || b == 'E' {
			continue
		}

		return false
	}

	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// appendRuneByteSlice appends a rune to a byte slice with an ASCII fast path.
func appendRuneByteSlice(dst []byte, r rune) []byte {
	if r >= 0 && r < utf8.RuneSelf {
		return append(dst, byte(r))
	}

	var tmp [utf8.UTFMax]byte
	n := utf8.EncodeRune(tmp[:], r)
	return append(dst, tmp[:n]...)
}

// isClassKeyword checks whether literal is "class" in ASCII case-insensitive form.
func isClassKeyword(s string) bool {
	return len(s) == 5 &&
		asciiLower(s[0]) == 'c' &&
		asciiLower(s[1]) == 'l' &&
		asciiLower(s[2]) == 'a' &&
		asciiLower(s[3]) == 's' &&
		asciiLower(s[4]) == 's'
}

// isASCII reports whether rune is within ASCII range.
func isASCII(r rune) bool {
	return r >= 0 && r < utf8.RuneSelf
}

// asciiLower lowercases ASCII letter bytes.
func asciiLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}

	return b
}
