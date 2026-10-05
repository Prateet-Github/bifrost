package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/index"
	"github.com/Prateet-Github/bifrost/internal/ingestion"
	"github.com/Prateet-Github/bifrost/internal/search"
	"github.com/Prateet-Github/bifrost/internal/storage"
)

const dbPath = "bifrost.db"

func main() {
	analyzer := analysis.NewAnalyzer()

	// check whether a persisted index already exists
	_, statErr := os.Stat(dbPath)
	dbExists := statErr == nil

	store, err := storage.Open(dbPath)
	if err != nil {
		fmt.Printf("failed to open storage: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	var idx *index.InvertedIndex

	if dbExists {
		// load the existing index from disk
		fmt.Println("Loading index from disk...")

		idx, err = store.LoadIndex()
		if err != nil {
			fmt.Printf("failed to load index: %v\n", err)
			os.Exit(1)
		}
	} else {
		// build a new index from documents
		fmt.Println("Building index...")

		idx = index.NewInvertedIndex()

		ingester := ingestion.NewIngester(analyzer, idx)

		if err := ingester.IngestDirectory("documents"); err != nil {
			fmt.Printf("failed to ingest documents: %v\n", err)
			os.Exit(1)
		}

		// persist the newly built index
		if err := store.SaveIndex(idx); err != nil {
			fmt.Printf("failed to save index: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Index saved to disk.")
	}

	stats := idx.Stats()

	fmt.Printf("Indexed %d documents\n", stats.Documents)
	fmt.Printf("Average document length: %.2f\n", stats.AverageDocLength)

	searcher := search.NewSearcher(idx)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nQuery: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		query := strings.TrimSpace(input)

		if query == "" {
			continue
		}

		if query == ":quit" {
			break
		}

		results := searcher.Search(query, 10)

		if len(results) == 0 {
			fmt.Println("No results found.")
			continue
		}

		fmt.Println("\nResults:")

		for rank, result := range results {
			fmt.Printf(
				"%d. %s (%.4f)\n",
				rank+1,
				result.DocID,
				result.Score,
			)
		}
	}
}
