package filter

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

func TestStopWordFilter(t *testing.T) {
	filter := NewStopWordFilter()

	tokens := []tokenizer.Token{
		{
			Text:        "the",
			Position:    0,
			StartOffset: 0,
			EndOffset:   3,
		},
		{
			Text:        "quick",
			Position:    1,
			StartOffset: 4,
			EndOffset:   9,
		},
		{
			Text:        "fox",
			Position:    2,
			StartOffset: 10,
			EndOffset:   13,
		},
		{
			Text:        "and",
			Position:    3,
			StartOffset: 14,
			EndOffset:   17,
		},
		{
			Text:        "fast",
			Position:    4,
			StartOffset: 18,
			EndOffset:   22,
		},
	}

	result := filter.Filter(tokens)

	if len(result) != 3 {
		t.Fatalf("got %d tokens, want 3", len(result))
	}

	expected := []string{"quick", "fox", "fast"}

	for i, token := range result {
		if token.Text != expected[i] {
			t.Errorf(
				"result[%d].Text = %q, want %q",
				i,
				token.Text,
				expected[i],
			)
		}
	}
}

func TestStopWordFilterPreservesMetadata(t *testing.T) {
	filter := NewStopWordFilter()

	tokens := []tokenizer.Token{
		{
			Text:        "the",
			Position:    0,
			StartOffset: 0,
			EndOffset:   3,
		},
		{
			Text:        "quick",
			Position:    1,
			StartOffset: 4,
			EndOffset:   9,
		},
	}

	result := filter.Filter(tokens)

	if len(result) != 1 {
		t.Fatalf("got %d tokens, want 1", len(result))
	}

	token := result[0]

	if token.Text != "quick" {
		t.Errorf("Text = %q, want %q", token.Text, "quick")
	}

	if token.Position != 1 {
		t.Errorf("Position = %d, want 1", token.Position)
	}

	if token.StartOffset != 4 {
		t.Errorf("StartOffset = %d, want 4", token.StartOffset)
	}

	if token.EndOffset != 9 {
		t.Errorf("EndOffset = %d, want 9", token.EndOffset)
	}
}

func TestStopWordFilterDoesNotMatchPartialWords(t *testing.T) {
	filter := NewStopWordFilter()

	tokens := []tokenizer.Token{
		{Text: "the"},
		{Text: "there"},
		{Text: "their"},
		{Text: "theory"},
	}

	result := filter.Filter(tokens)

	if len(result) != 3 {
		t.Fatalf("got %d tokens, want 3", len(result))
	}

	expected := []string{"there", "their", "theory"}

	for i, token := range result {
		if token.Text != expected[i] {
			t.Errorf(
				"result[%d].Text = %q, want %q",
				i,
				token.Text,
				expected[i],
			)
		}
	}
}

func TestStopWordFilterEmptyInput(t *testing.T) {
	filter := NewStopWordFilter()

	result := filter.Filter(nil)

	if len(result) != 0 {
		t.Fatalf("got %d tokens, want 0", len(result))
	}
}

func TestStopWordFilterAllStopWords(t *testing.T) {
	filter := NewStopWordFilter()

	tokens := []tokenizer.Token{
		{Text: "the"},
		{Text: "and"},
		{Text: "is"},
		{Text: "of"},
	}

	result := filter.Filter(tokens)

	if len(result) != 0 {
		t.Fatalf("got %d tokens, want 0", len(result))
	}
}
