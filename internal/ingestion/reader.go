package ingestion

import (
	"os"
	"path/filepath"

	"github.com/Prateet-Github/bifrost/internal/document"
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Read(path string) (document.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return document.Document{}, err
	}

	return document.Document{
		ID:    filepath.Base(path),
		Title: filepath.Base(path),
		Body:  string(data),
	}, nil
}
