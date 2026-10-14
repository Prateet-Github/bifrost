package search

import "github.com/Prateet-Github/bifrost/internal/index"

func phraseMatchField(
	idx *index.InvertedIndex,
	docID string,
	terms []string,
	isTitle bool,
) bool {
	if len(terms) == 0 {
		return false
	}

	currentPositions := positionsForTerm(
		idx,
		terms[0],
		docID,
		isTitle,
	)

	if len(currentPositions) == 0 {
		return false
	}

	for _, term := range terms[1:] {
		nextPositions := positionsForTerm(
			idx,
			term,
			docID,
			isTitle,
		)

		if len(nextPositions) == 0 {
			return false
		}

		matchedPositions := make([]int, 0)

		for _, currentPos := range currentPositions {
			for _, nextPos := range nextPositions {
				if nextPos == currentPos+1 {
					matchedPositions = append(
						matchedPositions,
						nextPos,
					)
				}
			}
		}

		if len(matchedPositions) == 0 {
			return false
		}

		currentPositions = matchedPositions
	}

	return true
}

func positionsForTerm(
	idx *index.InvertedIndex,
	term string,
	docID string,
	isTitle bool,
) []int {
	var postings []index.Posting

	if isTitle {
		postings = idx.LookupTitle(term)
	} else {
		postings = idx.Lookup(term)
	}

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
