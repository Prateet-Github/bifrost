package ingestion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeDirFingerprint(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "go.txt")
	if err := os.WriteFile(filePath, []byte("Go is fast"), 0644); err != nil {
		t.Fatal(err)
	}

	first, err := ComputeDirFingerprint(dir)
	if err != nil {
		t.Fatalf("first fingerprint failed: %v", err)
	}

	second, err := ComputeDirFingerprint(dir)
	if err != nil {
		t.Fatalf("second fingerprint failed: %v", err)
	}

	if first != second {
		t.Fatal("fingerprints should be identical for an unchanged directory")
	}
}

func TestComputeDirFingerprintDetectsChanges(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "go.txt")
	if err := os.WriteFile(filePath, []byte("Go is fast"), 0644); err != nil {
		t.Fatal(err)
	}

	original, err := ComputeDirFingerprint(dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("adding txt file", func(t *testing.T) {
		newFile := filepath.Join(dir, "rust.txt")
		if err := os.WriteFile(newFile, []byte("Rust is safe"), 0644); err != nil {
			t.Fatal(err)
		}

		updated, err := ComputeDirFingerprint(dir)
		if err != nil {
			t.Fatal(err)
		}

		if original == updated {
			t.Fatal("fingerprint should change when a .txt file is added")
		}

		if err := os.Remove(newFile); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("modifying txt file", func(t *testing.T) {
		if err := os.WriteFile(filePath, []byte("Go is fast and efficient"), 0644); err != nil {
			t.Fatal(err)
		}

		updated, err := ComputeDirFingerprint(dir)
		if err != nil {
			t.Fatal(err)
		}

		if original == updated {
			t.Fatal("fingerprint should change when a .txt file is modified")
		}
	})

	t.Run("deleting txt file", func(t *testing.T) {
		if err := os.Remove(filePath); err != nil {
			t.Fatal(err)
		}

		updated, err := ComputeDirFingerprint(dir)
		if err != nil {
			t.Fatal(err)
		}

		if original == updated {
			t.Fatal("fingerprint should change when a .txt file is deleted")
		}
	})
}

func TestComputeDirFingerprintIgnoresNonTextFiles(t *testing.T) {
	dir := t.TempDir()

	txtFile := filepath.Join(dir, "go.txt")
	if err := os.WriteFile(txtFile, []byte("Go is fast"), 0644); err != nil {
		t.Fatal(err)
	}

	before, err := ComputeDirFingerprint(dir)
	if err != nil {
		t.Fatal(err)
	}

	otherFile := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(otherFile, []byte("Some notes"), 0644); err != nil {
		t.Fatal(err)
	}

	after, err := ComputeDirFingerprint(dir)
	if err != nil {
		t.Fatal(err)
	}

	if before != after {
		t.Fatal("fingerprint should not change when a non-.txt file is added")
	}
}

func TestComputeDirFingerprintMissingDirectory(t *testing.T) {
	_, err := ComputeDirFingerprint(filepath.Join(t.TempDir(), "missing"))

	if err == nil {
		t.Fatal("expected an error for a nonexistent directory")
	}
}
