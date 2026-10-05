# Bifrost

Bifrost is a search engine written from scratch in Go, built to understand the core components behind traditional lexical search.

```text
Documents
    ↓
Ingestion
    ↓
Tokenization
    ↓
Normalization
    ↓
Filtering
    ↓
Porter Stemming
    ↓
Inverted Index
    ↓
bbolt Persistence
```

## Analysis

- Unicode-aware tokenization with positions and offsets
- Lowercasing and accent normalization
- Stop-word filtering
- Porter stemming
- Positional inverted index

## Search

```text
Query
  ↓
Query Analysis
  ↓
Candidate Generation
  ↓
BM25
  ↓
Top-K Results
```

- BM25 ranking
- Term/document statistics
- Positional phrase matching
- Deterministic ranking
- Regression-tested search pipeline

## Persistence

Uses bbolt for persistent index storage, including postings, document statistics, and index restoration across restarts.

## Tech Stack

- Go
- bbolt
