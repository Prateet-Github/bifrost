package ingestion

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/document"
	"github.com/Prateet-Github/bifrost/internal/index"
)

func TestIngester(t *testing.T) {
	analyzer := analysis.NewAnalyzer()
	idx := index.NewInvertedIndex()

	ingester := NewIngester(analyzer, idx)

	doc := document.Document{
		ID:    "doc-1",
		Title: "Go Programming",
		Body:  "The Go developers are building fast servers",
	}

	err := ingester.Ingest(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	postings := idx.Lookup("develop")

	if len(postings) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(postings))
	}

	if postings[0].DocID != "doc-1" {
		t.Fatalf("expected doc-1, got %s", postings[0].DocID)
	}
}
