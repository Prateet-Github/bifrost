package ingestion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReader(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "search.txt")

	content := "Bifrost is a search engine written in Go."

	err := os.WriteFile(
		path,
		[]byte(content),
		0644,
	)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	reader := NewReader()

	doc, err := reader.Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.ID != "search.txt" {
		t.Fatalf("expected ID search.txt, got %s", doc.ID)
	}

	if doc.Title != "search.txt" {
		t.Fatalf("expected title search.txt, got %s", doc.Title)
	}

	if doc.Body != content {
		t.Fatalf("expected body %q, got %q", content, doc.Body)
	}
}
