package ingestion

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ComputeDirFingerprint(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	var records []string

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txt" {
			continue
		}

		path := filepath.Join(dirPath, entry.Name())

		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}

		record := fmt.Sprintf(
			"%s|%d|%d",
			entry.Name(),
			info.Size(),
			info.ModTime().UnixNano(),
		)

		records = append(records, record)
	}

	sort.Strings(records)

	hash := sha256.Sum256([]byte(strings.Join(records, "\n")))

	return hex.EncodeToString(hash[:]), nil
}
