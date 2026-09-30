package index

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
