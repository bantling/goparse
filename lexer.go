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
	lexSkip   uint = 0x1
	lexUnread uint = 0x2
	lexEOFOK  uint = 0x4
	lexDone   uint = 0x8
)

// The next table row to jump to and/or which actions to take
type lexActions struct {
	row     uint
	actions uint
	lexType lexType
}

// Lexical errors
const (
	lexErrSyntax = "Syntax error at line %d character %d"
	lexErrEOF    = "Invalid EOF at line %d character %d"
)

// LexError describes a lexical error
type LexError struct {
	err      string
	line     int
	position int
}

// Panic with a LexError
func panicLexError(msg string, line, position int) {
	panic(
		LexError{
			err:      fmt.Sprintf(msg, line, position),
			line:     line,
			position: position,
		},
	)
}

// Error is error interface
func (l LexError) Error() string {
	return l.err
}

var (
	// Lexical analyzer table, where each row is compressed into a map.
	// Since a rune is actually an int32, use -1 to refer to any other character.
	// If a row does not contain an entry for a given rune, and contains no -1 entry, it is a syntax error.
	lexTable = []map[rune]lexActions{
		// 0 - skip ws
		{
			'\t': {actions: lexSkip | lexEOFOK, lexType: lexEOF},
			'\n': {actions: lexSkip | lexEOFOK, lexType: lexEOF}, // goiter.RunePositionIter coalesces all EOLs into \n
			' ':  {actions: lexSkip | lexEOFOK, lexType: lexEOF},
			'/':  {row: 1},
		},
		// 1
		{
			'/': {actions: lexEOFOK, row: 2, lexType: lexCommentOneLine},
			'*': {row: 3},
		},
		// 2 - comment-one-line
		{
			'\r': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			'\n': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			-1:   {actions: lexEOFOK, lexType: lexCommentOneLine, row: 2},
		},
		// 3 - comment-multi-line
		{
			'*': {row: 4},
			-1:  {row: 3},
		},
		// 4
		{
			'*': {row: 4},
			'/': {actions: lexDone, lexType: lexCommentMultiLine},
			-1:  {row: 3},
		},
	}
)

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
				panicLexError(lexErrSyntax, l.iter.Line(), l.iter.Position())
			}
		} else {
			if eofOK = (theLexActions.actions & lexEOFOK) > 0; !eofOK {
				panicLexError(lexErrEOF, l.iter.Line(), l.iter.Position())
			}
			break
		}

		writeChar = true

		// skipping chars only occurs before we recognize the first char of the next token
		// advance line and position if the token is empty, so it points at first char we care about
		if (theLexActions.actions & lexSkip) > 0 {
			line = l.iter.Line()
			position = l.iter.Position()
			writeChar = false
		}

		// either the char is unread because it belongs to next token, or we write it as part of this token
		if (theLexActions.actions & lexUnread) > 0 {
			l.iter.Unread(nextChar)
			writeChar = false
		}

		if (theLexActions.actions & lexDone) > 0 {
			break
		}

		if writeChar {
			token.WriteRune(nextChar)
		}

		// jump to next row (which could be same row)
		row = lexTable[theLexActions.row]
	}

	// cannot not encounter EOF in the middle of a token unless allowed
	if (theLexActions.lexType == lexEOF) && (!eofOK) {
		panicLexError(lexErrEOF, l.iter.Line(), l.iter.Position())
	}

	// have a valid token
	return lexicalToken{
		lexType:  theLexActions.lexType,
		token:    token.String(),
		line:     line,
		position: position,
	}
}
