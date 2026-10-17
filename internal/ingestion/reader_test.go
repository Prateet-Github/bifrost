package ingestion

import (
	"os"
	"path/filepath"
	"strings"
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

	if doc.Title != content {
		t.Fatalf("expected title %q, got %q", content, doc.Title)
	}

	if doc.Body != content {
		t.Fatalf("expected body %q, got %q", content, doc.Body)
	}
}

func TestReaderDerivesTitle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		expected string
	}{
		{
			name:     "markdown H1",
			filename: "document.txt",
			content:  "# Bifrost Search Engine\n\nSome content.",
			expected: "Bifrost Search Engine",
		},
		{
			name:     "first non-empty line",
			filename: "document.txt",
			content:  "\n\nBifrost search engine\nSome content.",
			expected: "Bifrost search engine",
		},
		{
			name:     "filename fallback",
			filename: "distributed-systems_notes.txt",
			content:  "\n\n",
			expected: "distributed systems notes",
		},
		{
			name:     "title capped at 120 characters",
			filename: "document.txt",
			content:  strings.Repeat("a", 150),
			expected: strings.Repeat("a", 120),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			reader := NewReader()
			doc, err := reader.Read(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if doc.Title != tt.expected {
				t.Errorf("expected title %q, got %q", tt.expected, doc.Title)
			}
		})
	}
}
