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

func BenchmarkAnalyzer(b *testing.B) {
	analyzer := NewAnalyzer()

	doc := document.Document{
		ID:    "bench-1",
		Title: "Benchmark Document",
		Body: `
			The quick brown fox is running through the forest.
			Developers are building high performance applications
			using Go, distributed systems, databases, networking,
			and modern search technologies. Bifrost is a search
			engine built from scratch to understand how information
			retrieval systems work internally.
		`,
	}

	b.ReportAllocs()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = analyzer.Analyze(doc)
	}
}
