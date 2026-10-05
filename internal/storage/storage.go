package storage

import (
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
