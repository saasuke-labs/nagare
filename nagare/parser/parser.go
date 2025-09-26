package parser

import (
	"errors"
	"fmt"
	"nagare/tokenizer"
	"strings"
)

// NodeType represents the type of AST node
type NodeType string

const (
	NODE_ELEMENT   NodeType = "Element"   // Default type for leaf nodes
	NODE_CONTAINER NodeType = "Container" // Default type for nodes with children
)

// Node represents a node in the AST
type Node struct {
	Type     NodeType // Can be a predefined type or a custom type string
	Text     string   // The name/label of the node
	Children []Node
	Depth    int // Track nesting level
}

func (n Node) String() string {
	tabs := strings.Repeat("  ", n.Depth)
	childrenStr := ""

	for _, child := range n.Children {
		childrenStr += "<" + child.String() + ">"
	}

	if childrenStr != "" {
		return fmt.Sprintf("%s%s(%d) {\n%s%s}\n", tabs, n.Text, len(n.Children), childrenStr, tabs)
	}

	return fmt.Sprintf("%s%s\n", tabs, n.Text)
}

// Parse converts tokens into an AST
func Parse(tokens []tokenizer.Token) (Node, error) {
	parser := &Parser{
		tokens:  tokens,
		current: 0,
	}
	return parser.parse(0)
}

type Parser struct {
	tokens  []tokenizer.Token
	current int
}

func (p *Parser) parse(depth int) (Node, error) {
	if depth > 1 {
		return Node{}, errors.New("nesting depth exceeded maximum of 1")
	}

	root := Node{
		Type:  NODE_ELEMENT,
		Depth: depth,
	}

	for p.current < len(p.tokens) {
		token := p.tokens[p.current]

		switch token.Type {
		case tokenizer.RIGHT_BRACE:
			if depth == 0 {
				return Node{}, errors.New("unexpected closing brace at root level")
			}
			return root, nil
		case tokenizer.IDENTIFIER:
			// Look ahead for type declaration
			nodeName := strings.TrimSpace(token.Value)
			nodeType := NODE_ELEMENT // Default type
			p.current++              // Move past identifier

			// Check if next token is a colon (type declaration)
			if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenizer.COLON {
				p.current++ // Move past colon
				if p.current >= len(p.tokens) {
					return Node{}, errors.New("unexpected end of input after colon")
				}
				if p.tokens[p.current].Type != tokenizer.IDENTIFIER {
					return Node{}, errors.New("expected type after colon")
				}
				// Use the declared type
				nodeType = NodeType(p.tokens[p.current].Value)
				p.current++ // Move past type
			}

			// Check if it's a container (has braces)
			isContainer := p.current < len(p.tokens) &&
				p.tokens[p.current].Type == tokenizer.LEFT_BRACE

			if isContainer {
				// This is a container node
				containerNode := Node{
					Type:  NODE_CONTAINER, // Containers always use NODE_CONTAINER type
					Text:  nodeName,
					Depth: depth,
				}
				p.current++ // Skip the left brace

				// Parse children until we find closing brace
				for p.current < len(p.tokens) && p.tokens[p.current].Type != tokenizer.RIGHT_BRACE {
					childNode, err := p.parse(depth + 1)
					if err != nil {
						return Node{}, err
					}
					containerNode.Children = append(containerNode.Children, childNode)
				}

				if p.current >= len(p.tokens) {
					return Node{}, errors.New("unexpected end of input: missing closing brace")
				}
				if p.tokens[p.current].Type != tokenizer.RIGHT_BRACE {
					return Node{}, errors.New("expected closing brace")
				}
				p.current++ // Skip the right brace

				if depth == 0 {
					root.Children = append(root.Children, containerNode)
				} else {
					return containerNode, nil
				}
			} else {
				// Regular node
				node := Node{
					Type:  nodeType, // Use declared type or default
					Text:  nodeName,
					Depth: depth,
				}

				if depth == 0 {
					root.Children = append(root.Children, node)
				} else {
					return node, nil
				}
			}
		default:
			p.current++
		}
	}

	return root, nil
}
