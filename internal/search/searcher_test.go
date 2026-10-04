package search

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
	"github.com/Prateet-Github/bifrost/internal/index"
)

func TestCandidates(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "fast", Position: 1},
		{Text: "go", Position: 2},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "server", Position: 1},
	})

	idx.AddDocument("D3", []tokenizer.Token{
		{Text: "rust", Position: 0},
		{Text: "fast", Position: 1},
	})

	searcher := NewSearcher(idx)

	candidates := searcher.Candidates("go fast")

	if len(candidates) != 3 {
		t.Fatalf(
			"expected 3 candidates, got %d",
			len(candidates),
		)
	}

	byDoc := make(map[string]Candidate)

	for _, candidate := range candidates {
		byDoc[candidate.DocID] = candidate
	}

	// D1 should match both "go" and "fast"
	d1, ok := byDoc["D1"]
	if !ok {
		t.Fatal("expected D1 to be a candidate")
	}

	if len(d1.Matches) != 2 {
		t.Fatalf(
			"expected D1 to have 2 matches, got %d",
			len(d1.Matches),
		)
	}

	// D2 should only match "go"
	d2, ok := byDoc["D2"]
	if !ok {
		t.Fatal("expected D2 to be a candidate")
	}

	if len(d2.Matches) != 1 {
		t.Fatalf(
			"expected D2 to have 1 match, got %d",
			len(d2.Matches),
		)
	}

	if d2.Matches[0].Term != "go" {
		t.Errorf(
			"expected D2 match to be 'go', got %q",
			d2.Matches[0].Term,
		)
	}

	// D3 should only match "fast"
	d3, ok := byDoc["D3"]
	if !ok {
		t.Fatal("expected D3 to be a candidate")
	}

	if len(d3.Matches) != 1 {
		t.Fatalf(
			"expected D3 to have 1 match, got %d",
			len(d3.Matches),
		)
	}

	if d3.Matches[0].Term != "fast" {
		t.Errorf(
			"expected D3 match to be 'fast', got %q",
			d3.Matches[0].Term,
		)
	}
}

func TestSearchRanksResults(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "go", Position: 1},
		{Text: "server", Position: 2},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	idx.AddDocument("D3", []tokenizer.Token{
		{Text: "server", Position: 0},
	})

	searcher := NewSearcher(idx)

	results := searcher.Search("go server", 3)

	if len(results) != 3 {
		t.Fatalf(
			"expected 3 results, got %d",
			len(results),
		)
	}

	if results[0].DocID != "D1" {
		t.Errorf(
			"expected D1 to rank first, got %s",
			results[0].DocID,
		)
	}

	if results[0].Score <= results[1].Score {
		t.Errorf(
			"expected first result to have higher score: %f <= %f",
			results[0].Score,
			results[1].Score,
		)
	}
}

func TestSearchTopK(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	idx.AddDocument("D3", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	searcher := NewSearcher(idx)

	results := searcher.Search("go", 2)

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}
}

func TestSearchTopKGreaterThanResults(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	searcher := NewSearcher(idx)

	results := searcher.Search("go", 10)

	if len(results) != 1 {
		t.Fatalf(
			"expected 1 result, got %d",
			len(results),
		)
	}
}

func TestSearchZeroK(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	searcher := NewSearcher(idx)

	results := searcher.Search("go", 0)

	if len(results) != 0 {
		t.Fatalf(
			"expected 0 results, got %d",
			len(results),
		)
	}
}
