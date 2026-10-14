package index

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

func TestAddDocument(t *testing.T) {
	idx := NewInvertedIndex()

	tokens := []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "fast", Position: 1},
		{Text: "go", Position: 2},
		{Text: "go", Position: 3},
	}

	idx.AddDocument("D1", tokens)

	// Check document statistics.
	stats, ok := idx.documents["D1"]
	if !ok {
		t.Fatal("expected document D1 to exist")
	}

	if stats.Length != 4 {
		t.Errorf("expected document length 4, got %d", stats.Length)
	}

	// Check "go" posting.
	goPostings, ok := idx.terms["go"]
	if !ok {
		t.Fatal("expected term 'go' to exist")
	}

	if len(goPostings) != 1 {
		t.Fatalf("expected 1 posting for 'go', got %d", len(goPostings))
	}

	goPosting := goPostings[0]

	if goPosting.DocID != "D1" {
		t.Errorf("expected DocID D1, got %s", goPosting.DocID)
	}

	if goPosting.TermFreq != 3 {
		t.Errorf("expected term frequency 3, got %d", goPosting.TermFreq)
	}

	expectedPositions := []int{0, 2, 3}

	if len(goPosting.Positions) != len(expectedPositions) {
		t.Fatalf(
			"expected %d positions, got %d",
			len(expectedPositions),
			len(goPosting.Positions),
		)
	}

	for i, position := range expectedPositions {
		if goPosting.Positions[i] != position {
			t.Errorf(
				"expected position %d at index %d, got %d",
				position,
				i,
				goPosting.Positions[i],
			)
		}
	}

	// Check "fast" posting.
	fastPostings, ok := idx.terms["fast"]
	if !ok {
		t.Fatal("expected term 'fast' to exist")
	}

	if len(fastPostings) != 1 {
		t.Fatalf("expected 1 posting for 'fast', got %d", len(fastPostings))
	}

	if fastPostings[0].TermFreq != 1 {
		t.Errorf(
			"expected term frequency 1, got %d",
			fastPostings[0].TermFreq,
		)
	}

	if len(fastPostings[0].Positions) != 1 ||
		fastPostings[0].Positions[0] != 1 {
		t.Errorf(
			"expected positions [1], got %v",
			fastPostings[0].Positions,
		)
	}
}

func TestAddMultipleDocuments(t *testing.T) {
	idx := NewInvertedIndex()

	doc1 := []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "fast", Position: 1},
	}

	doc2 := []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "server", Position: 1},
	}

	idx.AddDocument("D1", doc1)
	idx.AddDocument("D2", doc2)

	postings := idx.terms["go"]

	if len(postings) != 2 {
		t.Fatalf("expected 2 postings for 'go', got %d", len(postings))
	}

	if postings[0].DocID != "D1" {
		t.Errorf("expected first posting to be D1, got %s", postings[0].DocID)
	}

	if postings[1].DocID != "D2" {
		t.Errorf("expected second posting to be D2, got %s", postings[1].DocID)
	}

	if postings[0].TermFreq != 1 {
		t.Errorf("expected D1 TF = 1, got %d", postings[0].TermFreq)
	}

	if postings[1].TermFreq != 1 {
		t.Errorf("expected D2 TF = 1, got %d", postings[1].TermFreq)
	}
}

func TestLookup(t *testing.T) {
	idx := NewInvertedIndex()

	tokens := []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "fast", Position: 1},
		{Text: "go", Position: 2},
	}

	idx.AddDocument("D1", tokens)

	postings := idx.Lookup("go")

	if len(postings) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(postings))
	}

	if postings[0].DocID != "D1" {
		t.Errorf("expected DocID D1, got %s", postings[0].DocID)
	}

	if postings[0].TermFreq != 2 {
		t.Errorf("expected TF 2, got %d", postings[0].TermFreq)
	}

	expectedPositions := []int{0, 2}

	if len(postings[0].Positions) != len(expectedPositions) {
		t.Fatalf(
			"expected %d positions, got %d",
			len(expectedPositions),
			len(postings[0].Positions),
		)
	}

	for i, expected := range expectedPositions {
		if postings[0].Positions[i] != expected {
			t.Errorf(
				"expected position %d at index %d, got %d",
				expected,
				i,
				postings[0].Positions[i],
			)
		}
	}
}

func TestLookupUnknownTerm(t *testing.T) {
	idx := NewInvertedIndex()

	tokens := []tokenizer.Token{
		{Text: "go", Position: 0},
	}

	idx.AddDocument("D1", tokens)

	postings := idx.Lookup("rust")

	if len(postings) != 0 {
		t.Fatalf(
			"expected no postings for unknown term, got %d",
			len(postings),
		)
	}
}

