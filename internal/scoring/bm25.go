package scoring

import (
	"math"

	"github.com/Prateet-Github/bifrost/internal/index"
)

type BM25 struct {
	K1         float64
	B          float64
	TitleBoost float64
	index      *index.InvertedIndex
}

type QueryScorer struct {
	bm25         *BM25
	avgDocLength float64
	idfCache     map[string]float64
}

func NewBM25(idx *index.InvertedIndex) *BM25 {
	return &BM25{
		K1:         1.2,
		B:          0.75,
		TitleBoost: 2.0,
		index:      idx,
	}
}

func (b *BM25) TitleIDF(term string) float64 {
	stats := b.index.TitleStats()
	df := b.index.TitleDocumentFrequency(term)

	if df == 0 {
		return 0
	}

	return math.Log(
		1 + (float64(stats.Documents)-float64(df)+0.5)/
			(float64(df)+0.5),
	)
}

func (b *BM25) TitleTermScore(
	term string,
	tf int,
	titleLength int,
) float64 {
	if tf == 0 {
		return 0
	}

	stats := b.index.TitleStats()

	if stats.AverageDocLength == 0 {
		return 0
	}

	idf := b.TitleIDF(term)

	tfScore := b.TermFrequencyScore(
		tf,
		titleLength,
		stats.AverageDocLength,
	)

	return idf * tfScore
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

func (b *BM25) PrepareTitleQuery(terms []string) *QueryScorer {
	stats := b.index.TitleStats()

	scorer := &QueryScorer{
		bm25:         b,
		avgDocLength: stats.AverageDocLength,
		idfCache:     make(map[string]float64, len(terms)),
	}

	for _, term := range terms {
		if _, exists := scorer.idfCache[term]; exists {
			continue
		}

		df := b.index.TitleDocumentFrequency(term)

		if df == 0 {
			scorer.idfCache[term] = 0
			continue
		}

		scorer.idfCache[term] = math.Log(
			1 + (float64(stats.Documents)-float64(df)+0.5)/
				(float64(df)+0.5),
		)
	}

	return scorer
}

func (b *BM25) CombineFieldScores(
	bodyScore float64,
	titleScore float64,
) float64 {
	return bodyScore + b.TitleBoost*titleScore
}
