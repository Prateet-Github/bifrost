package storage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
	"github.com/Prateet-Github/bifrost/internal/codec"
	"github.com/Prateet-Github/bifrost/internal/index"
	"go.etcd.io/bbolt"
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

	originalStats := original.Stats()
	loadedStats := loaded.Stats()

	if loadedStats.Documents != originalStats.Documents {
		t.Errorf(
			"document count after loading = %d, want %d",
			loadedStats.Documents,
			originalStats.Documents,
		)
	}

	if loadedStats.AverageDocLength != originalStats.AverageDocLength {
		t.Errorf(
			"average document length after loading = %f, want %f",
			loadedStats.AverageDocLength,
			originalStats.AverageDocLength,
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

	// Open storage
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

func BenchmarkSaveIndex(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Docs%d", size), func(b *testing.B) {
			analyzer := analysis.NewAnalyzer()
			idx := index.NewInvertedIndex()

			// Build the index outside the timed section.
			for i := 0; i < size; i++ {
				text := fmt.Sprintf(
					"distributed systems use concurrency and networking document number %d contains searchable content with indexing ranking retrieval and storage",
					i,
				)

				tokens := analyzer.Analyze(text)
				docID := fmt.Sprintf("doc-%06d.txt", i)

				idx.AddDocument(docID, tokens)
			}

			b.ReportMetric(float64(size), "documents")

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()

				dir := b.TempDir()
				dbPath := filepath.Join(dir, "bifrost.db")

				store, err := Open(dbPath)
				if err != nil {
					b.Fatal(err)
				}

				b.StartTimer()

				err = store.SaveIndex(idx, "benchmark-fingerprint")

				b.StopTimer()

				if err != nil {
					store.Close()
					b.Fatal(err)
				}

				if err := store.Close(); err != nil {
					b.Fatal(err)
				}

				info, err := os.Stat(dbPath)
				if err != nil {
					b.Fatal(err)
				}

				b.ReportMetric(float64(info.Size()), "db-bytes")
			}
		})
	}
}

func TestSaveAndLoadIndexPreservesPostingDetails(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	idx := index.NewInvertedIndex()

	idx.AddDocument("alpha.txt", []tokenizer.Token{
		{Text: "search", Position: 0},
		{Text: "engine", Position: 1},
		{Text: "search", Position: 2},
	})

	idx.AddDocument("beta.txt", []tokenizer.Token{
		{Text: "search", Position: 0},
		{Text: "systems", Position: 1},
	})

	if err := store.SaveIndex(idx, "test-fingerprint"); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	loaded, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	for _, term := range []string{"search", "engine", "systems"} {
		expected := idx.Lookup(term)
		actual := loaded.Lookup(term)

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(
				"postings for %q differ after reload:\ngot:  %+v\nwant: %+v",
				term,
				actual,
				expected,
			)
		}
	}
}

func TestSaveIndexUsesCompactPostingFormat(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	idx := index.NewInvertedIndex()
	idx.AddDocument("alpha.txt", []tokenizer.Token{
		{Text: "search", Position: 0},
	})

	if err := store.SaveIndex(idx, "test-fingerprint"); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	err = store.db.View(func(tx *bbolt.Tx) error {
		terms := tx.Bucket(termsBucket)
		data := terms.Get([]byte("search"))

		if data == nil {
			t.Fatal("compact posting data not found")
		}

		if !bytes.HasPrefix(data, compactPostingMagic) {
			t.Fatalf("posting data does not have compact format marker")
		}

		decoded, err := codec.DecodePostings(
			data[len(compactPostingMagic):],
		)
		if err != nil {
			t.Fatalf("failed to decode compact postings: %v", err)
		}

		if len(decoded) != 1 {
			t.Fatalf("posting count = %d, want 1", len(decoded))
		}

		if decoded[0].DocID != 1 {
			t.Errorf("numeric document ID = %d, want 1", decoded[0].DocID)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("failed to inspect stored postings: %v", err)
	}
}

func TestLoadIndexRejectsCorruptedPostings(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	idx := index.NewInvertedIndex()
	idx.AddDocument("alpha.txt", []tokenizer.Token{
		{Text: "search", Position: 0},
	})

	if err := store.SaveIndex(idx, "test-fingerprint"); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// Replace the valid compact posting with malformed data.
	err = store.db.Update(func(tx *bbolt.Tx) error {
		terms := tx.Bucket(termsBucket)

		corrupted := append(
			append([]byte(nil), compactPostingMagic...),
			0xFF,
		)

		return terms.Put([]byte("search"), corrupted)
	})
	if err != nil {
		t.Fatalf("failed to corrupt posting data: %v", err)
	}

	_, err = store.LoadIndex()
	if err == nil {
		t.Fatal("expected an error when loading corrupted postings")
	}
}

func TestLoadIndexRejectsMissingDocumentMapping(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	idx := index.NewInvertedIndex()
	idx.AddDocument("alpha.txt", []tokenizer.Token{
		{Text: "search", Position: 0},
	})

	if err := store.SaveIndex(idx, "test-fingerprint"); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// The only document receives numeric ID 1.
	err = store.db.Update(func(tx *bbolt.Tx) error {
		docIDs := tx.Bucket(docIDsBucket)

		return docIDs.Delete([]byte{0, 0, 0, 1})
	})
	if err != nil {
		t.Fatalf("failed to remove document mapping: %v", err)
	}

	_, err = store.LoadIndex()
	if err == nil {
		t.Fatal("expected an error when a document mapping is missing")
	}
}

func TestSaveAndLoadIndexPreservesTitlePostings(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	idx := index.NewInvertedIndex()

	titleTokens := []tokenizer.Token{
		{Text: "search", Position: 0},
		{Text: "engine", Position: 1},
		{Text: "search", Position: 2},
	}

	bodyTokens := []tokenizer.Token{
		{Text: "distributed", Position: 0},
		{Text: "systems", Position: 1},
	}

	idx.AddDocumentFields(
		"alpha.txt",
		titleTokens,
		bodyTokens,
	)

	if err := store.SaveIndex(idx, "test-fingerprint"); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	loaded, err := store.LoadIndex()
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	// Verify title postings, including frequencies and positions.
	for _, term := range []string{"search", "engine"} {
		expected := idx.LookupTitle(term)
		actual := loaded.LookupTitle(term)

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(
				"title postings for %q differ after reload:\ngot:  %+v\nwant: %+v",
				term,
				actual,
				expected,
			)
		}
	}

	// Verify that body postings are still preserved.
	for _, term := range []string{"distributed", "systems"} {
		expected := idx.Lookup(term)
		actual := loaded.Lookup(term)

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(
				"body postings for %q differ after reload:\ngot:  %+v\nwant: %+v",
				term,
				actual,
				expected,
			)
		}
	}

	// Verify body and title document lengths.
	expectedStats, expectedOK := idx.DocumentStats("alpha.txt")
	actualStats, actualOK := loaded.DocumentStats("alpha.txt")

	if !expectedOK || !actualOK {
		t.Fatal("document statistics missing before or after reload")
	}

	if actualStats != expectedStats {
		t.Errorf(
			"document stats differ after reload: got %+v, want %+v",
			actualStats,
			expectedStats,
		)
	}

	if actualStats.TitleLength != len(titleTokens) {
		t.Errorf(
			"title length = %d, want %d",
			actualStats.TitleLength,
			len(titleTokens),
		)
	}
}
