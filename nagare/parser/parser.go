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

// State represents a named set of props for a node
type State struct {
	Name     string
	PropsDef string // Raw props definition string to be parsed by components
}

// Node represents a node in the AST
type Node struct {
	Type        NodeType // Can be a predefined type or a custom type string
	Text        string   // The name/label of the node
	Children    []Node
	Depth       int    // Track nesting level
	State       string // Current state name if specified with @
	States      map[string]State
	Globals     map[string]State
	Connections []Connection
}

// AnchorDescriptor captures anchor metadata for connection endpoints.
type AnchorDescriptor struct {
	Raw        string
	Horizontal float64
	Vertical   float64
}

// Connection represents a link between two nodes in the AST.
type Connection struct {
	FromID     string
	FromAnchor AnchorDescriptor
	ToID       string
	ToAnchor   AnchorDescriptor
	Style      string
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

// findNodesWithState returns all nodes in the tree that use the given state
func (p *Parser) findNodesWithState(root *Node, stateName string) []*Node {
	var nodes []*Node
	if root.State == stateName {
		nodes = append(nodes, root)
	}
	for i := range root.Children {
		nodes = append(nodes, p.findNodesWithState(&root.Children[i], stateName)...)
	}
	return nodes
}

// findNodesWithName returns all nodes in the tree that have the given identifier
func (p *Parser) findNodesWithName(root *Node, name string) []*Node {
	var nodes []*Node
	if root.Text == name {
		nodes = append(nodes, root)
	}
	for i := range root.Children {
		nodes = append(nodes, p.findNodesWithName(&root.Children[i], name)...)
	}
	return nodes
}

func (p *Parser) parseState() (*State, error) {
	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenizer.AT {
		return nil, errors.New("expected @ for state definition")
	}
	p.current++ // Move past @

	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenizer.IDENTIFIER {
		return nil, errors.New("expected state name after @")
	}
	name := p.tokens[p.current].Value
	p.current++ // Move past state name

	if p.current >= len(p.tokens) || p.tokens[p.current].Type != tokenizer.LEFT_PAREN {
		return nil, errors.New("expected ( after state name")
	}
	p.current++ // Move past (

	// Collect everything until the matching )
	var propsDef strings.Builder
	parenCount := 1
	inQuotes := false

	for p.current < len(p.tokens) && parenCount > 0 {
		token := p.tokens[p.current]

		switch {
		case token.Type == tokenizer.IDENTIFIER && (token.Value == "\"" || token.Value == "'"):
			inQuotes = !inQuotes
			propsDef.WriteString("\"") // Always use double quotes
		case token.Type == tokenizer.COMMA && !inQuotes:
			propsDef.WriteString(",")
		case token.Type == tokenizer.COLON && !inQuotes:
			propsDef.WriteString(":")
		case token.Type == tokenizer.LEFT_PAREN:
			parenCount++
			if parenCount > 1 {
				propsDef.WriteString("(")
			}
		case token.Type == tokenizer.RIGHT_PAREN:
			parenCount--
			if parenCount > 0 {
				propsDef.WriteString(")")
			}
		default:
			propsDef.WriteString(token.Value)
			if !inQuotes && token.Type != tokenizer.COLON && token.Type != tokenizer.COMMA {
				nextToken := p.peekNext()
				if nextToken != nil && nextToken.Type != tokenizer.COLON && nextToken.Type != tokenizer.COMMA {
					propsDef.WriteString(" ")
				}
			}
		}
		p.current++
	}

	if parenCount > 0 {
		return nil, errors.New("unclosed parenthesis in state definition")
	}

	return &State{
		Name:     name,
		PropsDef: propsDef.String(),
	}, nil
}

func (p *Parser) peekNext() *tokenizer.Token {
	if p.current+1 < len(p.tokens) {
		return &p.tokens[p.current+1]
	}
	return nil
}

func (p *Parser) parse(depth int) (Node, error) {
	if depth > 1 {
		return Node{}, errors.New("nesting depth exceeded maximum of 1")
	}

	root := Node{
		Type:    NODE_ELEMENT,
		Depth:   depth,
		Globals: make(map[string]State),
	}

	for p.current < len(p.tokens) {
		token := p.tokens[p.current]

		switch token.Type {
		case tokenizer.AT:
			if depth > 0 {
				return Node{}, errors.New("state definitions must be at root level")
			}
			state, err := p.parseState()
			if err != nil {
				return Node{}, err
			}

			// Store the state definition for lookup during layout/render phases
			root.Globals[state.Name] = *state

			// Associate the state with nodes that explicitly reference it by state name
			fmt.Printf("State %s props: %s\n", state.Name, state.PropsDef)
			nodesByState := p.findNodesWithState(&root, state.Name)
			for _, node := range nodesByState {
				if node.States == nil {
					node.States = make(map[string]State)
				}
				node.States[state.Name] = *state
			}

			// Also associate the state with nodes that match the identifier/name
			nodesByName := p.findNodesWithName(&root, state.Name)
			for _, node := range nodesByName {
				if node.States == nil {
					node.States = make(map[string]State)
				}
				node.States[state.Name] = *state
			}

		case tokenizer.RIGHT_BRACE:
			if depth == 0 {
				return Node{}, errors.New("unexpected closing brace at root level")
			}
			return root, nil
		case tokenizer.IDENTIFIER:
			if connection, ok, err := p.tryParseConnection(); err != nil {
				return Node{}, err
			} else if ok {
				root.Connections = append(root.Connections, *connection)
				continue
			}

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

			// Check for state declaration with @
			var stateName string
			if p.current < len(p.tokens) && p.tokens[p.current].Type == tokenizer.AT {
				p.current++ // Move past @
				if p.current >= len(p.tokens) {
					return Node{}, errors.New("unexpected end of input after @")
				}
				if p.tokens[p.current].Type != tokenizer.IDENTIFIER {
					return Node{}, errors.New("expected state name after @")
				}
				stateName = p.tokens[p.current].Value
				p.current++ // Move past state name
			}

			// Check if it's a container (has braces)
			isContainer := p.current < len(p.tokens) &&
				p.tokens[p.current].Type == tokenizer.LEFT_BRACE

			var states map[string]State

			if isContainer {
				if nodeType == NODE_ELEMENT {
					nodeType = NODE_CONTAINER
				}
				// This is a container node
				containerNode := Node{
					Type:   nodeType, // Use declared type or default
					Text:   nodeName,
					Depth:  depth,
					State:  stateName,
					States: states,
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
					Type:   nodeType, // Use declared type or default
					Text:   nodeName,
					Depth:  depth,
					State:  stateName,
					States: states,
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

	if len(root.Globals) == 0 {
		root.Globals = nil
	}
	return root, nil
}

func (p *Parser) tryParseConnection() (*Connection, bool, error) {
	start := p.current
	if start >= len(p.tokens) || p.tokens[start].Type != tokenizer.IDENTIFIER {
		return nil, false, nil
	}

	if start+6 >= len(p.tokens) {
		return nil, false, nil
	}

	fromID := strings.TrimSpace(p.tokens[start].Value)
	if p.tokens[start+1].Type != tokenizer.DOT {
		return nil, false, nil
	}

	if p.tokens[start+2].Type != tokenizer.IDENTIFIER {
		return nil, false, errors.New("expected anchor identifier after dot in connection")
	}

	if p.tokens[start+3].Type != tokenizer.ARROW {
		return nil, false, nil
	}

	if p.tokens[start+4].Type != tokenizer.IDENTIFIER {
		return nil, false, errors.New("expected target identifier after connection arrow")
	}

	if p.tokens[start+5].Type != tokenizer.DOT {
		return nil, false, errors.New("expected dot before target anchor in connection")
	}

	if p.tokens[start+6].Type != tokenizer.IDENTIFIER {
		return nil, false, errors.New("expected anchor identifier after dot in connection")
	}

	toID := strings.TrimSpace(p.tokens[start+4].Value)

	connection := &Connection{
		FromID:     fromID,
		FromAnchor: parseAnchorDescriptor(p.tokens[start+2].Value),
		ToID:       toID,
		ToAnchor:   parseAnchorDescriptor(p.tokens[start+6].Value),
	}

	p.current = start + 7
	return connection, true, nil
}

func parseAnchorDescriptor(raw string) AnchorDescriptor {
	descriptor := AnchorDescriptor{Raw: strings.TrimSpace(raw)}
	if descriptor.Raw == "" {
		return descriptor
	}

	normalized := strings.ToLower(descriptor.Raw)
	horizontal := 0.0
	vertical := 0.0

	for _, r := range normalized {
		switch r {
		case 'w':
			horizontal = -1.0
		case 'e':
			horizontal = 1.0
		case 'n':
			vertical = -1.0
		case 's':
			vertical = 1.0
		}
	}

	descriptor.Horizontal = horizontal
	descriptor.Vertical = vertical
	return descriptor
}
