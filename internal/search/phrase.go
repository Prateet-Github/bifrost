package search

import "github.com/Prateet-Github/bifrost/internal/index"

func phraseMatch(
	idx *index.InvertedIndex,
	docID string,
	terms []string,
) bool {
	if len(terms) == 0 {
		return false
	}

	if len(terms) == 1 {
		return len(
			positionsForTerm(idx, terms[0], docID),
		) > 0
	}

	firstPositions := positionsForTerm(
		idx,
		terms[0],
		docID,
	)

	secondPositions := positionsForTerm(
		idx,
		terms[1],
		docID,
	)

	return positionsMatch(
		firstPositions,
		secondPositions,
	)
}

func positionsForTerm(
	idx *index.InvertedIndex,
	term string,
	docID string,
) []int {
	postings := idx.Lookup(term)

	for _, posting := range postings {
		if posting.DocID == docID {
			return posting.Positions
		}
	}

	return nil
}

func positionsMatch(
	first []int,
	second []int,
) bool {
	for _, firstPos := range first {
		for _, secondPos := range second {
			if secondPos == firstPos+1 {
				return true
			}
		}
	}

	return false
}
