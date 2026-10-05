package ingestion

import (
	"os"
	"path/filepath"
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

func TestIngesterDirectory(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(dir, "go.txt"),
		[]byte("Go is fast"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(dir, "search.txt"),
		[]byte("Search engines rank documents"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(dir, "ignored.md"),
		[]byte("This should not be indexed"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	analyzer := analysis.NewAnalyzer()
	idx := index.NewInvertedIndex()

	ingester := NewIngester(analyzer, idx)

	err = ingester.IngestDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := idx.Stats()

	if stats.Documents != 2 {
		t.Fatalf("expected 2 documents, got %d", stats.Documents)
	}

	if len(idx.Lookup("go")) != 1 {
		t.Fatal("expected go to be indexed")
	}

	if len(idx.Lookup("search")) != 1 {
		t.Fatal("expected search to be indexed")
	}
}
