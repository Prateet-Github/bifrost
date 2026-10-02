package query

import "testing"

func TestAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()

	tokens := analyzer.Analyze(
		"The developers are building Go servers",
	)

	expected := []string{
		"develop",
		"build",
		"go",
		"server",
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
