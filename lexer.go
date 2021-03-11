package goparse

import (
	"fmt"
	"io"
	"strings"

	"github.com/bantling/goiter"
)

// Lexical token type
type lexType uint

const (
	lexInvalid lexType = iota
	lexEOF
	lexCommentOneLine
	lexCommentMultiLine
	lexString
	lexRange
	lexN
	lexM
	lexZeroOrOne
	lexZeroOrMore
	lexOneOrMore
	lexIdentifier
	lexJoin
)

// Lexical table actions
const (
	lexSkip    uint = 0x01
	lexAdvance uint = 0x02
	lexUnread  uint = 0x04
	lexDone    uint = 0x08
	lexEOFOK   uint = 0x10
	lexError   uint = 0x20
)

// The next table row to jump to and/or which actions to take
type lexActions struct {
	actions uint
	row     uint
	lexType lexType
	errCode string
}

// Lexical errors
const (
	lexErrPosition   = " at line %d position %d"
	lexErrSyntax     = "Syntax error"
	lexErrSyntaxCode = "-1"
	lexErrEOF        = "Invalid EOF"
	lexErrEOFCode    = "-2"
)

// LexError describes a lexical error
type LexError struct {
	err      string
	code     string
	line     int
	position int
}

// Panic with a LexError
func panicLexError(msg string, code string, line, position int) {
	panic(
		LexError{
			err:      fmt.Sprintf("%s%s", msg, fmt.Sprintf(lexErrPosition, line, position)),
			code:     code,
			line:     line,
			position: position,
		},
	)
}

// Error is error interface
func (l LexError) Error() string {
	return l.err
}

// Lexical token to return to parser
type lexicalToken struct {
	lexType  lexType
	token    string
	line     int
	position int
}

// Lexical analyzer
type lexer struct {
	iter *goiter.RunePositionIter
}

// Construct lexer
func newLexer(source io.Reader) *lexer {
	return &lexer{
		iter: goiter.NewRunePositionIter(source),
	}
}

// Read next lexical token
func (l *lexer) next() lexicalToken {
	var (
		nextChar rune
		token    strings.Builder
		// line and position where token started
		line     = 1
		position = 1
		row      = lexTable[0]
		// initial actions in case we read EOF on first call to iter.Next
		theLexActions = lexActions{actions: lexSkip | lexEOFOK, lexType: lexEOF}
		haveActions   bool
		eofOK         bool
		writeChar     bool
	)

	for {
		haveActions = false
		if l.iter.Next() {
			nextChar = l.iter.Value()

			// get actions for char if they exist
			theLexActions, haveActions = row[nextChar]
			if !haveActions {
				// get default actions, if they exist
				theLexActions, haveActions = row[-1]
			}
			if !haveActions {
				// panic at current line and position, not where token started
				panicLexError(lexErrSyntax, lexErrSyntaxCode, l.iter.Line(), l.iter.Position()-1)
			}
		} else {
			if eofOK = (theLexActions.actions & lexEOFOK) > 0; !eofOK {
				// panic at current line and position, not where token started
				panicLexError(lexErrEOF, lexErrEOFCode, l.iter.Line(), l.iter.Position()-1)
			}
			break
		}

		writeChar = true

		// A char to be skipped is a delimiter at the beginning or end of a token
		if (theLexActions.actions & lexSkip) > 0 {
			writeChar = false
		}

		// Advance the position, this character is not part of a token
		if (theLexActions.actions & lexAdvance) > 0 {
			line = l.iter.Line()
			position = l.iter.Position()
		}

		// either the char is unread because it belongs to next token, or we write it as part of this token
		if (theLexActions.actions & lexUnread) > 0 {
			l.iter.Unread(nextChar)
			writeChar = false
		}

		if writeChar {
			token.WriteRune(nextChar)
		}

		if (theLexActions.actions & lexError) > 0 {
			panicLexError(lexErrors[theLexActions.errCode], theLexActions.errCode, l.iter.Line(), l.iter.Position()-1)
		}

		if (theLexActions.actions & lexDone) > 0 {
			break
		}

		// jump to next row (which could be same row)
		row = lexTable[theLexActions.row]
	}

	// cannot not encounter EOF in the middle of a token unless allowed
	if (theLexActions.lexType == lexEOF) && (!eofOK) {
		panicLexError(lexErrEOF, lexErrEOFCode, l.iter.Line(), l.iter.Position())
	}

	// have a valid token
	return lexicalToken{
		lexType:  theLexActions.lexType,
		token:    token.String(),
		line:     line,
		position: position,
	}
}
