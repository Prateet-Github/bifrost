package tokenizer

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tokenizer := NewTokenizer()

	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "basic text",
			input: "hello world",
			expected: []Token{
				{Text: "hello", Position: 0, StartOffset: 0, EndOffset: 5},
				{Text: "world", Position: 1, StartOffset: 6, EndOffset: 11},
			},
		},
		{
			name:  "multiple spaces",
			input: "hello   world",
			expected: []Token{
				{Text: "hello", Position: 0, StartOffset: 0, EndOffset: 5},
				{Text: "world", Position: 1, StartOffset: 8, EndOffset: 13},
			},
		},
		{
			name:  "technical terms",
			input: "C++ Node.js",
			expected: []Token{
				{Text: "C++", Position: 0, StartOffset: 0, EndOffset: 3},
				{Text: "Node.js", Position: 1, StartOffset: 4, EndOffset: 11},
			},
		},
		{
			name:  "punctuation preserved",
			input: "Hello, world!",
			expected: []Token{
				{Text: "Hello,", Position: 0, StartOffset: 0, EndOffset: 6},
				{Text: "world!", Position: 1, StartOffset: 7, EndOffset: 13},
			},
		},
		{
			name:  "unicode",
			input: "café résumé",
			expected: []Token{
				{Text: "café", Position: 0, StartOffset: 0, EndOffset: 5},
				{Text: "résumé", Position: 1, StartOffset: 6, EndOffset: 14},
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []Token{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: []Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.Tokenize(tt.input)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, got)
			}
		})
	}
}
