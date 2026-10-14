package engine

import (
	"fmt"

	"github.com/Prateet-Github/bifrost/internal/analysis"
	"github.com/Prateet-Github/bifrost/internal/index"
	"github.com/Prateet-Github/bifrost/internal/ingestion"
	"github.com/Prateet-Github/bifrost/internal/search"
	"github.com/Prateet-Github/bifrost/internal/storage"
)

type Config struct {
	DBPath   string
	DocsPath string
}

type Engine struct {
	store    *storage.Storage
	searcher *search.Searcher
	stats    index.Stats
}

func Open(cfg Config) (*Engine, error) {
	analyzer := analysis.NewAnalyzer()

	store, err := storage.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage: %w", err)
	}

	idx, err := loadOrBuildIndex(store, analyzer, cfg.DocsPath)
	if err != nil {
		store.Close()
		return nil, err
	}

	return &Engine{
		store:    store,
		searcher: search.NewSearcher(idx),
		stats:    idx.Stats(),
	}, nil
}

func loadOrBuildIndex(
	store *storage.Storage,
	analyzer *analysis.Analyzer,
	docsPath string,
) (*index.InvertedIndex, error) {
	fingerprint, err := ingestion.ComputeDirFingerprint(docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute source fingerprint: %w", err)
	}

	ready, err := store.IsIndexReady()
	if err != nil {
		return nil, fmt.Errorf("failed to check index status: %w", err)
	}

	if ready {
		storedFingerprint, err := store.LoadSourceFingerprint()
		if err != nil {
			return nil, fmt.Errorf("failed to load source fingerprint: %w", err)
		}

		if storedFingerprint == fingerprint {
			fmt.Println("Loading index from disk...")

			idx, err := store.LoadIndex()
			if err != nil {
				return nil, fmt.Errorf("failed to load index: %w", err)
			}

			return idx, nil
		}

		fmt.Println("Source documents changed. Rebuilding index...")
	}

	fmt.Println("Building index...")

	idx := index.NewInvertedIndex()
	ingester := ingestion.NewIngester(analyzer, idx)

	if err := ingester.IngestDirectory(docsPath); err != nil {
		return nil, fmt.Errorf("failed to ingest documents: %w", err)
	}

	if err := store.SaveIndex(idx, fingerprint); err != nil {
		return nil, fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Println("Index saved to disk.")

	return idx, nil
}

func (e *Engine) Search(query string, k int) []search.Result {
	return e.searcher.Search(query, k)
}

func (e *Engine) SearchPhrase(phrase string, k int) []search.Result {
	return e.searcher.SearchPhrase(phrase, k)
}

func (e *Engine) Stats() index.Stats {
	return e.stats
}

func (e *Engine) Close() error {
	return e.store.Close()
}
