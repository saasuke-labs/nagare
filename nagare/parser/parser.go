package parser

import "nagare/tokenizer"

// NodeType represents the type of AST node
type NodeType int

const (
	NODE_ELEMENT NodeType = iota
)

// Node represents a node in the AST
type Node struct {
	Type     NodeType
	Text     string
	Children []Node
}

// Parse converts tokens into an AST
func Parse(tokens []tokenizer.Token) Node {
	// For now, we create a single element node
	if len(tokens) > 0 {
		return Node{
			Type: NODE_ELEMENT,
			Text: tokens[0].Value,
		}
	}
	return Node{} // Empty node if no tokens
}
