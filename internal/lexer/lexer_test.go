package lexer

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipWhitespaceEOF(t *testing.T) {
	var (
		text   string
		reader io.Reader
		lexer  *Lexer
		eof    Token
	)

	text = " \t \r \n \r\n  \t\t\r\r\n\n\r\n\r\n"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	eof = lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.Token())
	assert.Equal(t, "", eof.String())
}

func TestComment(t *testing.T) {
	var (
		text   string
		reader io.Reader
		lexer  *Lexer
		token  Token
		eof    Token
	)

	// Single line comment
	text = " a \t one-liner"
	reader = strings.NewReader(fmt.Sprintf("//%s\n", text))
	lexer = NewLexer(reader)
	token = lexer.Next()
	assert.Equal(t, Comment, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())

	eof = lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.String())

	// Multiline on one line
	reader = strings.NewReader(fmt.Sprintf("/*%s*/", text))
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Comment, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())

	eof = lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.String())

	// Multiline across two lines
	text = " a two\nliner"
	reader = strings.NewReader(fmt.Sprintf("/*%s*/", text))
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Comment, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())

	eof = lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.Token())
	assert.Equal(t, "", eof.String())
}

func TestIdentifier(t *testing.T) {
	var (
		text   string
		reader io.Reader
		lexer  *Lexer
		token  Token
		eof    Token
	)

	text = "agr8_name"
	reader = strings.NewReader(fmt.Sprintf("%s ", text))
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Identifier, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())

	eof = lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.Token())
	assert.Equal(t, "", eof.String())

	// No space after identifier, die at EOF.
	// Die because an Identifier ends by reading a non-identifier char
	// and top of loop dies if an EOF is read unless it's first char of a token.
	func() {
		defer func() {
			assert.Equal(t, ErrUnexpectedEOF, recover())
		}()

		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must die at EOF after identifier")
	}()
}

func TestString(t *testing.T) {
	var (
		text   string
		quoted string
		reader io.Reader
		lexer  *Lexer
		token  Token
	)

	text = "single quoted"
	quoted = fmt.Sprintf("'%s'", text)
	reader = strings.NewReader(quoted)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, String, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, quoted, token.String())

	text = "single \\\\ \\t \\r \\n \\' \" quoted"
	quoted = fmt.Sprintf("'%s'", text)
	reader = strings.NewReader(quoted)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, String, token.Type())
	assert.Equal(t, "single \\ \t \r \n ' \" quoted", token.Token())
	assert.Equal(t, quoted, token.String())

	text = "double quoted"
	quoted = fmt.Sprintf("\"%s\"", text)
	reader = strings.NewReader(quoted)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, String, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, quoted, token.String())

	text = "double \\\\ \\t \\r \\n ' \\\" quoted"
	quoted = fmt.Sprintf("\"%s\"", text)
	reader = strings.NewReader(quoted)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, String, token.Type())
	assert.Equal(t, "double \\ \t \r \n ' \" quoted", token.Token())
	assert.Equal(t, quoted, token.String())

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidStringEscape, recover())
		}()

		text = "'\\]'"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid string escape error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidStringEscape, recover())
		}()

		text = "'\\x'"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid string escape error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidStringEscape, recover())
		}()

		text = "\"\\]\""
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid string escape error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidStringEscape, recover())
		}()

		text = "\"\\x\""
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid string escape error")
	}()
}

func TestCharacterRange(t *testing.T) {
	var (
		text   string
		reader io.Reader
		lexer  *Lexer
		token  Token
		//		eof    Token
	)

	charsMap := func(chars ...rune) map[rune]bool {
		result := map[rune]bool{}

		for _, char := range chars {
			result[char] = true
		}

		return result
	}

	text = "[A]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('A'), token.Range())

	text = "[AB]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('A', 'B'), token.Range())

	text = "[ABC]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('A', 'B', 'C'), token.Range())

	text = "[-]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-'), token.Range())

	text = "[-A]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', 'A'), token.Range())

	text = "[A-]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', 'A'), token.Range())

	text = "[A-C]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('A', 'B', 'C'), token.Range())

	text = "[-A-C]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', 'A', 'B', 'C'), token.Range())

	text = "[A-C-]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', 'A', 'B', 'C'), token.Range())

	text = "[-A-C-]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', 'A', 'B', 'C'), token.Range())

	text = "[A-CE-G]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('A', 'B', 'C', 'E', 'F', 'G'), token.Range())

	text = "[A-CZE-G]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('A', 'B', 'C', 'E', 'F', 'G', 'Z'), token.Range())

	text = "[[]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('['), token.Range())

	text = "[\\\\\\t\\r\\n\\]]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('\\', '\t', '\r', '\n', ']'), token.Range())

	text = "[-]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-'), token.Range())

	text = "[--]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-'), token.Range())

	text = "[---]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-'), token.Range())

	text = "[--0]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '.', '/', '0'), token.Range())

	text = "[---0]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '0'), token.Range())

	text = "[----0]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '0'), token.Range())

	text = "[---0-]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '0'), token.Range())

	text = "[---0-2]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '0', '1', '2'), token.Range())

	text = "[----0-2]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '0', '1', '2'), token.Range())

	text = "[-----0-2]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.False(t, token.InvertedRange())
	assert.Equal(t, charsMap('-', '.', '/', '0', '2'), token.Range())

	invertedCharsMap := func(chars ...rune) map[rune]bool {
		result := map[rune]bool{}

		for k, v := range uselessChars {
			result[k] = v
		}

		for _, char := range chars {
			result[char] = true
		}

		return result
	}

	text = "[^]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.True(t, token.InvertedRange())
	assert.Equal(t, invertedCharsMap(), token.Range())

	text = "[^A]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.True(t, token.InvertedRange())
	assert.Equal(t, invertedCharsMap('A'), token.Range())

	text = "[^-A]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.True(t, token.InvertedRange())
	assert.Equal(t, invertedCharsMap('-', 'A'), token.Range())

	text = "[^^]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.True(t, token.InvertedRange())
	assert.Equal(t, invertedCharsMap('^'), token.Range())

	text = "[^^-a]"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, CharacterRange, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	assert.True(t, token.InvertedRange())
	assert.Equal(t, invertedCharsMap('^', '_', '`', 'a'), token.Range())

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidCharacterRangeEscape, recover())
		}()

		text = "[\\']"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid character range escape error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidCharacterRangeEscape, recover())
		}()

		text = "[\\\"]"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid character range escape error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidCharacterRangeEscape, recover())
		}()

		text = "[\\x]"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid character range escape error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrCharacterRangeEmpty, recover())
		}()

		text = "[]"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with range empty error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrCharacterRangeOutOfOrder, recover())
		}()

		text = "[2-0]"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with range out of order error")
	}()

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidCharacterRangeEscape, recover())
		}()

		text = "[\\']"
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)
		token = lexer.Next()
		assert.Fail(t, "Must panic with invalid character range escape error")
	}()
}

