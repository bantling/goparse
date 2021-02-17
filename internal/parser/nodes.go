package parser

import (
	"github.com/bantling/goparse/internal/lexer"
)

// ====

// SourceNode is the base structure for all nodes that provides the original source text via String()
type SourceNode struct {
	sourceString string
}

// OfSourceNode constructs a SourceNode
func OfSourceNode(sourceString string) SourceNode {
	return SourceNode{sourceString: sourceString}
}

// String returns the origin source string
func (s SourceNode) String() string {
	return s.sourceString
}

// ====

// Terminal is a string or character range
type Terminal struct {
	SourceNode
	theString string
	theRange  map[rune]bool
}

// OfTerminalString constructs a Terminal from a string
func OfTerminalString(sourceString, terminalString string) Terminal {
	return Terminal{
		SourceNode: OfSourceNode(sourceString),
		theString:  terminalString,
	}
}

// OfTerminalRange constructs a Terminal from a range
func OfTerminalRange(sourceString string, theRange map[rune]bool) Terminal {
	return Terminal{
		SourceNode: OfSourceNode(sourceString),
		theRange:   theRange,
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
	SourceNode
	ruleName string
	terminal Terminal
	options  []lexer.LexType
}

// OfListItemRuleName constructs a ListItem from a rule name and options
func OfListItemRuleName(sourceString string, ruleName string, options []lexer.LexType) ListItem {
	return ListItem{
		SourceNode: OfSourceNode(sourceString),
		ruleName:   ruleName,
		options:    options,
	}
}

// OfListItemTerminal constructs a ListItem from a terminal and options
func OfListItemTerminal(sourceString string, terminal Terminal, options []lexer.LexType) ListItem {
	return ListItem{
		SourceNode: OfSourceNode(sourceString),
		terminal:   terminal,
		options:    options,
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
	SourceNode
	list []ListItem
	n    int
	m    int
}

// OfExpressionItem constructs an ExpressionItem from a list of ListItem and n, m repetitions
func OfExpressionItem(sourceString string, list []ListItem, n, m int) ExpressionItem {
	return ExpressionItem{
		SourceNode: OfSourceNode(sourceString),
		list:       list,
		n:          n,
		m:          m,
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
	SourceNode
	items []ExpressionItem
}

// OfExpression constructs a Expression from a list of expression items
func OfExpression(sourceString string, items []ExpressionItem) Expression {
	return Expression{
		SourceNode: OfSourceNode(sourceString),
		items:      items,
	}
}

// Items is the expression items
func (e Expression) Items() []ExpressionItem {
	return e.items
}

// ====

// Rule is a rule name and expression
type Rule struct {
	SourceNode
	name string
	expr Expression
}

// OfRule constructs a rule from a name and expression
func OfRule(sourceString string, name string, expr Expression) Rule {
	return Rule{
		SourceNode: OfSourceNode(sourceString),
		name:       name,
		expr:       expr,
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

// ====

// Grammar is one or more rules
type Grammar struct {
	SourceNode
	rules []Rule
}

// OfGrammar constructs a Grammar from a list of rules
func OfGrammar(sourceString string, rules []Rule) Grammar {
	return Grammar{
		SourceNode: OfSourceNode(sourceString),
		rules:      rules,
	}
}

// Rules returns the set of rules
func (g Grammar) Rules() []Rule {
	return g.rules
}
