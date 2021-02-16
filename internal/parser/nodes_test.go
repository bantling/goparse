package parser

import (
	"testing"

	"github.com/bantling/goparse/internal/lexer"
	"github.com/stretchr/testify/assert"
)

func TestTerminal(t *testing.T) {
	src := "'single \\\\ \\t \\r \\n \\' \" quoted'"
	str := "single \\ \t \r \n ' \" quoted"
	term := OfTerminalString(src, str)
	assert.True(t, term.IsString())
	assert.False(t, term.IsRange())
	assert.Equal(t, str, term.TerminalString())
	assert.Equal(t, map[rune]bool(nil), term.TerminalRange())
	assert.Equal(t, src, term.String())

	src = "[A-C]"
	rng := map[rune]bool{'A': true, 'B': true, 'C': true}
	term = OfTerminalRange(src, rng)
	assert.False(t, term.IsString())
	assert.True(t, term.IsRange())
	assert.Equal(t, "", term.TerminalString())
	assert.Equal(t, rng, term.TerminalRange())
	assert.Equal(t, src, term.String())
}

func TestListItem(t *testing.T) {
	src := "myrulename"
	name := src
	item := OfListItemRuleName(src, name, nil)
	assert.True(t, item.IsRuleName())
	assert.False(t, item.IsTerminal())
	assert.Equal(t, name, item.RuleName())
	assert.Equal(t, Terminal{}, item.Terminal())
	assert.Equal(t, src, item.String())

	src = "myrulename:AST"
	name = "myrulename"
	item = OfListItemRuleName(src, name, []lexer.LexType{lexer.OptionAST})
	assert.True(t, item.IsRuleName())
	assert.False(t, item.IsTerminal())
	assert.Equal(t, name, item.RuleName())
	assert.Equal(t, Terminal{}, item.Terminal())
	assert.Equal(t, src, item.String())

	src = "'single \\\\ \\t \\r \\n \\' \" quoted'"
	str := "single \\ \t \r \n ' \" quoted"
	term := OfTerminalString(src, str)
	item = OfListItemTerminal(src, term, nil)
	assert.False(t, item.IsRuleName())
	assert.True(t, item.IsTerminal())
	assert.Equal(t, "", item.RuleName())
	assert.Equal(t, term, item.Terminal())
	assert.Equal(t, src, item.String())

	src = "'single \\\\ \\t \\r \\n \\' \" quoted':EOL:INDENT"
	str = "single \\ \t \r \n ' \" quoted"
	term = OfTerminalString(src, str)
	item = OfListItemTerminal(src, term, []lexer.LexType{lexer.OptionEOL, lexer.OptionIndent})
	assert.False(t, item.IsRuleName())
	assert.True(t, item.IsTerminal())
	assert.Equal(t, "", item.RuleName())
	assert.Equal(t, term, item.Terminal())
	assert.Equal(t, src, item.String())

	src = "[A-C]"
	term = OfTerminalRange(src, map[rune]bool{'A': true, 'B': true, 'C': true})
	item = OfListItemTerminal(src, term, nil)
	assert.False(t, item.IsRuleName())
	assert.True(t, item.IsTerminal())
	assert.Equal(t, "", item.RuleName())
	assert.Equal(t, term, item.Terminal())
	assert.Equal(t, src, item.String())

	src = "[A-C]:OUTDENT"
	term = OfTerminalRange(src, map[rune]bool{'A': true, 'B': true, 'C': true})
	item = OfListItemTerminal(src, term, []lexer.LexType{lexer.OptionOutdent})
	assert.False(t, item.IsRuleName())
	assert.True(t, item.IsTerminal())
	assert.Equal(t, "", item.RuleName())
	assert.Equal(t, term, item.Terminal())
	assert.Equal(t, src, item.String())

}

func TestExpressionItem(t *testing.T) {
	src := "myrulename"
	name := src
	item := OfListItemRuleName(src, name, nil)
	items := []ListItem{item}
	exprItem := OfExpressionItem(src, items, 1, 1)
	n, m := exprItem.Repetitions()

	assert.Equal(t, items, exprItem.Items())
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, m)
	assert.Equal(t, src, exprItem.String())

	src = "(myrulename){2,3}"
	name = "myrulename"
	item = OfListItemRuleName(src, name, nil)
	items = []ListItem{item}
	exprItem = OfExpressionItem(src, items, 2, 3)
	n, m = exprItem.Repetitions()

	assert.Equal(t, items, exprItem.Items())
	assert.Equal(t, 2, n)
	assert.Equal(t, 3, m)
	assert.Equal(t, src, exprItem.String())
}

func TestExpression(t *testing.T) {
	var (
		allSrc   string
		allItems []ExpressionItem
	)

	src := "myfirstrulename"
	name := src
	item := OfListItemRuleName(src, name, nil)
	items := []ListItem{item}
	exprItem := OfExpressionItem(src, items, 1, 1)
	exprItems := []ExpressionItem{exprItem}
	expr := OfExpression(src, exprItems)
	assert.Equal(t, exprItems, expr.Items())
	assert.Equal(t, src, expr.String())

	allSrc = src
	allItems = append(allItems, exprItem)

	src = "mysecondrulename"
	name = src
	item = OfListItemRuleName(src, name, nil)
	items = []ListItem{item}
	exprItem = OfExpressionItem(src, items, 1, 1)
	exprItems = []ExpressionItem{exprItem}
	expr = OfExpression(src, exprItems)
	assert.Equal(t, exprItems, expr.Items())
	assert.Equal(t, src, expr.String())

	allSrc = allSrc + " | " + src
	allItems = append(allItems, exprItem)

	// Multiple items
	expr = OfExpression(allSrc, allItems)
	assert.Equal(t, allItems, expr.Items())
	assert.Equal(t, allSrc, expr.String())
}
