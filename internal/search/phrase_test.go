package search

import (
	"testing"

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