func TestRepetition(t *testing.T) {
	var (
		text   string
		reader io.Reader
		lexer  *Lexer
		token  Token
		n      int
		m      int
	)

	text = "{2}"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, m)

	text = "{2,}"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 2, n)
	assert.Equal(t, -1, m)

	text = "{,5}"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 0, n)
	assert.Equal(t, 5, m)

	text = "{2,5}"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 2, n)
	assert.Equal(t, 5, m)

	text = "{0,}"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 0, n)
	assert.Equal(t, -1, m)

	text = "{0,1}"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 0, n)
	assert.Equal(t, 1, m)

	text = "?"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 0, n)
	assert.Equal(t, 1, m)

	text = "*"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 0, n)
	assert.Equal(t, -1, m)

	text = "+"
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)
	token = lexer.Next()

	assert.Equal(t, Repetition, token.Type())
	assert.Equal(t, text, token.Token())
	assert.Equal(t, text, token.String())
	n, m = token.Repetitions()
	assert.Equal(t, 1, n)
	assert.Equal(t, -1, m)

	panicChecker := func(badRepetition string) {
		defer func() {
			assert.Equal(t, ErrRepetitionForm, recover())
		}()

		reader = strings.NewReader(badRepetition)
		lexer = NewLexer(reader)
		lexer.Next()

		assert.Fail(t, "Must panic with ErrRepetitionForm")
	}

	for _, failCase := range []string{"{}", "{,}", "{0}", "{0,0}", "{1, 0}", "{2, 1}"} {
		panicChecker(failCase)
	}
}

func TestOptions(t *testing.T) {
	var (
		text   string
		reader io.Reader
		lexer  *Lexer
		token  Token
	)

	text = ":AST :EOL:INDENT :OUTDENT "
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)

	options := []string{":AST", ":EOL", ":INDENT", ":OUTDENT"}
	types := []LexType{OptionAST, OptionEOL, OptionIndent, OptionOutdent}
	for i, typ := range types {
		token = lexer.Next()
		assert.Equal(t, typ, token.Type())
		assert.Equal(t, options[i], token.Token())
		assert.Equal(t, options[i], token.String())
	}

	eof := lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.Token())
	assert.Equal(t, "", eof.String())

	func() {
		defer func() {
			assert.Equal(t, ErrInvalidOption, recover())
		}()

		text = ":NOSUCHOPT "
		reader = strings.NewReader(text)
		lexer = NewLexer(reader)

		lexer.Next()
		assert.Fail(t, "Must panic")
	}()
}

func TestSymbols(t *testing.T) {
	var (
		text    string
		symbols []string
		reader  io.Reader
		lexer   *Lexer
		token   Token
	)

	text = "^()|,===;"
	symbols = []string{"^", "(", ")", "|", ",", "==", "=", ";"}
	reader = strings.NewReader(text)
	lexer = NewLexer(reader)

	types := []LexType{Hat, OpenParens, CloseParens, Bar, Comma, DoubleEquals, Equals, SemiColon}
	for i, symbol := range symbols {
		token = lexer.Next()
		assert.Equal(t, types[i], token.Type())
		assert.Equal(t, symbol, token.Token())
		assert.Equal(t, symbol, token.String())
	}

	eof := lexer.Next()
	assert.Equal(t, EOF, eof.Type())
	assert.Equal(t, "", eof.Token())
	assert.Equal(t, "", eof.String())
}

func TestLineNumber(t *testing.T) {

}
