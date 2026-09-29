package filter

import "github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"

type StopWordFilter struct {
	stopWords map[string]struct{}
}

func NewStopWordFilter() *StopWordFilter {
	return &StopWordFilter{
		stopWords: map[string]struct{}{
			"a":    {},
			"an":   {},
			"and":  {},
			"are":  {},
			"as":   {},
			"at":   {},
			"be":   {},
			"by":   {},
			"for":  {},
			"from": {},
			"in":   {},
			"is":   {},
			"it":   {},
			"of":   {},
			"on":   {},
			"or":   {},
			"that": {},
			"the":  {},
			"this": {},
			"to":   {},
			"was":  {},
			"were": {},
			"with": {},
		},
	}
}

func (f *StopWordFilter) Filter(tokens []tokenizer.Token) []tokenizer.Token {
	filtered := make([]tokenizer.Token, 0, len(tokens))

	for _, token := range tokens {
		if _, exists := f.stopWords[token.Text]; exists {
			continue
		}

		filtered = append(filtered, token)
	}

	return filtered
}
