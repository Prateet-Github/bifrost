package scoring

import (
	"math"
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
	"github.com/Prateet-Github/bifrost/internal/index"
)

func TestBM25IDF(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	idx.AddDocument("D3", []tokenizer.Token{
		{Text: "rust", Position: 0},
	})

	bm25 := NewBM25(idx)

	idfGo := bm25.IDF("go")
	idfRust := bm25.IDF("rust")

	if idfRust <= idfGo {
		t.Errorf(
			"expected rarer term to have higher IDF: go=%f rust=%f",
			idfGo,
			idfRust,
		)
	}
}

func TestBM25TermFrequencyScore(t *testing.T) {
	idx := index.NewInvertedIndex()
	bm25 := NewBM25(idx)

	score := bm25.TermFrequencyScore(
		3,
		100,
		100,
	)

	expected := (3.0 * 2.2) / (3.0 + 1.2)

	if math.Abs(score-expected) > 1e-9 {
		t.Errorf(
			"expected %f, got %f",
			expected,
			score,
		)
	}
}

func TestBM25TermFrequencySaturation(t *testing.T) {
	idx := index.NewInvertedIndex()
	bm25 := NewBM25(idx)

	score1 := bm25.TermFrequencyScore(1, 100, 100)
	score2 := bm25.TermFrequencyScore(2, 100, 100)
	score10 := bm25.TermFrequencyScore(10, 100, 100)

	if score2 <= score1 {
		t.Error("expected TF=2 to score higher than TF=1")
	}

	if score10 <= score2 {
		t.Error("expected TF=10 to score higher than TF=2")
	}

	if score10 >= 2*score2 {
		t.Error("expected BM25 TF contribution to saturate")
	}
}

func TestBM25TermScore(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "go", Position: 1},
		{Text: "server", Position: 2},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "rust", Position: 0},
	})

	bm25 := NewBM25(idx)

	score := bm25.TermScore(
		"go",
		2,
		3,
	)

	if score <= 0 {
		t.Fatalf("expected positive score, got %f", score)
	}
}

func TestBM25RareTermScoresHigher(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "common", Position: 0},
		{Text: "rare", Position: 1},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "common", Position: 0},
	})

	bm25 := NewBM25(idx)

	commonScore := bm25.TermScore("common", 1, 2)
	rareScore := bm25.TermScore("rare", 1, 2)

	if rareScore <= commonScore {
		t.Errorf(
			"expected rare term to score higher: common=%f rare=%f",
			commonScore,
			rareScore,
		)
	}
}

func TestBM25Score(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "go", Position: 1},
		{Text: "server", Position: 2},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	bm25 := NewBM25(idx)

	scoreD1 := bm25.Score(
		[]string{"go", "server"},
		"D1",
	)

	scoreD2 := bm25.Score(
		[]string{"go", "server"},
		"D2",
	)

	if scoreD1 <= 0 {
		t.Fatalf("expected positive score for D1, got %f", scoreD1)
	}

	if scoreD2 <= 0 {
		t.Fatalf("expected positive score for D2, got %f", scoreD2)
	}

	if scoreD1 <= scoreD2 {
		t.Errorf(
			"expected D1 to score higher: D1=%f D2=%f",
			scoreD1,
			scoreD2,
		)
	}
}
