package parser

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bantling/goparse/internal/lexer"
)

// map of lex options to strings
var (
	lexTypeStrings = map[lexer.LexType]string{
		lexer.OptionAST:     ":AST",
		lexer.OptionEOL:     ":EOL",
		lexer.OptionIndent:  ":INDENT",
		lexer.OptionOutdent: ":OUTDENT",
	}
)

// ====

// Terminal is a string or character range.
// If the string is "", then the terminal is a character range, else it is a string.
type Terminal struct {
	theString string
	theRange  map[rune]bool
}

// OfTerminalString constructs a Terminal from a string
func OfTerminalString(theString string) Terminal {
	return Terminal{
		theString: theString,
	}
}

// OfTerminalRange constructs a Terminal from a range
func OfTerminalString(theRange map[rune]bool) Terminal {
	return Terminal{
		theRange: theRange,
	}
}

// IsString returns true of the terminal is a string
func (t Terminal) IsString() bool {
	return len(t.theString) > 0
}

// IsRange returns true of the terminal is a character range
func (t Terminal) IsRange() bool {
	return len(t.theRange) > 0
}

// TerminalString is the terminal string
func (t Terminal) TerminalString() string {
	return t.theString
}

// TerminalRange is the terminal range
func (t Terminal) TerminalRange() map[rune]bool {
	return t.theRange
}

// String returns a formatted string or range for the Terminal.
// A string is enclosed in double quotes and escapes control characters and backslashes.
// A range is enclosed in square brackets, with each maximum length contiguous set of characters provided as a range.
func (t Terminal) String() string {
	var result string
	
	if t.IsString() {
		result = fmt.Sprintf("%q", t.theString)
	} else {
		// Converting range map into a sorting slice of runes
		var allChars []rune
		
		for aChar, _ := range t.theRange {
			allChars = append(allChars, aChar)
		}
		
		sort.Slice(allChars, func(i, j int) bool {return allChars[i] < allChars[j]})
		
		// Scan sorted slice for contiguous ranges, building a slice of one or more such ranges
		var allRanges []map[rune]bool
	}
	
	return result 
}

// ====

// ListItem is a rule name or a terminal, and possibly some options.
// If the rule name is "", then the item is a terminal, else it is a rule name.
// Options can be applied to a rule name or a terminal.
type ListItem struct {
	ruleName string
	terminal Terminal
	options  []lexer.LexType
}

// RuleName is the rule name
func (itm ListItem) RuleName() string {
	return itm.ruleName
}

// Terminal is the terminal
func (itm ListItem) Terminal() Terminal {
	return itm.terminal
}

// String returns a formatted string, using rule name or terminal string, followed by any options.
// If the result is enclosed in double quotes, it's a terminal, else it's a rule name.
func (itm ListItem) String() string {
	// Start with rule name or terminal string
	var result strings.Builder
	if len(itm.ruleName) > 0 {
		result.WriteString(itm.ruleName)
	}

	result.WriteString(itm.terminal.String())

	// Add options without any spaces between them
	for _, opt := range itm.options {
		result.WriteString(lexTypeStrings[opt])
	}

	return result.String()
}

// ====

// ExpressionItem is a group of one or more list items that are repeated.
// N and M are the lower and upper bounds, respectively.
// There is always a lower bound.
// If M == -1, there is no upper bound.
type ExpressionItem struct {
	list []ListItem
	n    int
	m    int
}

// Items is the list items
func (itm ExpressionItem) Items() []ListItem {
	return itm.list
}

// Repetitions returns the mnumbeer of repetitions (N, M) of the item.
// N is the lower bound, it is >= 0.
// M is the upper bound, it is -1 if there is no upper bound, else >= 0.
func (itm ExpressionItem) Repetitions() (n, m int) {
	return itm.n, itm.m
}

