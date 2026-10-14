package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenBuildsIndex(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bifrost.db")
	docsPath := t.TempDir()

	writeDocument(t, docsPath, "go.txt", "Go programming language")

	bifrost, err := Open(Config{
		DBPath:   dbPath,
		DocsPath: docsPath,
	})
	if err != nil {
		t.Fatalf("failed to open engine: %v", err)
	}

	results := bifrost.Search("programming", 10)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].DocID != "go.txt" {
		t.Fatalf("expected go.txt, got %q", results[0].DocID)
	}

	if err := bifrost.Close(); err != nil {
		t.Fatalf("failed to close engine: %v", err)
	}
}

func TestOpenLoadsExistingIndex(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bifrost.db")
	docsPath := t.TempDir()

	writeDocument(t, docsPath, "go.txt", "Go programming language")

	first, err := Open(Config{
		DBPath:   dbPath,
		DocsPath: docsPath,
	})
	if err != nil {
		t.Fatalf("failed to open engine initially: %v", err)
	}

	if err := first.Close(); err != nil {
		t.Fatalf("failed to close first engine: %v", err)
	}

	second, err := Open(Config{
		DBPath:   dbPath,
		DocsPath: docsPath,
	})
	if err != nil {
		t.Fatalf("failed to reopen engine: %v", err)
	}
	defer second.Close()

	results := second.Search("programming", 10)

	if len(results) != 1 {
		t.Fatalf("expected 1 result after reopening, got %d", len(results))
	}

	if results[0].DocID != "go.txt" {
		t.Fatalf("expected go.txt, got %q", results[0].DocID)
	}
}

func TestOpenRebuildsIndexWhenDocumentsChange(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bifrost.db")
	docsPath := t.TempDir()

	writeDocument(t, docsPath, "go.txt", "Go programming language")

	first, err := Open(Config{
		DBPath:   dbPath,
		DocsPath: docsPath,
	})
	if err != nil {
		t.Fatalf("failed to open engine initially: %v", err)
	}

	if err := first.Close(); err != nil {
		t.Fatalf("failed to close first engine: %v", err)
	}

	if err := os.Remove(filepath.Join(docsPath, "go.txt")); err != nil {
		t.Fatalf("failed to remove original document: %v", err)
	}

	writeDocument(t, docsPath, "rust.txt", "Rust systems programming")

	second, err := Open(Config{
		DBPath:   dbPath,
		DocsPath: docsPath,
	})
	if err != nil {
		t.Fatalf("failed to reopen engine after document changes: %v", err)
	}
	defer second.Close()

	oldResults := second.Search("language", 10)
	if len(oldResults) != 0 {
		t.Fatalf("expected no results for removed document content, got %d", len(oldResults))
	}

	newResults := second.Search("systems", 10)
	if len(newResults) != 1 {
		t.Fatalf("expected 1 result for new document content, got %d", len(newResults))
	}

	if newResults[0].DocID != "rust.txt" {
		t.Fatalf("expected rust.txt, got %q", newResults[0].DocID)
	}
}

func writeDocument(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write document %s: %v", name, err)
	}
}
