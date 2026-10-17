package tokenizer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

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
	position := 0

	start := -1

	for offset, r := range text {
		if isSeparator(r) {
			if start != -1 {
				tokens = appendToken(tokens, text, start, offset, position)
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
		tokens = appendToken(tokens, text, start, len(text), position)
	}

	return tokens
}

func appendToken(tokens []Token, text string, start, end, position int) []Token {
	raw := text[start:end]
	term := strings.TrimFunc(raw, unicode.IsPunct)
	term = strings.ReplaceAll(term, "\x00", "")

	if term == "" {
		return tokens
	}

	switch strings.ToLower(raw) {
	case "c++", "c#", ".net":
		term = raw
	}

	startOffset := start
	endOffset := end

	switch strings.ToLower(raw) {
	case "c++", "c#", ".net":
		startOffset = start
		endOffset = end
	}

	if !utf8.ValidString(term) {
		return tokens
	}

	return append(tokens, Token{
		Text:        term,
		Position:    position,
		StartOffset: startOffset,
		EndOffset:   endOffset,
	})
}

func isSeparator(r rune) bool {
	switch r {
	case '/', '\\', '-', '—', '–', '\x00':
		return true
	default:
		return unicode.IsSpace(r)
	}
}
