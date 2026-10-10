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

func NewBM25(idx *index.InvertedIndex) *BM25 {
	return &BM25{
		K1:    1.2,
		B:     0.75,
		index: idx,
	}
}

func (b *BM25) IDF(term string) float64 {
	stats := b.index.Stats()

	return b.idf(
		term,
		stats.Documents,
	)
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

func (b *BM25) Score(
	terms []string,
	docID string,
) float64 {
	doc, exists := b.index.DocumentStats(docID)
	if !exists {
		return 0
	}

	stats := b.index.Stats()

	idfCache := make(map[string]float64, len(terms))

	var score float64

	for _, term := range terms {
		idf, cached := idfCache[term]

		if !cached {
			idf = b.idf(term, stats.Documents)
			idfCache[term] = idf
		}

		postings := b.index.Lookup(term)

		for _, posting := range postings {
			if posting.DocID == docID {
				tfScore := b.TermFrequencyScore(
					posting.TermFreq,
					doc.Length,
					stats.AverageDocLength,
				)

				score += idf * tfScore
				break
			}
		}
	}

	return score
}
