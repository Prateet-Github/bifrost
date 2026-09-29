package normalizer

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "Go",
			expected: "go",
		},
		{
			name:     "uppercase",
			input:    "GO",
			expected: "go",
		},
		{
			name:     "mixed case",
			input:    "GoLang",
			expected: "golang",
		},
		{
			name:     "accent folding",
			input:    "Café",
			expected: "cafe",
		},
		{
			name:     "uppercase accent folding",
			input:    "RÉSUMÉ",
			expected: "resume",
		},
		{
			name:     "diaeresis",
			input:    "Naïve",
			expected: "naive",
		},
		{
			name:     "technical term C++",
			input:    "C++",
			expected: "c++",
		},
		{
			name:     "technical term C#",
			input:    "C#",
			expected: "c#",
		},
		{
			name:     "technical term Node.js",
			input:    "Node.js",
			expected: "node.js",
		},
		{
			name:     "technical term HTTP/2",
			input:    "HTTP/2",
			expected: "http/2",
		},
	}

	normalizer := NewNormalizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tokenizer.Token{
				Text: tt.input,
			}

			result := normalizer.Normalize(token)

			if result.Text != tt.expected {
				t.Errorf(
					"Normalize(%q) = %q, want %q",
					tt.input,
					result.Text,
					tt.expected,
				)
			}
		})
	}
}

func TestNormalizePreservesMetadata(t *testing.T) {
	normalizer := NewNormalizer()

	token := tokenizer.Token{
		Text:        "Café",
		Position:    5,
		StartOffset: 42,
		EndOffset:   47,
	}

	result := normalizer.Normalize(token)

	if result.Text != "cafe" {
		t.Errorf("Text = %q, want %q", result.Text, "cafe")
	}

	if result.Position != token.Position {
		t.Errorf(
			"Position = %d, want %d",
			result.Position,
			token.Position,
		)
	}

	if result.StartOffset != token.StartOffset {
		t.Errorf(
			"StartOffset = %d, want %d",
			result.StartOffset,
			token.StartOffset,
		)
	}

	if result.EndOffset != token.EndOffset {
		t.Errorf(
			"EndOffset = %d, want %d",
			result.EndOffset,
			token.EndOffset,
		)
	}
}
