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

func TestBM25PreparedQueryScoresDocumentTerms(t *testing.T) {
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
	queryScorer := bm25.PrepareQuery([]string{"go", "server"})

	doc1, exists := idx.DocumentStats("D1")
	if !exists {
		t.Fatal("expected document D1 to exist")
	}

	doc2, exists := idx.DocumentStats("D2")
	if !exists {
		t.Fatal("expected document D2 to exist")
	}

	scoreD1 := queryScorer.ScoreTerm("go", 2, doc1.Length) +
		queryScorer.ScoreTerm("server", 1, doc1.Length)

	scoreD2 := queryScorer.ScoreTerm("go", 1, doc2.Length)

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

func TestBM25TitleIDF(t *testing.T) {
	idx := index.NewInvertedIndex()

	idx.AddDocumentFields(
		"D1",
		[]tokenizer.Token{{Text: "common", Position: 0}},
		[]tokenizer.Token{{Text: "body", Position: 0}},
	)
	idx.AddDocumentFields(
		"D2",
		[]tokenizer.Token{{Text: "common", Position: 0}},
		[]tokenizer.Token{{Text: "body", Position: 0}},
	)
	idx.AddDocumentFields(
		"D3",
		[]tokenizer.Token{{Text: "rare", Position: 0}},
		[]tokenizer.Token{{Text: "body", Position: 0}},
	)

	bm25 := NewBM25(idx)

	commonIDF := bm25.TitleIDF("common")
	rareIDF := bm25.TitleIDF("rare")

	if rareIDF <= commonIDF {
		t.Errorf(
			"expected rarer title term to have higher IDF: common=%f rare=%f",
			commonIDF,
			rareIDF,
		)
	}
}

func TestBM25CombineFieldScores(t *testing.T) {
	bm25 := NewBM25(index.NewInvertedIndex())

	score := bm25.CombineFieldScores(0.5, 0.8)

	expected := 0.5 + 2.0*0.8

	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected %f, got %f", expected, score)
	}

	bm25.TitleBoost = 3.0

	score = bm25.CombineFieldScores(0.5, 0.8)
	expected = 0.5 + 3.0*0.8

	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected %f with updated boost, got %f", expected, score)
	}
}
