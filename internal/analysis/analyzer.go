package analysis

import (
	"github.com/Prateet-Github/bifrost/internal/analysis/filter"
	"github.com/Prateet-Github/bifrost/internal/analysis/normalizer"
	"github.com/Prateet-Github/bifrost/internal/analysis/stemmer"
	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
	"github.com/Prateet-Github/bifrost/internal/document"
)

type Analyzer struct {
	tokenizer  *tokenizer.Tokenizer
	normalizer *normalizer.Normalizer
	filter     *filter.StopWordFilter
	stemmer    *stemmer.PorterStemmer
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		tokenizer:  tokenizer.NewTokenizer(),
		normalizer: normalizer.NewNormalizer(),
		filter:     filter.NewStopWordFilter(),
		stemmer:    stemmer.NewPorterStemmer(),
	}
}

func (a *Analyzer) Analyze(doc document.Document) []tokenizer.Token {
	tokens := a.tokenizer.Tokenize(doc.Body)

	for i := range tokens {
		tokens[i] = a.normalizer.Normalize(tokens[i])
	}

	tokens = a.filter.Filter(tokens)

	for i := range tokens {
		tokens[i] = a.stemmer.StemToken(tokens[i])
	}

	return tokens
}
