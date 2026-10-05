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
