package storage

import (
	"bytes"
	"encoding/gob"

	"github.com/Prateet-Github/bifrost/internal/index"
	"go.etcd.io/bbolt"
)

var (
	termsBucket     = []byte("terms")
	documentsBucket = []byte("documents")
	metadataBucket  = []byte("metadata")
)

var (
	builtKey             = []byte("built")
	schemaVersionKey     = []byte("schema_version")
	sourceFingerprintKey = []byte("source_fingerprint")
)

const currentSchemaVersion = "1"

type Storage struct {
	db *bbolt.DB
}

func Open(path string) (*Storage, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(termsBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(documentsBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(metadataBucket); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SavePostings(term string, postings []index.Posting) error {
	data, err := encode(postings)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(termsBucket)

		return bucket.Put([]byte(term), data)
	})
}

func (s *Storage) LoadPostings(term string) ([]index.Posting, error) {
	var postings []index.Posting

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(termsBucket)

		data := bucket.Get([]byte(term))
		if data == nil {
			return nil
		}

		return decode(data, &postings)
	})

	return postings, err
}

func encode(value any) ([]byte, error) {
	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decode(data []byte, value any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(value)
}

func (s *Storage) SaveDocumentStats(
	docID string,
	stats index.DocumentStats,
) error {
	data, err := encode(stats)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(documentsBucket)

		return bucket.Put([]byte(docID), data)
	})
}

func (s *Storage) LoadDocumentStats(
	docID string,
) (index.DocumentStats, error) {
	var stats index.DocumentStats

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(documentsBucket)

		data := bucket.Get([]byte(docID))
		if data == nil {
			return nil
		}

		return decode(data, &stats)
	})

	return stats, err
}

func (s *Storage) SaveStats(stats index.Stats) error {
	data, err := encode(stats)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)

		return bucket.Put([]byte("stats"), data)
	})
}

func (s *Storage) LoadStats() (index.Stats, error) {
	var stats index.Stats

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)

		data := bucket.Get([]byte("stats"))
		if data == nil {
			return nil
		}

		return decode(data, &stats)
	})

	return stats, err
}

func (s *Storage) SaveIndex(idx *index.InvertedIndex, fingerprint string) error {
	termsData := idx.Terms()
	documentsData := idx.Documents()
	stats := idx.Stats()

	return s.db.Update(func(tx *bbolt.Tx) error {
		terms := tx.Bucket(termsBucket)
		documents := tx.Bucket(documentsBucket)
		metadata := tx.Bucket(metadataBucket)

		for term, postings := range termsData {
			data, err := encode(postings)
			if err != nil {
				return err
			}

			if err := terms.Put([]byte(term), data); err != nil {
				return err
			}
		}

		for docID, stats := range documentsData {
			data, err := encode(stats)
			if err != nil {
				return err
			}

			if err := documents.Put([]byte(docID), data); err != nil {
				return err
			}
		}

		data, err := encode(stats)
		if err != nil {
			return err
		}

		if err := metadata.Put([]byte("stats"), data); err != nil {
			return err
		}

		if err := metadata.Put(builtKey, []byte("true")); err != nil {
			return err
		}

		if err := metadata.Put(sourceFingerprintKey, []byte(fingerprint)); err != nil {
			return err
		}

		return nil
	})
}

func (s *Storage) LoadIndex() (*index.InvertedIndex, error) {
	idx := index.NewInvertedIndex()

	terms, err := s.loadAllTerms()
	if err != nil {
		return nil, err
	}

	for term, postings := range terms {
		for _, posting := range postings {
			idx.AddPosting(term, posting)
		}
	}

	documents, err := s.loadAllDocuments()
	if err != nil {
		return nil, err
	}

	for docID, stats := range documents {
		idx.SetDocumentStats(docID, stats)
	}

	return idx, nil
}

func (s *Storage) loadAllTerms() (map[string][]index.Posting, error) {
	terms := make(map[string][]index.Posting)

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(termsBucket)

		return bucket.ForEach(func(key, value []byte) error {
			var postings []index.Posting

			if err := decode(value, &postings); err != nil {
				return err
			}

			terms[string(key)] = postings

			return nil
		})
	})

	return terms, err
}

func (s *Storage) loadAllDocuments() (map[string]index.DocumentStats, error) {
	documents := make(map[string]index.DocumentStats)

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(documentsBucket)

		return bucket.ForEach(func(key, value []byte) error {
			var stats index.DocumentStats

			if err := decode(value, &stats); err != nil {
				return err
			}

			documents[string(key)] = stats

			return nil
		})
	})

	return documents, err
}

func (s *Storage) IsIndexReady() (bool, error) {
	var ready bool

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)

		built := bucket.Get(builtKey)
		version := bucket.Get(schemaVersionKey)

		if string(built) != "true" {
			return nil
		}

		if string(version) != currentSchemaVersion {
			return nil
		}

		ready = true
		return nil
	})

	return ready, err
}

func (s *Storage) SaveSourceFingerprint(fingerprint string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)
		return bucket.Put(sourceFingerprintKey, []byte(fingerprint))
	})
}

func (s *Storage) LoadSourceFingerprint() (string, error) {
	var fingerprint string

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)

		value := bucket.Get(sourceFingerprintKey)
		if value == nil {
			return nil
		}

		fingerprint = string(value)
		return nil
	})

	return fingerprint, err
}
