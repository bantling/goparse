package parser

import (
	"strings"

	"github.com/bantling/goparse/internal/lexer"
)

// ====

// LexicalNode is the base structure for all nodes that want the original lexical string as part or all of String().
type LexicalNode struct {
	lexicalString string
}

// String returns the origin lexical string
func (n LexicalNode) String() string {
	return n.lexicalString
}

// ====

// Terminal is a string or character range.
// If the string is "", then the terminal is a character range, else it is a string.
type Terminal struct {
	LexicalNode
	theString string
	theRange  map[rune]bool
}

// OfTerminalString constructs a Terminal from a string
func OfTerminalString(lexicalString, terminalString string) Terminal {
	return Terminal{
		LexicalNode: LexicalNode{
			lexicalString: lexicalString,
		},
		theString: terminalString,
	}
}

// OfTerminalRange constructs a Terminal from a range
func OfTerminalRange(lexicalString string, theRange map[rune]bool) Terminal {
	return Terminal{
		LexicalNode: LexicalNode{
			lexicalString: lexicalString,
		},
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

// ====

// ListItem is a rule name or a terminal, and possibly some options.
// If the rule name is "", then the item is a terminal, else it is a rule name.
// Options can be applied to a rule name or a terminal.
type ListItem struct {
	LexicalNode
	ruleName string
	terminal Terminal
	options  []lexer.LexType
}

// OfListItemRuleName constructs a ListItem from a rule name and options
func OfListItemRuleName(lexicalString, ruleName string, options []lexer.LexType) ListItem {
	return ListItem{
		LexicalNode: LexicalNode{
			lexicalString: lexicalString,
		},
		ruleName: ruleName,
		options:  options,
	}
}

// OfListItemTerminal constructs a ListItem from a terminal and options
func OfListItemTerminal(terminal Terminal, options []lexer.LexType) ListItem {
	return ListItem{
		LexicalNode: terminal.LexicalNode,
		terminal:    terminal,
		options:     options,
	}
}

// IsRuleName returns true if the ListItem was constructed with a rule name
func (itm ListItem) IsRuleName() bool {
	return len(itm.ruleName) > 0
}

// IsTerminal returns true if the ListItem was constructed with a terminal
func (itm ListItem) IsTerminal() bool {
	return len(itm.ruleName) == 0
}

// RuleName is the rule name
func (itm ListItem) RuleName() string {
	return itm.ruleName
}

// Terminal is the terminal
func (itm ListItem) Terminal() Terminal {
	return itm.terminal
}

// ====

// ExpressionItem is a group of one or more list items that are repeated.
// N and M are the lower and upper bounds, respectively.
// There is always a lower bound.
// If M == -1, there is no upper bound.
type ExpressionItem struct {
	LexicalNode
	list []ListItem
	n    int
	m    int
}

// OfExpressionItem constructs an ExpressionItem from a list of ListItem and n, m repetitions
func OfExpressionItem(lexicalString string, list []ListItem, n, m int) ExpressionItem {
	return ExpressionItem{
		LexicalNode: LexicalNode{
			lexicalString: lexicalString,
		},
		list: list,
		n:    n,
		m:    m,
	}
}

// Items is the list items
func (itm ExpressionItem) Items() []ListItem {
	return itm.list
}

// Repetitions returns the number of repetitions (N, M) of the item.
// N is the lower bound, it is >= 0.
// M is the upper bound, it is -1 if there is no upper bound, else >= 0.
func (itm ExpressionItem) Repetitions() (n, m int) {
	return itm.n, itm.m
}

// ====

// Expression is one or more expression items
type Expression struct {
	items []ExpressionItem
}

// OfExpression constructs a Expression from a list of expression items
func OfExpression(items []ExpressionItem) Expression {
	return Expression{
		items: items,
	}
}

// Items is the expression items
func (e Expression) Items() []ExpressionItem {
	return e.items
}

// String returns a formatted string, with a bar between each expression item.
// If the expression fits within 80 chars, it is on one line, else it is multi-line.
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

// OfRule constructs a rule from a name and expression
func OfRule(name string, expr Expression) Rule {
	return Rule{
		name: name,
		expr: expr,
	}
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

// OfGrammar constructs a Grammar from a list of rules
func OfGrammar(rules []Rule) Grammar {
	return Grammar{
		rules: rules,
	}
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
