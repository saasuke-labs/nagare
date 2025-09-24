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
