package parser

// NodeType is the type of a parse node
type NodeType uint

// NodeType constants
const (
	InvalidNodeType NodeType = iota
	Rule
	Identifier
	Terminal
	Optional
	Repetition
	Grouping
	Alternation
	Concatenation
	Exception
)

// Node is a single nodde in the parser tree
type Node struct {
}

// Parser performs parsing
type Parser struct {
}