func TestDocumentFrequency(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "fast", Position: 1},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "server", Position: 1},
	})

	idx.AddDocument("D3", []tokenizer.Token{
		{Text: "rust", Position: 0},
	})

	tests := []struct {
		term string
		want int
	}{
		{"go", 2},
		{"fast", 1},
		{"rust", 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		got := idx.DocumentFrequency(tt.term)

		if got != tt.want {
			t.Errorf(
				"DocumentFrequency(%q) = %d, want %d",
				tt.term,
				got,
				tt.want,
			)
		}
	}
}

func TestStats(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocument("D1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "fast", Position: 1},
	})

	idx.AddDocument("D2", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "server", Position: 1},
	})

	idx.AddDocument("D3", []tokenizer.Token{
		{Text: "rust", Position: 0},
	})

	stats := idx.Stats()

	if stats.Documents != 3 {
		t.Errorf(
			"expected 3 documents, got %d",
			stats.Documents,
		)
	}

	expectedAverage := 5.0 / 3.0

	if stats.AverageDocLength != expectedAverage {
		t.Errorf(
			"expected average document length %f, got %f",
			expectedAverage,
			stats.AverageDocLength,
		)
	}
}

func TestAddDocumentReplacesExistingDocument(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocument("go.txt", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "concurrency", Position: 1},
		{Text: "goroutines", Position: 2},
	})

	idx.AddDocument("rust.txt", []tokenizer.Token{
		{Text: "rust", Position: 0},
		{Text: "concurrency", Position: 1},
	})

	idx.AddDocument("go.txt", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "compilation", Position: 1},
	})

	for _, term := range []string{"concurrency", "goroutines"} {
		postings := idx.Lookup(term)

		for _, posting := range postings {
			if posting.DocID == "go.txt" {
				t.Errorf(
					"stale posting found: term=%q, docID=%q",
					term,
					posting.DocID,
				)
			}
		}
	}

	postings := idx.Lookup("concurrency")

	if len(postings) != 1 {
		t.Fatalf("expected 1 concurrency posting, got %d", len(postings))
	}

	if postings[0].DocID != "rust.txt" {
		t.Errorf(
			"expected concurrency to reference rust.txt, got %q",
			postings[0].DocID,
		)
	}

	postings = idx.Lookup("compilation")

	if len(postings) != 1 || postings[0].DocID != "go.txt" {
		t.Errorf("expected compilation to reference go.txt, got %v", postings)
	}

	stats, exists := idx.DocumentStats("go.txt")
	if !exists {
		t.Fatal("expected go.txt to exist in document statistics")
	}

	if stats.Length != 2 {
		t.Errorf("expected document length 2, got %d", stats.Length)
	}
}

func TestStatsAfterAddAndRemove(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocument("doc1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "systems", Position: 1},
	})

	idx.AddDocument("doc2", []tokenizer.Token{
		{Text: "rust", Position: 0},
		{Text: "safe", Position: 1},
		{Text: "fast", Position: 2},
	})

	stats := idx.Stats()

	if stats.Documents != 2 {
		t.Fatalf("expected 2 documents, got %d", stats.Documents)
	}

	if stats.AverageDocLength != 2.5 {
		t.Fatalf("expected average length 2.5, got %f", stats.AverageDocLength)
	}

	idx.Remove("doc1")

	stats = idx.Stats()

	if stats.Documents != 1 {
		t.Fatalf("expected 1 document, got %d", stats.Documents)
	}

	if stats.AverageDocLength != 3 {
		t.Fatalf("expected average length 3, got %f", stats.AverageDocLength)
	}

	idx.Remove("doc2")

	stats = idx.Stats()

	if stats.Documents != 0 {
		t.Fatalf("expected 0 documents, got %d", stats.Documents)
	}

	if stats.AverageDocLength != 0 {
		t.Fatalf("expected average length 0, got %f", stats.AverageDocLength)
	}
}

func TestStatsAfterDocumentReplacement(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocument("doc1", []tokenizer.Token{
		{Text: "go", Position: 0},
		{Text: "systems", Position: 1},
		{Text: "concurrency", Position: 2},
	})

	idx.AddDocument("doc1", []tokenizer.Token{
		{Text: "go", Position: 0},
	})

	stats := idx.Stats()

	if stats.Documents != 1 {
		t.Fatalf("expected 1 document, got %d", stats.Documents)
	}

	if stats.AverageDocLength != 1 {
		t.Fatalf("expected average length 1, got %f", stats.AverageDocLength)
	}
}

func TestSetDocumentStatsUpdatesTotalDocumentLength(t *testing.T) {
	idx := NewInvertedIndex()

	idx.SetDocumentStats("doc1", DocumentStats{Length: 100})
	idx.SetDocumentStats("doc2", DocumentStats{Length: 200})

	stats := idx.Stats()

	if stats.Documents != 2 {
		t.Fatalf("expected 2 documents, got %d", stats.Documents)
	}

	if stats.AverageDocLength != 150 {
		t.Fatalf(
			"expected average document length 150, got %f",
			stats.AverageDocLength,
		)
	}

	idx.SetDocumentStats("doc1", DocumentStats{Length: 200})

	stats = idx.Stats()

	if stats.AverageDocLength != 200 {
		t.Fatalf(
			"expected average document length 200 after replacement, got %f",
			stats.AverageDocLength,
		)
	}
}

