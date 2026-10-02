package query

import (
	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

type Analyzer struct {
	textAnalyzer *analysis.Analyzer
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		textAnalyzer: analysis.NewAnalyzer(),
	}
}

func (a *Analyzer) Analyze(query string) []tokenizer.Token {
	return a.textAnalyzer.Analyze(query)
}
