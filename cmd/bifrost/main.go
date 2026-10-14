package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Prateet-Github/bifrost/internal/engine"
	"github.com/Prateet-Github/bifrost/internal/search"
)

const (
	dbPath        = "bifrost.db"
	documentsPath = "documents"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	bifrost, err := engine.Open(engine.Config{
		DBPath:   dbPath,
		DocsPath: documentsPath,
	})
	if err != nil {
		return err
	}
	defer bifrost.Close()

	stats := bifrost.Stats()

	fmt.Printf("Indexed %d documents\n", stats.Documents)
	fmt.Printf("Average document length: %.2f\n", stats.AverageDocLength)

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

		var results []search.Result

		if strings.HasPrefix(query, `"`) && strings.HasSuffix(query, `"`) && len(query) >= 2 {
			phrase := query[1 : len(query)-1]
			results = bifrost.SearchPhrase(phrase, 10)
		} else {
			results = bifrost.Search(query, 10)
		}

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

	return nil
}
