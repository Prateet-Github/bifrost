package search

import (
	"github.com/Prateet-Github/bifrost/internal/index"
	"github.com/Prateet-Github/bifrost/internal/query"
)

type Match struct {
	Term    string
	Posting index.Posting
}

type Candidate struct {
	DocID   string
	Matches []Match
}

type Searcher struct {
	index         *index.InvertedIndex
	queryAnalyzer *query.Analyzer
}

func NewSearcher(idx *index.InvertedIndex) *Searcher {
	return &Searcher{
		index:         idx,
		queryAnalyzer: query.NewAnalyzer(),
	}
}

func (s *Searcher) Candidates(rawQuery string) []Candidate {
	tokens := s.queryAnalyzer.Analyze(rawQuery)

	candidates := make(map[string]*Candidate)

	for _, token := range tokens {
		postings := s.index.Lookup(token.Text)

		for _, posting := range postings {
			candidate, exists := candidates[posting.DocID]

			if !exists {
				candidate = &Candidate{
					DocID: posting.DocID,
				}

				candidates[posting.DocID] = candidate
			}

			candidate.Matches = append(candidate.Matches, Match{
				Term:    token.Text,
				Posting: posting,
			})
		}
	}

	results := make([]Candidate, 0, len(candidates))

	for _, candidate := range candidates {
		results = append(results, *candidate)
	}

	return results
}
