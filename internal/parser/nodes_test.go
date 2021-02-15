package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminal(t *testing.T) {
	lex := "single \\\\ \\t \\r \\n \\' \" quoted"
	str := "single \\ \t \r \n ' \" quoted"
	term := OfTerminalString(lex, str)
	assert.True(t, term.IsString())
	assert.False(t, term.IsRange())
	assert.Equal(t, str, term.TerminalString())
	assert.Equal(t, map[rune]bool(nil), term.TerminalRange())
	assert.Equal(t, lex, term.String())

	lex = "[A-C]"
	rng := map[rune]bool{'A': true, 'B': true, 'C': true}
	term = OfTerminalRange(lex, rng)
	assert.False(t, term.IsString())
	assert.True(t, term.IsRange())
	assert.Equal(t, "", term.TerminalString())
	assert.Equal(t, rng, term.TerminalRange())
	assert.Equal(t, lex, term.String())
}
