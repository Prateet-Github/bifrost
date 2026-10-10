package index

type Stats struct {
	Documents        int
	AverageDocLength float64
}

func (idx *InvertedIndex) DocumentFrequency(term string) int {
	return len(idx.terms[term])
}

func (idx *InvertedIndex) Stats() Stats {
	documentCount := len(idx.documents)

	if documentCount == 0 {
		return Stats{}
	}

	return Stats{
		Documents: documentCount,
		AverageDocLength: float64(idx.totalDocumentLength) /
			float64(documentCount),
	}
}
