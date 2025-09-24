package parser

import (
	"nagare/tokenizer"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []tokenizer.Token
		expected Node
	}{
		{
			name: "single identifier",
			tokens: []tokenizer.Token{
				{Type: tokenizer.IDENTIFIER, Value: "Server"},
			},
			expected: Node{
				Type: NODE_ELEMENT,
				Text: "Server",
			},
		},
		{
			name:     "empty tokens",
			tokens:   []tokenizer.Token{},
			expected: Node{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.tokens)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Parse() = %v, want %v", got, tt.expected)
			}
		})
	}
}
