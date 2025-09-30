package tokenizer

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single word",
			input: "Server",
			expected: []Token{
				{Type: IDENTIFIER, Value: "Server"},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []Token{},
		},
		{
			name:  "multiple lines",
			input: "Server1\nServer2\nServer3",
			expected: []Token{
				{Type: IDENTIFIER, Value: "Server1"},
				{Type: IDENTIFIER, Value: "Server2"},
				{Type: IDENTIFIER, Value: "Server3"},
			},
		},
		{
			name:  "with empty lines",
			input: "Server1\n\nServer2",
			expected: []Token{
				{Type: IDENTIFIER, Value: "Server1"},
				{Type: IDENTIFIER, Value: "Server2"},
			},
		},
		{
			name:  "nested structure",
			input: "Browser\nVM {\n    nginx\n    app\n}",
			expected: []Token{
				{Type: IDENTIFIER, Value: "Browser"},
				{Type: IDENTIFIER, Value: "VM"},
				{Type: LEFT_BRACE},
				{Type: IDENTIFIER, Value: "nginx"},
				{Type: IDENTIFIER, Value: "app"},
				{Type: RIGHT_BRACE},
			},
		},
		{
			name:  "connection arrow",
			input: "foo:w --> bar:e",
			expected: []Token{
				{Type: IDENTIFIER, Value: "foo"},
				{Type: COLON},
				{Type: IDENTIFIER, Value: "w"},
				{Type: ARROW, Value: "-->"},
				{Type: IDENTIFIER, Value: "bar"},
				{Type: COLON},
				{Type: IDENTIFIER, Value: "e"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tokenize(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Tokenize() = %v, want %v", got, tt.expected)
			}
		})
	}
}
