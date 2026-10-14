package search

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
	"github.com/Prateet-Github/bifrost/internal/index"
)

func TestPositionsForTerm(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "go", Position: 1},
		{Text: "server", Position: 2},
		{Text: "go", Position: 3},
	})

	positions := positionsForTerm(
		idx,
		"go",
		"D1",
		false,
	)

	expected := []int{0, 1, 3}

	if len(positions) != len(expected) {
		t.Fatalf(
			"expected %d positions, got %d",
			len(expected),
			len(positions),
		)
	}

	for i := range expected {
		if positions[i] != expected[i] {
			t.Errorf(
				"expected position %d at index %d, got %d",
				expected[i],
				i,
				positions[i],
			)
		}
	}
}

func TestPositionsMatch(t *testing.T) {
	tests := []struct {
		name   string
		first  []int
		second []int
		want   bool
	}{
		{
			name:   "adjacent positions",
			first:  []int{2},
			second: []int{3},
			want:   true,
		},
		{
			name:   "not adjacent",
			first:  []int{2},
			second: []int{5},
			want:   false,
		},
		{
			name:   "one matching position",
			first:  []int{2, 10, 20},
			second: []int{3, 15, 21},
			want:   true,
		},
		{
			name:   "no matching position",
			first:  []int{2, 10, 20},
			second: []int{5, 15, 30},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := positionsMatch(
				tt.first,
				tt.second,
			)

			if got != tt.want {
				t.Errorf(
					"expected %v, got %v",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestPhraseMatch_DoesNotCollapseStopWordPositions(t *testing.T) {
	analyzer := analysis.NewAnalyzer()
	idx := index.NewInvertedIndex()

	tokens := analyzer.Analyze("cat in the hat")

	idx.AddDocument("D1", tokens)

	searcher := NewSearcher(idx)

	if searcher.PhraseMatch("cat hat", "D1") {
		t.Fatal(`"cat hat" should not match "cat in the hat"`)
	}
}

func TestPhraseMatch_Generalized(t *testing.T) {

	analyzer := analysis.NewAnalyzer()
	idx := index.NewInvertedIndex()
	tokens := analyzer.Analyze(
		"distributed systems are fascinating because distributed systems",
	)
	idx.AddDocument("D1", tokens)
	searcher := NewSearcher(idx)

	tests := []struct {
		name   string
		phrase string
		docID  string
		want   bool
	}{
		{
			name:   "single term",
			phrase: "distributed",
			docID:  "D1",
			want:   true,
		},
		{
			name:   "two consecutive terms",
			phrase: "distributed systems",
			docID:  "D1",
			want:   true,
		},
		{
			name:   "three consecutive terms",
			phrase: "distributed systems distributed",
			docID:  "D1",
			want:   false,
		},
		{
			name:   "longer consecutive phrase",
			phrase: "distributed systems are fascinating",
			docID:  "D1",
			want:   false,
		},
		{
			name:   "terms in wrong order",
			phrase: "systems distributed",
			docID:  "D1",
			want:   false,
		},
		{
			name:   "terms are not consecutive",
			phrase: "distributed fascinating",
			docID:  "D1",
			want:   false,
		},
		{
			name:   "repeated term phrase",
			phrase: "systems systems",
			docID:  "D1",
			want:   false,
		},
		{
			name:   "missing term",
			phrase: "distributed databases",
			docID:  "D1",
			want:   false,
		},
		{
			name:   "missing document",
			phrase: "distributed systems",
			docID:  "D2",
			want:   false,
		},
		{
			name:   "empty phrase",
			phrase: "",
			docID:  "D1",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searcher.PhraseMatch(tt.phrase, tt.docID)

			if got != tt.want {
				t.Errorf(
					"PhraseMatch(%q, %q) = %v, want %v",
					tt.phrase,
					tt.docID,
					got,
					tt.want,
				)
			}
		})
	}
}
