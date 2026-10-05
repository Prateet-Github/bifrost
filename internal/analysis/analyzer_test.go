package analysis

import (
	"testing"
)

func TestAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()

	tokens := analyzer.Analyze(
		"The developers are programming in Go",
	)

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

	text := `
		The quick brown fox is running through the forest.
		Developers are building high performance applications
		using Go, distributed systems, databases, networking,
		and modern search technologies. Bifrost is a search
		engine built from scratch to understand how information
		retrieval systems work internally.
	`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = analyzer.Analyze(text)
	}
}

func TestAnalyzer_PreservesPositionsAfterStopWordRemoval(t *testing.T) {
	analyzer := NewAnalyzer()

	tokens := analyzer.Analyze("cat in the hat")

	if len(tokens) != 2 {
		t.Fatalf("got %d tokens, want 2", len(tokens))
	}

	if tokens[0].Text != "cat" {
		t.Errorf("tokens[0].Text = %q, want %q", tokens[0].Text, "cat")
	}

	if tokens[0].Position != 0 {
		t.Errorf("tokens[0].Position = %d, want 0", tokens[0].Position)
	}

	if tokens[1].Text != "hat" {
		t.Errorf("tokens[1].Text = %q, want %q", tokens[1].Text, "hat")
	}

	if tokens[1].Position != 3 {
		t.Errorf("tokens[1].Position = %d, want 3", tokens[1].Position)
	}
}
