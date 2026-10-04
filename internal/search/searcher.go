package search

import (
	"sort"

	"github.com/Prateet-Github/bifrost/internal/index"
	"github.com/Prateet-Github/bifrost/internal/query"
	"github.com/Prateet-Github/bifrost/internal/scoring"
)

type Match struct {
	Term    string
	Posting index.Posting
}

type Candidate struct {
	DocID   string
	Matches []Match
}

type Result struct {
	DocID string
	Score float64
}

type Searcher struct {
	index         *index.InvertedIndex
	queryAnalyzer *query.Analyzer
	scorer        *scoring.BM25
}

func NewSearcher(idx *index.InvertedIndex) *Searcher {
	return &Searcher{
		index:         idx,
		queryAnalyzer: query.NewAnalyzer(),
		scorer:        scoring.NewBM25(idx),
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

func (s *Searcher) Search(rawQuery string, k int) []Result {
	tokens := s.queryAnalyzer.Analyze(rawQuery)

	if len(tokens) == 0 || k <= 0 {
		return nil
	}

	candidates := s.Candidates(rawQuery)

	results := make([]Result, 0, len(candidates))

	for _, candidate := range candidates {
		terms := make([]string, 0, len(candidate.Matches))

		for _, match := range candidate.Matches {
			terms = append(terms, match.Term)
		}

		score := s.scorer.Score(
			terms,
			candidate.DocID,
		)

		results = append(results, Result{
			DocID: candidate.DocID,
			Score: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}

		return results[i].Score > results[j].Score
	})

	if k > len(results) {
		k = len(results)
	}

	return results[:k]
}

func (s *Searcher) PhraseMatch(
	rawPhrase string,
	docID string,
) bool {
	tokens := s.queryAnalyzer.Analyze(rawPhrase)

	if len(tokens) != 2 {
		return false
	}

	return phraseMatch(
		s.index,
		docID,
		[]string{
			tokens[0].Text,
			tokens[1].Text,
		},
	)
}
