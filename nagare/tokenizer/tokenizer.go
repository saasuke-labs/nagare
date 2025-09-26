package tokenizer

import "strings"

// TokenType represents the type of token
type TokenType int

const (
	IDENTIFIER TokenType = iota
	LEFT_BRACE
	RIGHT_BRACE
	COLON // For type declarations like "name:Type"
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

	var tokens []Token
	var currentWord strings.Builder

	// Helper to add word token if there's any
	flushWord := func() {
		if currentWord.Len() > 0 {
			tokens = append(tokens, Token{
				Type:  IDENTIFIER,
				Value: strings.TrimSpace(currentWord.String()),
			})
			currentWord.Reset()
		}
	}

	for i := 0; i < len(input); i++ {
		char := input[i]
		switch char {
		case '{':
			flushWord()
			tokens = append(tokens, Token{Type: LEFT_BRACE})
		case '}':
			flushWord()
			tokens = append(tokens, Token{Type: RIGHT_BRACE})
		case ':':
			flushWord()
			tokens = append(tokens, Token{Type: COLON})
		case '\n':
			flushWord() // Force a word break on newline
		case ' ', '\t':
			flushWord() // Split on significant whitespace
		default:
			currentWord.WriteByte(char)
		}
	}

	flushWord() // Flush any remaining word
	return tokens
}
