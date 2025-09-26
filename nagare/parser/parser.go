package parser

import (
	"errors"
	"fmt"
	"nagare/tokenizer"
	"strings"
)

// NodeType represents the type of AST node
type NodeType int

const (
	NODE_ELEMENT   NodeType = iota
	NODE_CONTAINER          // Node that has nested elements
)

// Node represents a node in the AST
type Node struct {
	Type     NodeType
	Text     string
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
		case tokenizer.IDENTIFIER:
			if p.current+1 < len(p.tokens) && p.tokens[p.current+1].Type == tokenizer.LEFT_BRACE {
				// This is a container node
				containerNode := Node{
					Type:  NODE_CONTAINER,
					Text:  strings.TrimSpace(token.Value),
					Depth: depth,
				}
				p.current += 2 // Skip the identifier and left brace

				// Parse children until we find closing brace
				for p.current < len(p.tokens) {
					if p.tokens[p.current].Type == tokenizer.RIGHT_BRACE {
						break
					}
					if p.tokens[p.current].Type == tokenizer.IDENTIFIER {
						node := Node{
							Type:  NODE_ELEMENT,
							Text:  strings.TrimSpace(p.tokens[p.current].Value),
							Depth: depth + 1,
						}
						containerNode.Children = append(containerNode.Children, node)
					}
					p.current++
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
					Type:  NODE_ELEMENT,
					Text:  strings.TrimSpace(token.Value),
					Depth: depth,
				}

				if depth == 0 {
					root.Children = append(root.Children, node)
				} else {
					return node, nil
				}
				p.current++
			}
		case tokenizer.RIGHT_BRACE:
			if depth == 0 {
				return Node{}, errors.New("unexpected closing brace")
			}
			return root, nil
		default:
			p.current++
		}
	}

	return root, nil
}
