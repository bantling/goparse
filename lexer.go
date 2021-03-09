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
	lexInvalid = iota
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
	lexUnread      uint = 0x1
	lexDone        uint = 0x2
	lexSyntaxError uint = 0x4
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
)

// LexError describes a lexical error
type LexError struct {
	err      string
	line     int
	position int
}

// panic with a LexError
func panicLexError(line, position int) {
	panic(
		LexError{
			err:      fmt.Sprintf(lexErrSyntax, line, position),
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
		// 0
		{
			'\t': {},
			'\r': {},
			'\n': {},
			' ':  {},
			'/':  {row: 1},
		},
		// 1
		{
			'/': {row: 2},
			'*': {row: 3},
		},
		// 2 - comment-one-line
		{
			'\r': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			'\n': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			-1:   {row: 2},
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
		row         map[rune]lexActions
		lexActions  lexActions
		haveActions bool
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
			panicLexError(l.iter.Line(), l.iter.Position())
		}

		// skipping chars only occurs before we recognize the first char of the next token
		// advance line and position if the token is empty, so it points at first char we care about
		if token.Len() == 0 {
			line = l.iter.Line()
			position = l.iter.Position()
		}

		// either the char is unread because it belongs to next token, or we write it as part of this token
		if (lexActions.actions & lexUnread) > 0 {
			l.iter.Unread(nextChar)
		} else {
			token.WriteRune(nextChar)
		}

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

	return result
}
