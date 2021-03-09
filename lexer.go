package goparse

import (
	"fmt"
	"io"
	"strings"

	"github.com/bantling/goiter"
)

// lexical token type
type lexType uint

const (
	lexEOF = iota
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

// lexical table actions
const (
	lexSkip   uint = 0x1
	lexUnread uint = 0x2
	lexEOFOK  uint = 0x4
	lexDone   uint = 0x8
)

// the next table row to jump to and/or which actions to take
type lexActions struct {
	row     uint
	actions uint
	lexType lexType
}

// lexical errors
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

// panic with a LexError
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
	// lexical analyzer table, where each row is compressed into a map.
	// Since a rune is actually an int32, use -1 to refer to any other character.
	// If a row does not contain an entry for a given rune, and contains no -1 entry, it is a syntax error.
	lexTable = []map[rune]lexActions{
		// 0 - skip ws
		{
			'\t': {actions: lexSkip},
			'\r': {actions: lexSkip},
			'\n': {actions: lexSkip},
			' ':  {actions: lexSkip},
			'/':  {row: 1},
		},
		// 1
		{
			'/': {actions: lexEOFOK, row: 2},
			'*': {row: 3},
		},
		// 2 - comment-one-line
		{
			'\r': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			'\n': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			-1:   {actions: lexEOFOK, row: 2},
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

// lexical token to return to parser
type lexicalToken struct {
	lexType  lexType
	token    string
	line     int
	position int
}

// lexical analyzer
type lexer struct {
	iter *goiter.RunePositionIter
}

// construct lexer
func newLexer(source io.Reader) *lexer {
	return &lexer{
		iter: goiter.NewRunePositionIter(source),
	}
}

// read next lexical token
func (l *lexer) next() lexicalToken {
	var (
		token strings.Builder
		// line and position where token started
		line        int
		position    int
		row         = lexTable[0]
		lexActions  lexActions
		haveActions bool
		writeChar   bool
		eofOK       bool
		result      lexicalToken
	)

	for l.iter.Next() {
		nextChar := l.iter.Value()

		// get actions for char if they exist
		lexActions, haveActions = row[nextChar]
		if !haveActions {
			// get default actions, if they exist
			lexActions, haveActions = row[-1]
		}
		if !haveActions {
			// panic at current line and position, not where token started
			panicLexError(lexErrSyntax, l.iter.Line(), l.iter.Position())
		}

		writeChar = true

		// skipping chars only occurs before we recognize the first char of the next token
		// advance line and position if the token is empty, so it points at first char we care about
		if (lexActions.actions & lexSkip) > 0 {
			line = l.iter.Line()
			position = l.iter.Position()
			writeChar = false
		}

		// either the char is unread because it belongs to next token, or we write it as part of this token
		if (lexActions.actions & lexUnread) > 0 {
			l.iter.Unread(nextChar)
			writeChar = false
		}

		if writeChar {
			token.WriteRune(nextChar)
		}

		eofOK = (lexActions.actions & lexEOFOK) > 0

		// done means we've read final char of this token (or first char of next token in case of unread)
		if (lexActions.actions & lexDone) > 0 {
			result = lexicalToken{
				lexType:  lexActions.lexType,
				token:    token.String(),
				line:     line,
				position: position,
			}

			break
		}

		// jump to next row (which could be same row)
		row = lexTable[lexActions.row]
	}

	// cannot not encounter EOF in the middle of a token unless allowed
	if (token.Len() > 0) && (!eofOK) {
		panicLexError(lexErrEOF, l.iter.Line(), l.iter.Position())
	}

	// have a valid token
	return result
}