func TestAddDocumentFields(t *testing.T) {
	idx := NewInvertedIndex()

	titleTokens := []tokenizer.Token{
		{Text: "distributed", Position: 0},
		{Text: "systems", Position: 1},
	}

	bodyTokens := []tokenizer.Token{
		{Text: "distributed", Position: 3},
		{Text: "systems", Position: 4},
		{Text: "systems", Position: 5},
	}

	idx.AddDocumentFields("doc1", titleTokens, bodyTokens)

	titlePostings := idx.LookupTitle("distributed")
	bodyPostings := idx.Lookup("distributed")

	if len(titlePostings) != 1 {
		t.Fatalf("expected 1 title posting, got %d", len(titlePostings))
	}

	if len(bodyPostings) != 1 {
		t.Fatalf("expected 1 body posting, got %d", len(bodyPostings))
	}

	if titlePostings[0].TermFreq != 1 {
		t.Errorf("expected title term frequency 1, got %d", titlePostings[0].TermFreq)
	}

	if bodyPostings[0].TermFreq != 1 {
		t.Errorf("expected body term frequency 1, got %d", bodyPostings[0].TermFreq)
	}

	if titlePostings[0].Positions[0] != 0 {
		t.Errorf("expected title position 0, got %v", titlePostings[0].Positions)
	}

	if bodyPostings[0].Positions[0] != 3 {
		t.Errorf("expected body position 3, got %v", bodyPostings[0].Positions)
	}

	systemsTitle := idx.LookupTitle("systems")
	systemsBody := idx.Lookup("systems")

	if systemsTitle[0].TermFreq != 1 {
		t.Errorf("expected title systems frequency 1, got %d", systemsTitle[0].TermFreq)
	}

	if systemsBody[0].TermFreq != 2 {
		t.Errorf("expected body systems frequency 2, got %d", systemsBody[0].TermFreq)
	}
}

func TestAddDocumentFieldsStats(t *testing.T) {
	idx := NewInvertedIndex()

	titleTokens := []tokenizer.Token{
		{Text: "distributed", Position: 0},
		{Text: "systems", Position: 1},
	}

	bodyTokens := []tokenizer.Token{
		{Text: "systems", Position: 0},
		{Text: "scale", Position: 1},
		{Text: "well", Position: 2},
	}

	idx.AddDocumentFields("doc1", titleTokens, bodyTokens)

	stats, exists := idx.DocumentStats("doc1")
	if !exists {
		t.Fatal("expected document stats to exist")
	}

	if stats.TitleLength != 2 {
		t.Errorf("expected title length 2, got %d", stats.TitleLength)
	}

	if stats.Length != 3 {
		t.Errorf("expected body length 3, got %d", stats.Length)
	}

	indexStats := idx.Stats()

	if indexStats.AverageDocLength != 3 {
		t.Errorf("expected average body length 3, got %f", indexStats.AverageDocLength)
	}
}

func TestAddDocumentFieldsReplacesExistingDocument(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocumentFields(
		"doc1",
		[]tokenizer.Token{{Text: "oldtitle", Position: 0}},
		[]tokenizer.Token{{Text: "oldbody", Position: 0}},
	)

	idx.AddDocumentFields(
		"doc1",
		[]tokenizer.Token{{Text: "newtitle", Position: 0}},
		[]tokenizer.Token{{Text: "newbody", Position: 0}},
	)

	if len(idx.LookupTitle("oldtitle")) != 0 {
		t.Error("expected old title posting to be removed")
	}

	if len(idx.Lookup("oldbody")) != 0 {
		t.Error("expected old body posting to be removed")
	}

	if len(idx.LookupTitle("newtitle")) != 1 {
		t.Error("expected new title posting")
	}

	if len(idx.Lookup("newbody")) != 1 {
		t.Error("expected new body posting")
	}
}

func TestRemoveDocumentFields(t *testing.T) {
	idx := NewInvertedIndex()

	idx.AddDocumentFields(
		"doc1",
		[]tokenizer.Token{{Text: "titleterm", Position: 0}},
		[]tokenizer.Token{{Text: "bodyterm", Position: 0}},
	)

	idx.Remove("doc1")

	if len(idx.LookupTitle("titleterm")) != 0 {
		t.Error("expected title posting to be removed")
	}

	if len(idx.Lookup("bodyterm")) != 0 {
		t.Error("expected body posting to be removed")
	}

	if _, exists := idx.DocumentStats("doc1"); exists {
		t.Error("expected document stats to be removed")
	}
}
