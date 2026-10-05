package ingestion

import (
	"os"
	"path/filepath"

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

func (i *Ingester) IngestDirectory(path string) error {
	reader := NewReader()

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".txt" {
			continue
		}

		filePath := filepath.Join(path, entry.Name())

		doc, err := reader.Read(filePath)
		if err != nil {
			return err
		}

		if err := i.Ingest(doc); err != nil {
			return err
		}
	}

	return nil
}
