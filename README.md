# Bifrost

Bifrost is a search engine written from scratch in Go, built to understand and implement the core components behind traditional lexical search.

Analysis Pipeline

The analysis pipeline takes parsed document text and produces analyzed tokens.

```text
Document
   ↓
Tokenization
   ↓
Normalization
   ↓
Filtering
   ↓
Stemming
   ↓
Inverted Index
   ↓
BM25 Ranking
```

It includes:

* Tokenization — splits text into tokens while preserving positions and offsets.
* Normalization — lowercases and normalizes text.
* Filtering — removes stop words.
* Stemming — applies the Porter stemming algorithm.

## Search

The current search pipeline includes:

* Inverted index
* Term frequency and document statistics
* Query analysis
* Candidate generation
* BM25 scoring
* Top-K ranked results
* Two-word phrase matching using token positions
