package tokenizer

import "strings"

// TokenType represents the type of token
type TokenType int

const (
	IDENTIFIER TokenType = iota
	LEFT_BRACE
	RIGHT_BRACE
	COLON       // For type declarations like "name:Type"
	AT          // For @state declarations
	LEFT_PAREN  // For props lists
	RIGHT_PAREN // For props lists
	COMMA       // For props lists
	EQUALS      // For props assignments
	ARROW       // For straight connection arrows like -->
	DOT         // For component property access like "browser.e"
	AMPERSAND   // For component references like "&browser.c"
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
		case '@':
			flushWord()
			tokens = append(tokens, Token{Type: AT})
		case '(':
			flushWord()
			tokens = append(tokens, Token{Type: LEFT_PAREN})
		case ')':
			flushWord()
			tokens = append(tokens, Token{Type: RIGHT_PAREN})
		case ',':
			flushWord()
			tokens = append(tokens, Token{Type: COMMA})
		case '=':
			flushWord()
			tokens = append(tokens, Token{Type: EQUALS})
		case '.':
			flushWord()
			tokens = append(tokens, Token{Type: DOT})
		case '&':
			flushWord()
			tokens = append(tokens, Token{Type: AMPERSAND})
		case '-':
			if i+2 < len(input) && input[i:i+3] == "-->" {
				flushWord()
				tokens = append(tokens, Token{Type: ARROW, Value: "-->"})
				i += 2
				continue
			}
			currentWord.WriteByte(char)
		case '\n':
			flushWord() // Force a word break on newline
		case ' ', '\t':
			flushWord() // Split on significant whitespace
		default:
			// Handle string literals
			if char == '"' || char == '\'' {
				flushWord()
				// Add the opening quote
				tokens = append(tokens, Token{Type: IDENTIFIER, Value: string(char)})
				// Get the string content
				i++
				for i < len(input) && input[i] != char {
					currentWord.WriteByte(input[i])
					i++
				}
				if currentWord.Len() > 0 {
					tokens = append(tokens, Token{Type: IDENTIFIER, Value: currentWord.String()})
					currentWord.Reset()
				}
				// Add the closing quote if we found it
				if i < len(input) {
					tokens = append(tokens, Token{Type: IDENTIFIER, Value: string(char)})
				}
				continue
			}
			currentWord.WriteByte(char)
		}
	}

	flushWord() // Flush any remaining word
	return tokens
}
