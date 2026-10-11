package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrInvalidDocID     = errors.New("document IDs must be positive and strictly increasing")
	ErrInvalidTermFreq  = errors.New("term frequency must be positive")
	ErrInvalidPositions = errors.New("positions must be strictly increasing")
)

func EncodePostings(postings []Posting) ([]byte, error) {
	// validate
	if err := validatePostings(postings); err != nil {
		return nil, err
	}

	// estimating a capacity
	buf := make([]byte, 0, len(postings)*8+1)

	buf = appendUvarint(buf, uint64(len(postings)))

	var previousDocID uint32

	for _, posting := range postings {
		// delta encoding
		docIDDelta := posting.DocID - previousDocID
		buf = appendUvarint(buf, uint64(docIDDelta))
		previousDocID = posting.DocID

		buf = appendUvarint(buf, uint64(posting.TermFreq))
		buf = appendUvarint(buf, uint64(len(posting.Positions)))

		var previousPosition uint32

		for _, position := range posting.Positions {
			positionDelta := position - previousPosition
			buf = appendUvarint(buf, uint64(positionDelta))
			previousPosition = position
		}
	}

	return buf, nil
}

// delta encoding assumes sorted values so validating it at 1st
func validatePostings(postings []Posting) error {
	var previousDocID uint32

	for i, posting := range postings {
		if posting.DocID == 0 || posting.DocID <= previousDocID {
			return fmt.Errorf("%w at posting %d", ErrInvalidDocID, i)
		}

		if posting.TermFreq == 0 {
			return fmt.Errorf("%w at posting %d", ErrInvalidTermFreq, i)
		}

		if uint64(posting.TermFreq) != uint64(len(posting.Positions)) {
			return fmt.Errorf(
				"posting %d: term frequency %d does not match position count %d",
				i,
				posting.TermFreq,
				len(posting.Positions),
			)
		}

		var previousPosition uint32

		for j, position := range posting.Positions {
			if j > 0 && position <= previousPosition {
				return fmt.Errorf(
					"%w at posting %d, position %d",
					ErrInvalidPositions,
					i,
					j,
				)
			}

			previousPosition = position
		}

		previousDocID = posting.DocID
	}

	return nil
}

func appendUvarint(buf []byte, value uint64) []byte {
	var encoded [binary.MaxVarintLen64]byte

	n := binary.PutUvarint(encoded[:], value)

	return append(buf, encoded[:n]...)
}
