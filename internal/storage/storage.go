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
