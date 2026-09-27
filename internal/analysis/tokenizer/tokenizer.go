package tokenizer

import "unicode"

type Token struct {
	Text        string
	Position    int
	StartOffset int
	EndOffset   int
}

type Tokenizer struct{}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) Tokenize(text string) []Token {
	tokens := make([]Token, 0)

	start := -1
	position := 0

	for offset, r := range text {
		if unicode.IsSpace(r) {
			if start != -1 {
				tokens = append(tokens, Token{
					Text:        text[start:offset],
					Position:    position,
					StartOffset: start,
					EndOffset:   offset,
				})

				position++
				start = -1
			}

			continue
		}

		if start == -1 {
			start = offset
		}
	}

	if start != -1 {
		tokens = append(tokens, Token{
			Text:        text[start:],
			Position:    position,
			StartOffset: start,
			EndOffset:   len(text),
		})
	}

	return tokens
}
