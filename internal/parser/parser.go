package parser

import (
	"github.com/bantling/goparse/internal/lexer"
)

// Parser performs parsing
type Parser struct {
	lex lexer.Lexer
}
