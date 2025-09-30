package parser

import (
	"nagare/tokenizer"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		tokens        []tokenizer.Token
		expected      Node
		expectedError string
	}{
		{
			name: "single identifier",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "Server"},
			},
			expected: Node{
				Type:  NODE_ELEMENT,
				Depth: 0,
				Children: []Node{
					{Type: NODE_ELEMENT, Text: "Server", Depth: 0},
				},
			},
		},
		{
			name: "container with children",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "VM"},
				{Type: tokenizer.LEFT_BRACE},
				{Type: tokenizer.IDENTIFIER, Value: "nginx"},
				{Type: tokenizer.IDENTIFIER, Value: "app"},
				{Type: tokenizer.RIGHT_BRACE},
			},
			expected: Node{
				Type:  NODE_ELEMENT,
				Depth: 0,
				Children: []Node{
					{
						Type:  NODE_CONTAINER,
						Text:  "VM",
						Depth: 0,
						Children: []Node{
							{Type: NODE_ELEMENT, Text: "nginx", Depth: 1},
							{Type: NODE_ELEMENT, Text: "app", Depth: 1},
						},
					},
				},
			},
		},
		{
			name: "too deep nesting",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "VM"},
				{Type: tokenizer.LEFT_BRACE},
				{Type: tokenizer.IDENTIFIER, Value: "Server"},
				{Type: tokenizer.LEFT_BRACE},
				{Type: tokenizer.IDENTIFIER, Value: "App"},
				{Type: tokenizer.RIGHT_BRACE},
				{Type: tokenizer.RIGHT_BRACE},
			},
			expectedError: "nesting depth exceeded maximum of 1",
		},
		{
			name:     "empty tokens",
			tokens:   []tokenizer.Token{},
			expected: Node{Type: NODE_ELEMENT, Depth: 0},
		},
		{
			name: "connection single-letter anchors",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "foo"},
				{Type: tokenizer.DOT},
				{Type: tokenizer.IDENTIFIER, Value: "w"},
				{Type: tokenizer.ARROW, Value: "-->"},
				{Type: tokenizer.IDENTIFIER, Value: "bar"},
				{Type: tokenizer.DOT},
				{Type: tokenizer.IDENTIFIER, Value: "e"},
			},
			expected: Node{
				Type:  NODE_ELEMENT,
				Depth: 0,
				Connections: []Connection{
					{
						FromID: "foo",
						FromAnchor: AnchorDescriptor{
							Raw:        "w",
							Horizontal: -1,
							Vertical:   0,
						},
						ToID: "bar",
						ToAnchor: AnchorDescriptor{
							Raw:        "e",
							Horizontal: 1,
							Vertical:   0,
						},
					},
				},
			},
		},
		{
			name: "connection compound anchors",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "source"},
				{Type: tokenizer.DOT},
				{Type: tokenizer.IDENTIFIER, Value: "wn"},
				{Type: tokenizer.ARROW, Value: "-->"},
				{Type: tokenizer.IDENTIFIER, Value: "sink"},
				{Type: tokenizer.DOT},
				{Type: tokenizer.IDENTIFIER, Value: "se"},
			},
			expected: Node{
				Type:  NODE_ELEMENT,
				Depth: 0,
				Connections: []Connection{
					{
						FromID: "source",
						FromAnchor: AnchorDescriptor{
							Raw:        "wn",
							Horizontal: -1,
							Vertical:   -1,
						},
						ToID: "sink",
						ToAnchor: AnchorDescriptor{
							Raw:        "se",
							Horizontal: 1,
							Vertical:   1,
						},
					},
				},
			},
		},
		{
			name: "type declaration remains",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "App"},
				{Type: tokenizer.COLON},
				{Type: tokenizer.IDENTIFIER, Value: "Service"},
			},
			expected: Node{
				Type:  NODE_ELEMENT,
				Depth: 0,
				Children: []Node{
					{
						Type:  NodeType("Service"),
						Text:  "App",
						Depth: 0,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.tokens)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Parse() expected error %v, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Parse() error = %v, want %v", err, tt.expectedError)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Parse() = %v, want %v", got, tt.expected)
			}
		})
	}
}
