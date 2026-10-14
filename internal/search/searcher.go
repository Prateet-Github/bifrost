package search

import (
	"container/heap"
	"sort"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
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

// deduplicate query terms
func deduplicateTokens(tokens []tokenizer.Token) []tokenizer.Token {
	seen := make(map[string]struct{}, len(tokens))
	unique := make([]tokenizer.Token, 0, len(tokens))

	for _, token := range tokens {
		if _, exists := seen[token.Text]; exists {
			continue
		}

		seen[token.Text] = struct{}{}
		unique = append(unique, token)
	}

	return unique
}

func (s *Searcher) Candidates(rawQuery string) []Candidate {
	tokens := s.queryAnalyzer.Analyze(rawQuery)

	return s.candidates(tokens)
}

func (s *Searcher) candidates(tokens []tokenizer.Token) []Candidate {
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
	// tokens := s.queryAnalyzer.Analyze(rawQuery)
	tokens := deduplicateTokens(s.queryAnalyzer.Analyze(rawQuery))

	if len(tokens) == 0 || k <= 0 {
		return nil
	}

	candidates := s.candidates(tokens)

	terms := make([]string, 0, len(tokens))

	for _, token := range tokens {
		terms = append(terms, token.Text)
	}

	queryScorer := s.scorer.PrepareQuery(terms)

	results := make(resultHeap, 0, min(k, len(candidates)))
	heap.Init(&results)

	for _, candidate := range candidates {
		doc, exists := s.index.DocumentStats(candidate.DocID)

		var score float64

		if exists {
			for _, match := range candidate.Matches {
				score += queryScorer.ScoreTerm(
					match.Term,
					match.Posting.TermFreq,
					doc.Length,
				)
			}
		}

		result := Result{
			DocID: candidate.DocID,
			Score: score,
		}

		if results.Len() < k {
			heap.Push(&results, result)
			continue
		}

		worst := results[0]

		better := result.Score > worst.Score ||
			(result.Score == worst.Score && result.DocID < worst.DocID)

		if better {
			heap.Pop(&results)
			heap.Push(&results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}

		return results[i].Score > results[j].Score
	})

	return results
}

func (s *Searcher) PhraseMatch(rawPhrase string, docID string) bool {
	tokens := s.queryAnalyzer.Analyze(rawPhrase)

	if len(tokens) == 0 {
		return false
	}

	terms := make([]string, 0, len(tokens))
	for _, token := range tokens {
		terms = append(terms, token.Text)
	}

	return phraseMatch(s.index, docID, terms)
}

type resultHeap []Result

func (h resultHeap) Len() int {
	return len(h)
}

func (h resultHeap) Less(i, j int) bool {
	if h[i].Score == h[j].Score {
		// for equal scores, the lexicographically larger DocID is worse
		return h[i].DocID > h[j].DocID
	}

	// tge lower score is worse and belongs at the root
	return h[i].Score < h[j].Score
}

func (h resultHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *resultHeap) Push(value any) {
	*h = append(*h, value.(Result))
}

func (h *resultHeap) Pop() any {
	old := *h
	n := len(old)
	value := old[n-1]

	*h = old[:n-1]

	return value
}

func (s *Searcher) SearchPhrase(rawPhrase string, k int) []Result {
	tokens := deduplicateTokens(s.queryAnalyzer.Analyze(rawPhrase))

	if len(tokens) == 0 || k <= 0 {
		return nil
	}

	terms := make([]string, 0, len(tokens))
	for _, token := range tokens {
		terms = append(terms, token.Text)
	}

	candidates := s.candidates(tokens)

	queryScorer := s.scorer.PrepareQuery(terms)

	results := make(resultHeap, 0, min(k, len(candidates)))
	heap.Init(&results)

	for _, candidate := range candidates {
		if !s.PhraseMatch(rawPhrase, candidate.DocID) {
			continue
		}

		doc, exists := s.index.DocumentStats(candidate.DocID)
		if !exists {
			continue
		}

		var score float64

		for _, match := range candidate.Matches {
			score += queryScorer.ScoreTerm(
				match.Term,
				match.Posting.TermFreq,
				doc.Length,
			)
		}

		result := Result{
			DocID: candidate.DocID,
			Score: score,
		}

		if results.Len() < k {
			heap.Push(&results, result)
			continue
		}

		worst := results[0]

		better := result.Score > worst.Score ||
			(result.Score == worst.Score && result.DocID < worst.DocID)

		if better {
			heap.Pop(&results)
			heap.Push(&results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}

		return results[i].Score > results[j].Score
	})

	return results
}
