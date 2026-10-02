package index

type Stats struct {
	Documents        int
	AverageDocLength float64
}

func (idx *InvertedIndex) DocumentFrequency(term string) int {
	return len(idx.terms[term])
}

func (idx *InvertedIndex) Stats() Stats {
	if len(idx.documents) == 0 {
		return Stats{}
	}

	totalLength := 0

	for _, doc := range idx.documents {
		totalLength += doc.Length
	}

	return Stats{
		Documents:        len(idx.documents),
		AverageDocLength: float64(totalLength) / float64(len(idx.documents)),
	}
}
