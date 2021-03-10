package goparse

import (
	// "fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipWhitespaceEOF(t *testing.T) {
	var (
		tests = []string{
			"",
			" \n \r\n  \t\t\r\r\n\n\r\n\r\n",
		}
		lines = []int{
			1,
			8,
		}
		reader io.Reader
		lexer  *lexer
		token  lexicalToken
	)

	for i, test := range tests {
		reader = strings.NewReader(test)
		lexer = newLexer(reader)
		token = lexer.next()
		assert.Equal(t, lexEOF, token.lexType)
		assert.Equal(t, "", token.token)
		assert.Equal(t, lines[i], token.line)
		assert.Equal(t, 1, token.position)
	}
}

func TestCommentOneLine(t *testing.T) {
	var (
		tests = []string{
			"//",
			" // yahdy //*/",
			"  // yahdy //*/\rhk",
		}
		results = []string{
			"//",
			"// yahdy //*/",
			"// yahdy //*/",
		}
		reader io.Reader
		lexer  *lexer
		token  lexicalToken
	)

	for i, test := range tests {
		reader = strings.NewReader(test)
		lexer = newLexer(reader)
		token = lexer.next()
		assert.Equal(t, lexCommentOneLine, token.lexType)
		assert.Equal(t, results[i], token.token)
		assert.Equal(t, 1, token.line)
		assert.Equal(t, strings.IndexRune(test, '/')+1, token.position)
	}
}
