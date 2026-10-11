package codec

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestEncodeDecodePostings(t *testing.T) {
	expected := []Posting{
		{
			DocID:     1,
			TermFreq:  2,
			Positions: []uint32{1, 5},
		},
		{
			DocID:     3,
			TermFreq:  1,
			Positions: []uint32{2},
		},
		{
			DocID:     130,
			TermFreq:  3,
			Positions: []uint32{0, 127, 300},
		},
	}

	encoded, err := EncodePostings(expected)
	if err != nil {
		t.Fatalf("EncodePostings() error = %v", err)
	}

	actual, err := DecodePostings(encoded)
	if err != nil {
		t.Fatalf("DecodePostings() error = %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("round trip mismatch:\ngot  %+v\nwant %+v", actual, expected)
	}
}

func TestEncodeDecodeEmptyPostings(t *testing.T) {
	encoded, err := EncodePostings(nil)
	if err != nil {
		t.Fatalf("EncodePostings() error = %v", err)
	}

	actual, err := DecodePostings(encoded)
	if err != nil {
		t.Fatalf("DecodePostings() error = %v", err)
	}

	if len(actual) != 0 {
		t.Fatalf("expected empty postings, got %+v", actual)
	}
}

func TestEncodePostingsRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		postings []Posting
	}{
		{
			name: "zero document ID",
			postings: []Posting{
				{DocID: 0, TermFreq: 1, Positions: []uint32{1}},
			},
		},
		{
			name: "unsorted document IDs",
			postings: []Posting{
				{DocID: 2, TermFreq: 1, Positions: []uint32{1}},
				{DocID: 1, TermFreq: 1, Positions: []uint32{2}},
			},
		},
		{
			name: "duplicate document IDs",
			postings: []Posting{
				{DocID: 1, TermFreq: 1, Positions: []uint32{1}},
				{DocID: 1, TermFreq: 1, Positions: []uint32{2}},
			},
		},
		{
			name: "zero term frequency",
			postings: []Posting{
				{DocID: 1, TermFreq: 0},
			},
		},
		{
			name: "frequency does not match positions",
			postings: []Posting{
				{DocID: 1, TermFreq: 2, Positions: []uint32{1}},
			},
		},
		{
			name: "unsorted positions",
			postings: []Posting{
				{DocID: 1, TermFreq: 2, Positions: []uint32{3, 2}},
			},
		},
		{
			name: "duplicate positions",
			postings: []Posting{
				{DocID: 1, TermFreq: 2, Positions: []uint32{2, 2}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := EncodePostings(tt.postings); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestDecodePostingsRejectsMalformedData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty input",
			data: nil,
		},
		{
			name: "truncated count varint",
			data: []byte{0x80},
		},
		{
			name: "truncated posting",
			data: []byte{0x01, 0x01},
		},
		{
			name: "zero document ID delta",
			data: []byte{0x01, 0x00, 0x01, 0x01, 0x01},
		},
		{
			name: "trailing bytes",
			data: []byte{0x00, 0x01},
		},
		{
			name: "overflowing varint",
			data: []byte{
				0x80, 0x80, 0x80, 0x80, 0x80,
				0x80, 0x80, 0x80, 0x80, 0x02,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := DecodePostings(tt.data); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestDecodePostingsRejectsTrailingData(t *testing.T) {
	encoded, err := EncodePostings([]Posting{
		{DocID: 1, TermFreq: 1, Positions: []uint32{0}},
	})
	if err != nil {
		t.Fatalf("EncodePostings() error = %v", err)
	}

	encoded = append(encoded, 0x01)

	_, err = DecodePostings(encoded)
	if !errors.Is(err, ErrTrailingData) {
		t.Fatalf("expected ErrTrailingData, got %v", err)
	}
}

func TestEncodePostingsDeterministic(t *testing.T) {
	postings := []Posting{
		{DocID: 2, TermFreq: 2, Positions: []uint32{1, 5}},
		{DocID: 10, TermFreq: 1, Positions: []uint32{3}},
	}

	first, err := EncodePostings(postings)
	if err != nil {
		t.Fatalf("first encode error = %v", err)
	}

	second, err := EncodePostings(postings)
	if err != nil {
		t.Fatalf("second encode error = %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("encoding should be deterministic")
	}
}

func FuzzDecodePostings(f *testing.F) {
	seeds := [][]byte{
		{0x00},                         // Empty posting list
		{0x01, 0x01, 0x01, 0x01, 0x00}, // One posting
		{0x80},                         // Truncated varint
		{},                             // Empty input
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		postings, err := DecodePostings(data)
		if err != nil {
			return
		}

		encoded, err := EncodePostings(postings)
		if err != nil {
			t.Fatalf("failed to re-encode decoded postings: %v", err)
		}

		decodedAgain, err := DecodePostings(encoded)
		if err != nil {
			t.Fatalf("failed to decode re-encoded postings: %v", err)
		}

		if !reflect.DeepEqual(postings, decodedAgain) {
			t.Fatalf("round trip mismatch: got %+v, want %+v", decodedAgain, postings)
		}
	})
}
