package tokenizer

import "strings"

// TokenType represents the type of token
type TokenType int

const (
	IDENTIFIER TokenType = iota
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
}

// Tokenize converts input text into a sequence of tokens
func Tokenize(input string) []Token {
	// Handle empty input
	input = strings.TrimSpace(input)
	if input == "" {
		return []Token{}
	}

	// Split input by newlines and create a token for each non-empty line
	lines := strings.Split(input, "\n")
	var tokens []Token

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			tokens = append(tokens, Token{
				Type:  IDENTIFIER,
				Value: line,
			})
		}
	}

	return tokens
}
