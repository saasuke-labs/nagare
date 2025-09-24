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
	root := Node{
		Type: NODE_ELEMENT,
		Text: "", // Root node doesn't have text
	}

	// Create a child node for each token
	for _, token := range tokens {
		child := Node{
			Type: NODE_ELEMENT,
			Text: token.Value,
		}
		root.Children = append(root.Children, child)
	}

	return root
}
