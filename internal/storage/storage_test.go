package storage

import (
	"path/filepath"
	"reflect"
	"testing"

	"go.etcd.io/bbolt"

	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
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

	err = store.SaveIndex(idx, "test-fingerprint")
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

	if err := store.SaveIndex(original, "test-fingerprint"); err != nil {
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

func TestSourceFingerprintPersistsAfterReopening(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Open storage.
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}

	// Save an index and its source fingerprint
	idx := index.NewInvertedIndex()
	expected := "test-fingerprint-123"

	if err := store.SaveIndex(idx, expected); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// Close the database to simulate application shutdown
	if err := store.Close(); err != nil {
		t.Fatalf("failed to close storage: %v", err)
	}

	// Reopen the same database.
	store, err = Open(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen storage: %v", err)
	}
	defer store.Close()

	// Verify that the fingerprint was persisted
	actual, err := store.LoadSourceFingerprint()
	if err != nil {
		t.Fatalf("failed to load source fingerprint: %v", err)
	}

	if actual != expected {
		t.Fatalf("fingerprint mismatch: got %q, want %q", actual, expected)
	}
}

func TestSaveIndexRemovesStalePostings(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	// Save the original index.
	original := index.NewInvertedIndex()

	original.AddDocument("go.txt", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "concurrency", Position: 1},
		{Text: "goroutines", Position: 2},
	})

	original.AddDocument("rust.txt", []tokenizer.Token{
		{Text: "rust", Position: 0},
		{Text: "concurrency", Position: 1},
	})

	if err := store.SaveIndex(original, "fingerprint-1"); err != nil {
		t.Fatalf("failed to save original index: %v", err)
	}

	original.AddDocument("go.txt", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "compilation", Position: 1},
	})

	if err := store.SaveIndex(original, "fingerprint-2"); err != nil {
		t.Fatalf("failed to save updated index: %v", err)
	}

	loaded, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("failed to load updated index: %v", err)
	}

	for _, term := range []string{"concurrency", "goroutines"} {
		for _, posting := range loaded.Lookup(term) {
			if posting.DocID == "go.txt" {
				t.Errorf("stale posting found: term=%q, docID=%q", term, posting.DocID)
			}
		}
	}

	concurrencyPostings := loaded.Lookup("concurrency")

	if len(concurrencyPostings) != 1 ||
		concurrencyPostings[0].DocID != "rust.txt" {
		t.Errorf("unexpected concurrency postings: %+v", concurrencyPostings)
	}

	compilationPostings := loaded.Lookup("compilation")

	if len(compilationPostings) != 1 ||
		compilationPostings[0].DocID != "go.txt" {
		t.Errorf("unexpected compilation postings: %+v", compilationPostings)
	}

	stats, exists := loaded.DocumentStats("go.txt")

	if !exists {
		t.Fatal("go.txt missing from document statistics")
	}

	if stats.Length != 2 {
		t.Errorf("expected document length 2, got %d", stats.Length)
	}
}
