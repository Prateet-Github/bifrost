package ingestion

import (
	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/document"
	"github.com/Prateet-Github/bifrost/internal/index"
)

type Ingester struct {
	analyzer *analysis.Analyzer
	index    *index.InvertedIndex
}

func NewIngester(
	analyzer *analysis.Analyzer,
	index *index.InvertedIndex,
) *Ingester {
	return &Ingester{
		analyzer: analyzer,
		index:    index,
	}
}

func (i *Ingester) Ingest(doc document.Document) error {
	tokens := i.analyzer.Analyze(doc.Body)

	i.index.AddDocument(doc.ID, tokens)

	return nil
}
