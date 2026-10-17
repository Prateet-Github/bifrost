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
			name:  "punctuation trimmed",
			input: "Hello, world!",
			expected: []Token{
				{Text: "Hello", Position: 0, StartOffset: 0, EndOffset: 6},
				{Text: "world", Position: 1, StartOffset: 7, EndOffset: 13},
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

		{
			name:  "leading and trailing punctuation trimmed",
			input: "Hello, world!",
			expected: []Token{
				{Text: "Hello", Position: 0, StartOffset: 0, EndOffset: 6},
				{Text: "world", Position: 1, StartOffset: 7, EndOffset: 13},
			},
		},
		{
			name:  "inner punctuation preserved",
			input: "Node.js don't",
			expected: []Token{
				{Text: "Node.js", Position: 0, StartOffset: 0, EndOffset: 7},
				{Text: "don't", Position: 1, StartOffset: 8, EndOffset: 13},
			},
		},
		{
			name:  "allowlisted technical terms",
			input: "C++ C# .NET",
			expected: []Token{
				{Text: "C++", Position: 0, StartOffset: 0, EndOffset: 3},
				{Text: "C#", Position: 1, StartOffset: 4, EndOffset: 6},
				{Text: ".NET", Position: 2, StartOffset: 7, EndOffset: 11},
			},
		},
		{
			name:  "dash slash and hyphen separators",
			input: "alpha—beta gamma–delta path/name distributed-systems",
			expected: []Token{
				{Text: "alpha", Position: 0, StartOffset: 0, EndOffset: 5},
				{Text: "beta", Position: 1, StartOffset: 8, EndOffset: 12},
				{Text: "gamma", Position: 2, StartOffset: 13, EndOffset: 18},
				{Text: "delta", Position: 3, StartOffset: 21, EndOffset: 26},
				{Text: "path", Position: 4, StartOffset: 27, EndOffset: 31},
				{Text: "name", Position: 5, StartOffset: 32, EndOffset: 36},
				{Text: "distributed", Position: 6, StartOffset: 37, EndOffset: 48},
				{Text: "systems", Position: 7, StartOffset: 49, EndOffset: 56},
			},
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
