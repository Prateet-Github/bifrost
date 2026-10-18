<!-- markdownlint-configure-file {"MD024": {"siblings_only": true}} -->

# Changelog

All notable changes to Bifrost are documented in this file.

## [0.3.0] - 2026-10-18

- Add all benchmarks
- Analysis fixes

## [0.2.0] - 2026-10-16

### Added

- Title and body indexing with title boosting.
- Positional phrase search.
- Automatic index rebuilding when source documents change.
- Engine lifecycle tests.
- Search and storage benchmarks.

### Improved

- Compact posting encoding to reduce index storage size.
- Top-K retrieval using a min-heap.
- Query analysis through query-term deduplication.
- Index persistence and stale-posting cleanup.
- Documentation and benchmark reporting.

## [0.1.0]

### Added

- Initial lexical search engine implementation in Go.
- Unicode-aware tokenization and text analysis.
- Inverted index with positional postings.
- BM25-based document ranking.
- Persistent index storage using bbolt.
