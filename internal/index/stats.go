package index

type Stats struct {
	Documents        int
	AverageDocLength float64
}

type FieldStats struct {
	Documents        int
	AverageDocLength float64
}

func (idx *InvertedIndex) DocumentFrequency(term string) int {
	return len(idx.terms[term])
}

func (idx *InvertedIndex) TitleDocumentFrequency(term string) int {
	return len(idx.titleTerms[term])
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

func (idx *InvertedIndex) TitleStats() FieldStats {
	documentCount := len(idx.documents)

	if documentCount == 0 {
		return FieldStats{}
	}

	var totalTitleLength int

	for _, stats := range idx.documents {
		totalTitleLength += stats.TitleLength
	}

	return FieldStats{
		Documents: documentCount,
		AverageDocLength: float64(totalTitleLength) /
			float64(documentCount),
	}
}
