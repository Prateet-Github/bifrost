package analysis

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/document"
)

func TestAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()

	doc := document.Document{
		ID:    "doc-1",
		Title: "Test",
		Body:  "The developers are programming in Go",
	}

	tokens := analyzer.Analyze(doc)

	expected := []string{
		"develop",
		"program",
		"go",
	}

	if len(tokens) != len(expected) {
		t.Fatalf(
			"got %d tokens, want %d",
			len(tokens),
			len(expected),
		)
	}

	for i, token := range tokens {
		if token.Text != expected[i] {
			t.Errorf(
				"tokens[%d] = %q, want %q",
				i,
				token.Text,
				expected[i],
			)
		}
	}
}
