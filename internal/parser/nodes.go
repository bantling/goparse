package parser

import (
	"fmt"
)

// ====

// Terminal is a string or character range.
// If the string is "", then the terminal is a character range, else it is a string.
type Terminal struct {
	theString string
	theRange  map[rune]bool
}

// TerminalString is the terminal string
func (t Terminal) TerminalString() string {
	return t.theString
}

// TerminalRange is the terminal range
func (t Terminal) TerminalRange() map[rune]bool {
	return t.theRange
}

// String returns a formatted string for the Terminal, using double quotes and escaping control characters and backslashes
func (t Terminal) String() string {
	return fmt.Sprintf("%q", t.theString)
}

// ====

// ListItem is a rule name or a terminal.
// If the rule name is "", then the item is a terminal, else it is a rule name.
type ListItem struct {
	ruleName string
	terminal Terminal
}

// RuleName is the rule name
func (i ListItem) RuleName() string {
	return i.ruleName
}

// Terminal is the terminal
func (i ListItem) Terminal() Terminal {
	return i.terminal
}

// String returns a formatted string, using rule name or terminal.String().
// If the result is enclosed in double quotes, it's a terminal, else it's a rule name.
func (i ListItem) String() string {
	if len(i.ruleName) > 0 {
		return i.ruleName
	}

	return i.terminal.String()
}

// ====

// ExpressionItem is a group of one or more list items that are repeated.
// n and m are the lower and upper bounds, respectively.
// There is always a lower bound.
// If m == -1, there is no upper bound.
type ExpressionItem struct {
	list []ListItem
	n    int
	m    int
}

// Items is the list items
func (i ExpressionItem) Items() []ListItem {
	return i.list
}

// N is the lower bound, it is >= 0.
func (i ExpressionItem) N() int {
	return i.n
}

// M is the upper bound, it is -1 if there is no upper bound, else >= 0.
func (i ExpressionItem) M() int {
	return i.m
}

// String returns a formatted string, with a single space between each ListItem.
// If there are repetitions, then the items are enclosed in parantheses, followed by the repetition as follows:
// N = 0, M = -1: *
// N = 1, M = -1: +
// N >= 0, M = N: {N}
// N > 0, M = -1: {N,}
// N = 0, M > 0: {,M}
// N > 0, M >= 0: {N,M}
//func (i ExpressionItem) String() string {
//
//}

// ====

// Expression is one or more expression items
type Expression struct {
	items []ExpressionItem
}

// Items is the expression items
func (e Expression) Items() []ExpressionItem {
	return e.items
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

// ====

// Grammar is one or more rules
type Grammar struct {
	rules []Rule
}

// Rules the set of rules
func (g Grammar) Rules() []Rule {
	return g.rules
}
