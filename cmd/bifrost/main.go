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
)

func main() {
	analyzer := analysis.NewAnalyzer()
	idx := index.NewInvertedIndex()

	ingester := ingestion.NewIngester(analyzer, idx)

	if err := ingester.IngestDirectory("documents"); err != nil {
		fmt.Printf("failed to ingest documents: %v\n", err)
		os.Exit(1)
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
