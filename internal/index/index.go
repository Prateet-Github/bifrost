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
	idx.Remove(docID)

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

func (idx *InvertedIndex) Lookup(term string) []Posting {
	return idx.terms[term]
}

func (idx *InvertedIndex) DocumentStats(docID string) (DocumentStats, bool) {
	stats, exists := idx.documents[docID]
	return stats, exists
}

func (idx *InvertedIndex) Terms() map[string][]Posting {
	terms := make(map[string][]Posting, len(idx.terms))

	for term, postings := range idx.terms {
		terms[term] = append([]Posting(nil), postings...)
	}

	return terms
}

func (idx *InvertedIndex) Documents() map[string]DocumentStats {
	documents := make(map[string]DocumentStats, len(idx.documents))

	for docID, stats := range idx.documents {
		documents[docID] = stats
	}

	return documents
}

func (idx *InvertedIndex) AddPosting(term string, posting Posting) {
	idx.terms[term] = append(idx.terms[term], posting)
}

func (idx *InvertedIndex) SetDocumentStats(
	docID string,
	stats DocumentStats,
) {
	idx.documents[docID] = stats
}

func (idx *InvertedIndex) Remove(docID string) {
	for term, postings := range idx.terms {
		remaining := postings[:0]

		for _, posting := range postings {
			if posting.DocID != docID {
				remaining = append(remaining, posting)
			}
		}

		if len(remaining) == 0 {
			delete(idx.terms, term)
			continue
		}

		idx.terms[term] = remaining
	}

	delete(idx.documents, docID)
}