// String returns a formatted string, with a single space between each list item.
// If there are repetitions, then the items are enclosed in parantheses, followed by the repetition as follows:
// N = 0, M = 1: ?
// N = 0, M = -1: *
// N = 1, M = -1: +
// N > 0, M = N: {N}
// N > 0, M = -1: {N,}
// N = 0, M > 0: {,M}
// N > 0, M > 0: {N,M}
func (itm ExpressionItem) String() string {
	var (
		result          strings.Builder
		haveRepetitions = !((itm.n == 1) && (itm.m == 1))
	)

	// Opening parantheses if there are repetitions
	if haveRepetitions {
		result.WriteRune('(')
	}

	// Space-separated list of items
	for i, listItem := range itm.list {
		if i > 0 {
			result.WriteRune(' ')
		}

		result.WriteString(listItem.String())
	}

	// Closing parantheses and repetitions, if there are any
	if haveRepetitions {
		result.WriteRune(')')

		switch {
		// ?
		case (itm.n == 0) && (itm.m == 1):
			result.WriteRune('?')

		// *
		case (itm.n == 0) && (itm.m == -1):
			result.WriteRune('*')

		// +
		case (itm.n == 1) && (itm.m == -1):
			result.WriteRune('+')

		// {N}
		case (itm.n > 0) && (itm.m == itm.n):
			result.WriteRune('{')
			result.WriteString(strconv.Itoa(itm.n))
			result.WriteRune('}')

		// {N,}
		case (itm.n > 0) && (itm.m == -1):
			result.WriteRune('{')
			result.WriteString(strconv.Itoa(itm.n))
			result.WriteRune(',')
			result.WriteRune('}')

		// {,M}
		case (itm.n == 0) && (itm.m > 0):
			result.WriteRune('{')
			result.WriteRune(',')
			result.WriteString(strconv.Itoa(itm.m))
			result.WriteRune('}')

		// {N,M}
		case (itm.n > 0) && (itm.m > 0):
			result.WriteRune('{')
			result.WriteString(strconv.Itoa(itm.n))
			result.WriteRune(',')
			result.WriteString(strconv.Itoa(itm.m))
			result.WriteRune('}')
		}
	}

	return result.String()
}

// ====

// Expression is one or more expression items
type Expression struct {
	items []ExpressionItem
}

// Items is the expression items
func (e Expression) Items() []ExpressionItem {
	return e.items
}

// String returns a formatted string, with a bar between each expression item
func (e Expression) String() string {
	var (
		result            strings.Builder
		sameLineSeparator = " | "
		nextLineSeparator = "\n    | "
		isSameLine        = true
	)

	// try to write on one line, but if line gets > 80 chars, switch to multiline
	for i, itm := range e.items {
		if i > 0 {
			if isSameLine {
				// Write separator on same line
				result.WriteString(sameLineSeparator)

				if result.Len() >= 80 {
					// Exceeded line length - replace every sameLineSeparator with a nextLineSeparator
					currentResult := result.String()
					result.Reset()
					result.WriteString(strings.ReplaceAll(currentResult, sameLineSeparator, nextLineSeparator))

					// Switch line mode to every item on a line by itself
					isSameLine = false
				}
			} else {
				result.WriteString(nextLineSeparator)
			}
		}

		result.WriteString(itm.String())
	}

	return result.String()
}

// ====

// Rule is a rule name and expression
type Rule struct {
	name string
	expr Expression
}

// Name the rule name
func (r Rule) Name() string {
	return r.name
}

// Expr the expression
func (r Rule) Expr() Expression {
	return r.expr
}

// String returns a formatted string, as rulename = expression;
func (r Rule) String() string {
	var result strings.Builder

	result.WriteString(r.name)
	result.WriteString(" = ")
	result.WriteString(r.expr.String())
	result.WriteRune(';')

	return result.String()
}

// ====

// Grammar is one or more rules
type Grammar struct {
	rules []Rule
}

// Rules returns the set of rules
func (g Grammar) Rules() []Rule {
	return g.rules
}

// String returns a formatted string as a series of rules, with a newline after each one
func (g Grammar) String() string {
	var result strings.Builder

	for i, r := range g.rules {
		if i > 0 {
			result.WriteRune('\n')
		}

		result.WriteString(r.String())
	}

	return result.String()
}
