package storage

import (
	"reflect"
	"testing"

	"go.etcd.io/bbolt"

	"github.com/Prateet-Github/bifrost/internal/analysis"
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

func TestSaveIndex(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	analyzer := analysis.NewAnalyzer()
	idx := index.NewInvertedIndex()

	tokens := analyzer.Analyze(
		"Go is a programming language",
	)

	idx.AddDocument("go.txt", tokens)

	err = store.SaveIndex(idx)
	if err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	postings, err := store.LoadPostings("program")
	if err != nil {
		t.Fatalf("failed to load postings: %v", err)
	}

	if len(postings) != 1 {
		t.Fatalf(
			"postings count = %d, want 1",
			len(postings),
		)
	}

	stats, err := store.LoadDocumentStats("go.txt")
	if err != nil {
		t.Fatalf("failed to load document stats: %v", err)
	}

	if stats.Length != len(tokens) {
		t.Fatalf(
			"document length = %d, want %d",
			stats.Length,
			len(tokens),
		)
	}
}

func TestSaveAndLoadIndex(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	analyzer := analysis.NewAnalyzer()
	original := index.NewInvertedIndex()

	tokens := analyzer.Analyze(
		"distributed systems search engine",
	)

	original.AddDocument("systems.txt", tokens)

	// First process: save
	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}

	if err := store.SaveIndex(original); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("failed to close storage: %v", err)
	}

	// Simulate process restart
	store, err = Open(path)
	if err != nil {
		t.Fatalf("failed to reopen storage: %v", err)
	}
	defer store.Close()

	// Second process: load
	loaded, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	// Verify term postings
	postings := loaded.Lookup("distribut")

	if len(postings) != 1 {
		t.Fatalf(
			"postings count = %d, want 1",
			len(postings),
		)
	}

	if postings[0].DocID != "systems.txt" {
		t.Fatalf(
			"docID = %s, want systems.txt",
			postings[0].DocID,
		)
	}

	// Verify document statistics
	stats, ok := loaded.DocumentStats("systems.txt")

	if !ok {
		t.Fatal("document stats not found")
	}

	if stats.Length != len(tokens) {
		t.Fatalf(
			"document length = %d, want %d",
			stats.Length,
			len(tokens),
		)
	}
}

func TestIsIndexReady(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	// The database file exists, but the index has not been built
	ready, err := store.IsIndexReady()
	if err != nil {
		t.Fatalf("failed to check index readiness: %v", err)
	}

	if ready {
		t.Fatal("index reported ready before being built")
	}

	// Simulate a successfully persisted index
	err = store.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)

		if err := bucket.Put(builtKey, []byte("true")); err != nil {
			return err
		}

		return bucket.Put(
			schemaVersionKey,
			[]byte(currentSchemaVersion),
		)
	})

	if err != nil {
		t.Fatalf("failed to mark index as built: %v", err)
	}

	ready, err = store.IsIndexReady()
	if err != nil {
		t.Fatalf("failed to check index readiness: %v", err)
	}

	if !ready {
		t.Fatal("index reported not ready after being built")
	}
}

func TestIsIndexReady_WrongSchemaVersion(t *testing.T) {
	path := t.TempDir() + "/bifrost.db"

	store, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	err = store.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)

		if err := bucket.Put(builtKey, []byte("true")); err != nil {
			return err
		}

		return bucket.Put(schemaVersionKey, []byte("999"))
	})

	if err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}

	ready, err := store.IsIndexReady()
	if err != nil {
		t.Fatalf("failed to check index readiness: %v", err)
	}

	if ready {
		t.Fatal("index reported ready with incompatible schema version")
	}
}
