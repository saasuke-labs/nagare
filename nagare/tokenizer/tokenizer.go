package tokenizer

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
	// For now, we just treat the entire input as a single identifier
	return []Token{
		{
			Type:  IDENTIFIER,
			Value: input,
		},
	}
}
