package scoring

import (
	"math"

	"github.com/Prateet-Github/bifrost/internal/index"
)

type BM25 struct {
	K1    float64
	B     float64
	index *index.InvertedIndex
}

type QueryScorer struct {
	bm25         *BM25
	avgDocLength float64
	idfCache     map[string]float64
}

func NewBM25(idx *index.InvertedIndex) *BM25 {
	return &BM25{
		K1:    1.2,
		B:     0.75,
		index: idx,
	}
}

func (b *BM25) IDF(term string) float64 {
	stats := b.index.Stats()

	return b.idf(term, stats.Documents)
}

func (b *BM25) idf(term string, totalDocuments int) float64 {
	df := b.index.DocumentFrequency(term)

	return math.Log(
		1 + (float64(totalDocuments)-float64(df)+0.5)/
			(float64(df)+0.5),
	)
}

func (b *BM25) TermFrequencyScore(
	tf int,
	docLength int,
	avgDocLength float64,
) float64 {
	if tf == 0 || avgDocLength == 0 {
		return 0
	}

	tfValue := float64(tf)
	dl := float64(docLength)

	normalization := 1 - b.B + b.B*(dl/avgDocLength)

	return (tfValue * (b.K1 + 1)) /
		(tfValue + b.K1*normalization)
}

func (b *BM25) TermScore(
	term string,
	tf int,
	docLength int,
) float64 {
	stats := b.index.Stats()

	idf := b.idf(term, stats.Documents)

	tfScore := b.TermFrequencyScore(
		tf,
		docLength,
		stats.AverageDocLength,
	)

	return idf * tfScore
}

// PrepareQuery calculates the scoring data that can be reused
// across all candidate documents for a query.
func (b *BM25) PrepareQuery(terms []string) *QueryScorer {
	stats := b.index.Stats()

	scorer := &QueryScorer{
		bm25:         b,
		avgDocLength: stats.AverageDocLength,
		idfCache:     make(map[string]float64, len(terms)),
	}

	for _, term := range terms {
		if _, exists := scorer.idfCache[term]; exists {
			continue
		}

		scorer.idfCache[term] = b.idf(
			term,
			stats.Documents,
		)
	}

	return scorer
}

// ScoreTerm scores a term using the prepared query-scoring data.
func (q *QueryScorer) ScoreTerm(
	term string,
	tf int,
	docLength int,
) float64 {
	if tf == 0 {
		return 0
	}

	idf, exists := q.idfCache[term]
	if !exists {
		return 0
	}

	tfScore := q.bm25.TermFrequencyScore(
		tf,
		docLength,
		q.avgDocLength,
	)

	return idf * tfScore
}

// Score retains the existing API for callers that score a document directly.
func (b *BM25) Score(
	terms []string,
	docID string,
) float64 {
	doc, exists := b.index.DocumentStats(docID)
	if !exists {
		return 0
	}

	queryScorer := b.PrepareQuery(terms)

	var score float64

	for _, term := range terms {
		postings := b.index.Lookup(term)

		for _, posting := range postings {
			if posting.DocID == docID {
				score += queryScorer.ScoreTerm(
					term,
					posting.TermFreq,
					doc.Length,
				)
				break
			}
		}
	}

	return score
}
