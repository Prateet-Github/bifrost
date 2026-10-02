package index

import "github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"

type Posting struct {
	DocID     string
	TermFreq  int
	Positions []int
}

type DocumentStats struct {
	Length int
}

type InvertedIndex struct {
	terms     map[string][]Posting
	documents map[string]DocumentStats
}

func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		terms:     make(map[string][]Posting),
		documents: make(map[string]DocumentStats),
	}
}

func (idx *InvertedIndex) AddDocument(
	docID string,
	tokens []tokenizer.Token,
) {
	idx.documents[docID] = DocumentStats{
		Length: len(tokens),
	}

	for _, token := range tokens {
		postings := idx.terms[token.Text]

		if len(postings) == 0 || postings[len(postings)-1].DocID != docID {
			idx.terms[token.Text] = append(postings, Posting{
				DocID:     docID,
				TermFreq:  1,
				Positions: []int{token.Position},
			})
			continue
		}

		posting := &postings[len(postings)-1]
		posting.TermFreq++
		posting.Positions = append(posting.Positions, token.Position)
	}
}
