package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/Prateet-Github/bifrost/internal/codec"
	"github.com/Prateet-Github/bifrost/internal/index"
	"go.etcd.io/bbolt"
)

var (
	termsBucket      = []byte("terms")
	titleTermsBucket = []byte("title_terms")
	documentsBucket  = []byte("documents")
	docIDsBucket     = []byte("doc_ids")
	metadataBucket   = []byte("metadata")
)

var (
	builtKey             = []byte("built")
	schemaVersionKey     = []byte("schema_version")
	sourceFingerprintKey = []byte("source_fingerprint")
)

var compactPostingMagic = []byte("BCP1")

const currentSchemaVersion = "3"

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

		if _, err := tx.CreateBucketIfNotExists(titleTermsBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(documentsBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(docIDsBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(metadataBucket); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) LoadPostings(term string) ([]index.Posting, error) {
	var postings []index.Posting

	err := s.db.View(func(tx *bbolt.Tx) error {
		terms := tx.Bucket(termsBucket)
		data := terms.Get([]byte(term))

		if data == nil {
			return nil
		}

		if bytes.HasPrefix(data, compactPostingMagic) {
			docIDs := tx.Bucket(docIDsBucket)
			if docIDs == nil {
				return fmt.Errorf("document ID bucket not found")
			}

			numericToDocument, err := loadDocumentIDMap(docIDs)
			if err != nil {
				return err
			}

			decoded, err := codec.DecodePostings(
				data[len(compactPostingMagic):],
			)
			if err != nil {
				return fmt.Errorf(
					"decode compact postings for term %q: %w",
					term,
					err,
				)
			}

			postings = make([]index.Posting, 0, len(decoded))

			for _, posting := range decoded {
				docID, exists := numericToDocument[posting.DocID]
				if !exists {
					return fmt.Errorf(
						"numeric document ID %d missing from mapping",
						posting.DocID,
					)
				}

				positions := make([]int, len(posting.Positions))
				for i, position := range posting.Positions {
					positions[i] = int(position)
				}

				postings = append(postings, index.Posting{
					DocID:     docID,
					TermFreq:  int(posting.TermFreq),
					Positions: positions,
				})
			}

			return nil
		}

		if err := decode(data, &postings); err != nil {
			return fmt.Errorf(
				"decode postings for term %q: %w",
				term,
				err,
			)
		}

		return nil
	})

	return postings, err
}

func (s *Storage) LoadTitlePostings(term string) ([]index.Posting, error) {
	var postings []index.Posting

	err := s.db.View(func(tx *bbolt.Tx) error {
		titleTerms := tx.Bucket(titleTermsBucket)
		if titleTerms == nil {
			return fmt.Errorf("title terms bucket not found")
		}

		data := titleTerms.Get([]byte(term))
		if data == nil {
			return nil
		}

		if !bytes.HasPrefix(data, compactPostingMagic) {
			return fmt.Errorf(
				"title term %q does not use the compact posting format",
				term,
			)
		}

		docIDs := tx.Bucket(docIDsBucket)
		if docIDs == nil {
			return fmt.Errorf("document ID bucket not found")
		}

		numericToDocument, err := loadDocumentIDMap(docIDs)
		if err != nil {
			return err
		}

		decoded, err := codec.DecodePostings(
			data[len(compactPostingMagic):],
		)
		if err != nil {
			return fmt.Errorf(
				"decode title postings for term %q: %w",
				term,
				err,
			)
		}

		postings = make([]index.Posting, 0, len(decoded))

		for _, posting := range decoded {
			docID, exists := numericToDocument[posting.DocID]
			if !exists {
				return fmt.Errorf(
					"numeric document ID %d missing from mapping",
					posting.DocID,
				)
			}

			positions := make([]int, len(posting.Positions))
			for i, position := range posting.Positions {
				positions[i] = int(position)
			}

			postings = append(postings, index.Posting{
				DocID:     docID,
				TermFreq:  int(posting.TermFreq),
				Positions: positions,
			})
		}

		return nil
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

func (s *Storage) SaveIndex(
	idx *index.InvertedIndex,
	fingerprint string,
) error {
	termsData := idx.Terms()
	titleTermsData := idx.TitleTerms()
	documentsData := idx.Documents()

	documentIDs := make([]string, 0, len(documentsData))
	for docID := range documentsData {
		documentIDs = append(documentIDs, docID)
	}

	documentToNumeric, numericToDocument, err :=
		buildDocumentIDMaps(documentIDs)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(termsBucket); err != nil &&
			err != bbolt.ErrBucketNotFound {
			return err
		}

		if _, err := tx.CreateBucket(termsBucket); err != nil {
			return err
		}

		if err := tx.DeleteBucket(titleTermsBucket); err != nil &&
			err != bbolt.ErrBucketNotFound {
			return err
		}

		if _, err := tx.CreateBucket(titleTermsBucket); err != nil {
			return err
		}

		if err := tx.DeleteBucket(documentsBucket); err != nil &&
			err != bbolt.ErrBucketNotFound {
			return err
		}

		if _, err := tx.CreateBucket(documentsBucket); err != nil {
			return err
		}

		if err := tx.DeleteBucket(docIDsBucket); err != nil &&
			err != bbolt.ErrBucketNotFound {
			return err
		}

		if _, err := tx.CreateBucket(docIDsBucket); err != nil {
			return err
		}

		terms := tx.Bucket(termsBucket)
		titleTerms := tx.Bucket(titleTermsBucket)
		documents := tx.Bucket(documentsBucket)
		docIDs := tx.Bucket(docIDsBucket)
		metadata := tx.Bucket(metadataBucket)

		terms.FillPercent = 1.0
		titleTerms.FillPercent = 1.0
		documents.FillPercent = 1.0

		// save body terms
		if err := saveTerms(
			terms,
			termsData,
			documentToNumeric,
		); err != nil {
			return err
		}

		// save title terms
		if err := saveTerms(
			titleTerms,
			titleTermsData,
			documentToNumeric,
		); err != nil {
			return err
		}

		// save doc stats
		for _, docID := range documentIDs {
			data, err := encode(documentsData[docID])
			if err != nil {
				return err
			}

			if err := documents.Put([]byte(docID), data); err != nil {
				return err
			}
		}

		numericIDs := make([]uint32, 0, len(numericToDocument))
		for numericID := range numericToDocument {
			numericIDs = append(numericIDs, numericID)
		}

		sort.Slice(numericIDs, func(i, j int) bool {
			return numericIDs[i] < numericIDs[j]
		})

		for _, numericID := range numericIDs {
			var key [4]byte
			binary.BigEndian.PutUint32(key[:], numericID)

			docID := numericToDocument[numericID]

			if err := docIDs.Put(key[:], []byte(docID)); err != nil {
				return err
			}
		}

		if err := metadata.Put(builtKey, []byte("true")); err != nil {
			return err
		}

		if err := metadata.Put(
			schemaVersionKey,
			[]byte(currentSchemaVersion),
		); err != nil {
			return err
		}

		if err := metadata.Put(
			sourceFingerprintKey,
			[]byte(fingerprint),
		); err != nil {
			return err
		}

		return nil
	})
}

// saveTerms writes one field's postings using the compact posting codec
func saveTerms(
	bucket *bbolt.Bucket,
	termsData map[string][]index.Posting,
	documentToNumeric map[string]uint32,
) error {
	termKeys := make([]string, 0, len(termsData))
	for term := range termsData {
		termKeys = append(termKeys, term)
	}

	sort.Strings(termKeys)

	for _, term := range termKeys {
		compactPostings, err := toCodecPostings(
			termsData[term],
			documentToNumeric,
		)
		if err != nil {
			return err
		}

		data, err := codec.EncodePostings(compactPostings)
		if err != nil {
			return err
		}

		data = append(
			append([]byte(nil), compactPostingMagic...),
			data...,
		)

		if err := bucket.Put([]byte(term), data); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) LoadIndex() (*index.InvertedIndex, error) {
	idx := index.NewInvertedIndex()

	// load the numeric-to-string document ID mapping first
	var numericToDocument map[uint32]string

	err := s.db.View(func(tx *bbolt.Tx) error {
		docIDs := tx.Bucket(docIDsBucket)
		if docIDs == nil {
			return fmt.Errorf("document ID bucket not found")
		}

		var err error
		numericToDocument, err = loadDocumentIDMap(docIDs)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Restore body postings
	terms, err := s.loadAllTerms(termsBucket, numericToDocument)
	if err != nil {
		return nil, err
	}

	for term, postings := range terms {
		for _, posting := range postings {
			idx.AddPosting(term, posting)
		}
	}

	titleTerms, err := s.loadAllTerms(titleTermsBucket, numericToDocument)
	if err != nil {
		return nil, err
	}

	for term, postings := range titleTerms {
		for _, posting := range postings {
			idx.AddTitlePosting(term, posting)
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

func (s *Storage) loadAllTerms(
	bucketName []byte,
	numericToDocument map[uint32]string,
) (map[string][]index.Posting, error) {
	terms := make(map[string][]index.Posting)

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf(
				"terms bucket %q not found",
				string(bucketName),
			)
		}

		return bucket.ForEach(func(key, value []byte) error {
			if !bytes.HasPrefix(value, compactPostingMagic) {
				return fmt.Errorf(
					"term %q does not use the compact posting format",
					string(key),
				)
			}

			decoded, err := codec.DecodePostings(
				value[len(compactPostingMagic):],
			)
			if err != nil {
				return fmt.Errorf(
					"decode postings for term %q: %w",
					string(key),
					err,
				)
			}

			postings := make([]index.Posting, 0, len(decoded))

			for _, posting := range decoded {
				docID, exists := numericToDocument[posting.DocID]
				if !exists {
					return fmt.Errorf(
						"numeric document ID %d missing from mapping",
						posting.DocID,
					)
				}

				positions := make([]int, len(posting.Positions))
				for i, position := range posting.Positions {
					positions[i] = int(position)
				}

				postings = append(postings, index.Posting{
					DocID:     docID,
					TermFreq:  int(posting.TermFreq),
					Positions: positions,
				})
			}

			terms[string(key)] = postings
			return nil
		})
	})

	return terms, err
}

func (s *Storage) loadAllDocuments() (
	map[string]index.DocumentStats,
	error,
) {
	documents := make(map[string]index.DocumentStats)

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(documentsBucket)
		if bucket == nil {
			return fmt.Errorf("documents bucket not found")
		}

		return bucket.ForEach(func(key, value []byte) error {
			var stats index.DocumentStats

			if err := decode(value, &stats); err != nil {
				return fmt.Errorf(
					"decode document stats for %q: %w",
					string(key),
					err,
				)
			}

			documents[string(key)] = stats
			return nil
		})
	})

	return documents, err
}

func loadDocumentIDMap(
	bucket *bbolt.Bucket,
) (map[uint32]string, error) {
	numericToDocument := make(map[uint32]string)

	err := bucket.ForEach(func(key, value []byte) error {
		if len(key) != 4 {
			return fmt.Errorf(
				"invalid numeric document ID key length: %d",
				len(key),
			)
		}

		numericID := binary.BigEndian.Uint32(key)

		if numericID == 0 {
			return fmt.Errorf("numeric document ID must be positive")
		}

		if len(value) == 0 {
			return fmt.Errorf(
				"empty document ID for numeric ID %d",
				numericID,
			)
		}

		if _, exists := numericToDocument[numericID]; exists {
			return fmt.Errorf(
				"duplicate numeric document ID %d",
				numericID,
			)
		}

		numericToDocument[numericID] = string(value)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return numericToDocument, nil
}

func (s *Storage) IsIndexReady() (bool, error) {
	var ready bool

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(metadataBucket)
		if bucket == nil {
			return fmt.Errorf("metadata bucket not found")
		}

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
		if bucket == nil {
			return fmt.Errorf("metadata bucket not found")
		}

		value := bucket.Get(sourceFingerprintKey)
		if value == nil {
			return nil
		}

		fingerprint = string(value)
		return nil
	})

	return fingerprint, err
}

func buildDocumentIDMaps(
	documentIDs []string,
) (map[string]uint32, map[uint32]string, error) {
	if uint64(len(documentIDs)) > uint64(^uint32(0)) {
		return nil, nil, fmt.Errorf("too many documents for uint32 IDs")
	}

	sort.Strings(documentIDs)

	for i := 1; i < len(documentIDs); i++ {
		if documentIDs[i] == documentIDs[i-1] {
			return nil, nil, fmt.Errorf(
				"duplicate document ID %q",
				documentIDs[i],
			)
		}
	}

	forward := make(map[string]uint32, len(documentIDs))
	reverse := make(map[uint32]string, len(documentIDs))

	for i, docID := range documentIDs {
		numericID := uint32(i + 1)
		forward[docID] = numericID
		reverse[numericID] = docID
	}

	return forward, reverse, nil
}

func toCodecPostings(
	postings []index.Posting,
	documentIDs map[string]uint32,
) ([]codec.Posting, error) {
	converted := make([]codec.Posting, 0, len(postings))

	for _, posting := range postings {
		numericID, exists := documentIDs[posting.DocID]
		if !exists {
			return nil, fmt.Errorf(
				"document ID %q missing from mapping",
				posting.DocID,
			)
		}

		if posting.TermFreq <= 0 ||
			uint64(posting.TermFreq) > uint64(^uint32(0)) {
			return nil, fmt.Errorf(
				"invalid term frequency %d for document %q",
				posting.TermFreq,
				posting.DocID,
			)
		}

		positions := make([]uint32, len(posting.Positions))

		for i, position := range posting.Positions {
			if position < 0 ||
				uint64(position) > uint64(^uint32(0)) {
				return nil, fmt.Errorf(
					"invalid position %d for document %q",
					position,
					posting.DocID,
				)
			}

			positions[i] = uint32(position)
		}

		converted = append(converted, codec.Posting{
			DocID:     numericID,
			TermFreq:  uint32(posting.TermFreq),
			Positions: positions,
		})
	}

	sort.Slice(converted, func(i, j int) bool {
		return converted[i].DocID < converted[j].DocID
	})

	return converted, nil
}
