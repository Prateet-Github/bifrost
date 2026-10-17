package ingestion

import (
	"os"
	"path/filepath"
	"strings"

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

	body := string(data)
	filename := filepath.Base(path)

	return document.Document{
		ID:    filename,
		Title: deriveTitle(body, filename),
		Body:  body,
	}, nil
}

func deriveTitle(body, filename string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			runes := []rune(line)
			if len(runes) > 120 {
				line = string(runes[:120])
			}
			return line
		}
	}

	title := strings.TrimSuffix(filename, filepath.Ext(filename))
	title = strings.NewReplacer("-", " ", "_", " ").Replace(title)
	return title
}
