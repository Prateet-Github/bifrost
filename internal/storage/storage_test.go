package storage

import (
	"reflect"
	"testing"

	"github.com/Prateet-Github/bifrost/internal/index"
)

func TestOpen(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Fatal("expected database, got nil")
	}
}

func TestSaveAndLoadPostings(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	expected := []index.Posting{
		{
			DocID:     "D1",
			TermFreq:  2,
			Positions: []int{1, 4},
		},
		{
			DocID:     "D2",
			TermFreq:  1,
			Positions: []int{3},
		},
	}

	err = store.SavePostings("search", expected)
	if err != nil {
		t.Fatalf("failed to save postings: %v", err)
	}

	actual, err := store.LoadPostings("search")
	if err != nil {
		t.Fatalf("failed to load postings: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("loaded postings = %+v, want %+v", actual, expected)
	}
}

func TestSaveAndLoadDocumentStats(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	expected := index.DocumentStats{
		Length: 42,
	}

	err = store.SaveDocumentStats("D1", expected)
	if err != nil {
		t.Fatalf("failed to save document stats: %v", err)
	}

	actual, err := store.LoadDocumentStats("D1")
	if err != nil {
		t.Fatalf("failed to load document stats: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(
			"loaded stats = %+v, want %+v",
			actual,
			expected,
		)
	}
}

func TestSaveAndLoadStats(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	expected := index.Stats{
		Documents:        10,
		AverageDocLength: 25.5,
	}

	err = store.SaveStats(expected)
	if err != nil {
		t.Fatalf("failed to save stats: %v", err)
	}

	actual, err := store.LoadStats()
	if err != nil {
		t.Fatalf("failed to load stats: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(
			"loaded stats = %+v, want %+v",
			actual,
			expected,
		)
	}
}
