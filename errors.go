package rvmat

import "errors"

var (
	// ErrBinaryRVMAT indicates the file is not a text RVMAT (likely binary surface data).
	ErrBinaryRVMAT = errors.New("binary rvmat")

	// ErrLex indicates a lexer failure.
	ErrLex = errors.New("lex error")

	// ErrParse indicates a parser failure.
	ErrParse = errors.New("parse error")
)
