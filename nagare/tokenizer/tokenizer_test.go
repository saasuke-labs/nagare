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
			name:  "empty input",
			input: "",
			expected: []Token{
				{Type: IDENTIFIER, Value: ""},
			},
		},
		{
			name:  "with whitespace",
			input: "  Server  ",
			expected: []Token{
				{Type: IDENTIFIER, Value: "  Server  "},
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
