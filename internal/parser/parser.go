package parser

import (
	"github.com/bantling/goparse/internal/lexer"
)

// Parser is the grammar parser that converts source text into a Grammar
type Parser struct {
	lex lexer.Lexer
}
