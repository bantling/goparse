package parser

//import (
//	"strings"
//
//	"github.com/bantling/goparse/internal/lexer"
//)
//
//// Error message constants
////const (
////	ErrNotATerminal = "Expected a string (single or double quoted) or a character range"
////	ErrNotAListItem = "Expected
////)
//
//// Parser is the recursive descent parser that converts source text into a Grammar
//type Parser struct {
//	lex *lexer.Lexer
//	unreadToken lexer.Token
//}
//
//// ofParser constructs a Parser from an io.Reader
//func ofParser(source io.Reader) Parser {
//	return Parser{
//		lex: lexer.NewLexer(source),
//	}
//}
//
//// nextToken reads the next token, which may be buffered or may require a call to the lexer
//func (p Parser) nextToken() lexer.Token {
//	var result lexer.Token
//
//	if p.unreadToken.Type() == lexer.InvalidLexType {
//		result = p.lex.Next()
//	} else {
//		result = p.unreadToken
//		p.unreadToken = lexer.Token{}
//	}
//
//	return result
//}
//
//// parseTerminal parses the terminal grammar rule.
////
//// <terminal-part> ::= <string> | <character-range>
//// <terminal-parts> ::= "" | <terminal-part> <terminal-parts>
//// <terminal> ::= <terminal-part> <terminal-parts>
////
//// parses as (String | CharacterRange)+
//func (p Parser) parseTerminal() Terminal {
//	var (
//		str strings.Builder
//
//	)
//
//	for token := p.nextToken() {
//		switch token.Type() {
//		case lexer.String:
//			result = OfTerminalString(token.String(), token.Token())
//
//		case lexer.CharacterRange:
//			result = OfTerminalRange(token.String(), token.Range())
//
//		default:
//			// Must be first token of next rule
//
//		}
//	}
//}
//
//// parseListItem parses the ListItem grammar rule.
////
//// <ast> ::= ":AST"
//// <fmt-eol> ::= ":EOL"
//// <fmt-indent> ::= ":INDENT"
//// <fmt-outdent> ::= "OUTDENT"
//// <list-item-option> ::= <ast> | <fmt-eol> | <fmt-indent> | <fmt-outdent>
//// <list-item-options> ::= "" | <list-item-option> <list-item-options>
//// <list-item> ::= <rule-name> <list-item-options> | <terminal> <list-item-options>
////
//// parses as Identifier (OptionAST | OptionEOL | OptionIndemt | OptionOutdent)*
//
//func (p Parser) parseListItem() ListItem {
//
//}
//
//// parseList parses the List nonterminal
//func (p Parser) parseList() List {
//
//}
//
//// parseExpressionItem parses the ExpressionItem nonterminal
//func (p Parser) parseExpressionItem() ExpressionItem {
//
//}
//
//// parseExpression parses the Expression nonterminal
//func (p Parser) parseExpression() Expression {
//
//}
//
//// parseRule parses the Rule nonterminal
//func (p Parser) parseRule() Rule {
//	return  Rule{}
//}
//
//// parseGrammar parses the Grammar nonterminal
//func (p Parser) parseRule() Grammar {
//	return Grammar{}
//}
