package index

import "github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"

type Posting struct {
	DocID     string
	TermFreq  int
	Positions []int
}

type DocumentStats struct {
	Length      int
	TitleLength int
}

type InvertedIndex struct {
	terms               map[string][]Posting
	titleTerms          map[string][]Posting
	documents           map[string]DocumentStats
	totalDocumentLength int
}

func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		terms:      make(map[string][]Posting),
		titleTerms: make(map[string][]Posting),
		documents:  make(map[string]DocumentStats),
	}
}

func (idx *InvertedIndex) AddDocument(
	docID string,
	tokens []tokenizer.Token,
) {
	idx.AddDocumentFields(docID, nil, tokens)
}

func (idx *InvertedIndex) AddDocumentFields(
	docID string,
	titleTokens []tokenizer.Token,
	bodyTokens []tokenizer.Token,
) {
	idx.Remove(docID)

	idx.documents[docID] = DocumentStats{
		Length:      len(bodyTokens),
		TitleLength: len(titleTokens),
	}

	idx.totalDocumentLength += len(bodyTokens)

	addTokens := func(
		terms map[string][]Posting,
		tokens []tokenizer.Token,
	) {
		for _, token := range tokens {
			postings := terms[token.Text]

			if len(postings) == 0 ||
				postings[len(postings)-1].DocID != docID {
				terms[token.Text] = append(postings, Posting{
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

	addTokens(idx.titleTerms, titleTokens)
	addTokens(idx.terms, bodyTokens)
}

func (idx *InvertedIndex) Lookup(term string) []Posting {
	return idx.terms[term]
}

func (idx *InvertedIndex) LookupTitle(term string) []Posting {
	return idx.titleTerms[term]
}

func (idx *InvertedIndex) TitleTerms() map[string][]Posting {
	terms := make(map[string][]Posting, len(idx.titleTerms))

	for term, postings := range idx.titleTerms {
		copied := make([]Posting, len(postings))

		for i, posting := range postings {
			copied[i] = Posting{
				DocID:     posting.DocID,
				TermFreq:  posting.TermFreq,
				Positions: append([]int(nil), posting.Positions...),
			}
		}

		terms[term] = copied
	}

	return terms
}

func (idx *InvertedIndex) DocumentStats(docID string) (DocumentStats, bool) {
	stats, exists := idx.documents[docID]
	return stats, exists
}

func (idx *InvertedIndex) Terms() map[string][]Posting {
	terms := make(map[string][]Posting, len(idx.terms))

	for term, postings := range idx.terms {
		copied := make([]Posting, len(postings))

		for i, posting := range postings {
			copied[i] = Posting{
				DocID:     posting.DocID,
				TermFreq:  posting.TermFreq,
				Positions: append([]int(nil), posting.Positions...),
			}
		}

		terms[term] = copied
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

func (idx *InvertedIndex) AddTitlePosting(
	term string,
	posting Posting,
) {
	idx.titleTerms[term] = append(
		idx.titleTerms[term],
		posting,
	)
}

func (idx *InvertedIndex) SetDocumentStats(
	docID string,
	stats DocumentStats,
) {
	if existing, exists := idx.documents[docID]; exists {
		idx.totalDocumentLength -= existing.Length
	}

	idx.documents[docID] = stats
	idx.totalDocumentLength += stats.Length
}

func (idx *InvertedIndex) Remove(docID string) {
	doc, exists := idx.documents[docID]
	if !exists {
		return
	}

	idx.totalDocumentLength -= doc.Length

	removePostings := func(terms map[string][]Posting) {
		for term, postings := range terms {
			remaining := postings[:0]

			for _, posting := range postings {
				if posting.DocID != docID {
					remaining = append(remaining, posting)
				}
			}

			if len(remaining) == 0 {
				delete(terms, term)
				continue
			}

			terms[term] = remaining
		}
	}

	removePostings(idx.terms)
	removePostings(idx.titleTerms)

	delete(idx.documents, docID)
}
