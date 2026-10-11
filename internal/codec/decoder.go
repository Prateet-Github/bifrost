package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrMalformedData = errors.New("malformed posting data")
	ErrTrailingData  = errors.New("trailing data after postings")
)

func DecodePostings(data []byte) ([]Posting, error) {
	if len(data) == 0 {
		return nil, ErrMalformedData
	}

	count, n := binary.Uvarint(data)
	if n <= 0 {
		return nil, ErrMalformedData
	}
	data = data[n:]

	if count > uint64(len(data))/4 {
		if count != 0 || len(data) != 0 {
			return nil, fmt.Errorf(
				"%w: impossible posting count %d",
				ErrMalformedData,
				count,
			)
		}
	}

	postings := make([]Posting, 0, int(count))
	var previousDocID uint32

	for i := uint64(0); i < count; i++ {
		docIDDelta, rest, err := readUvarint(data)
		if err != nil {
			return nil, err
		}
		data = rest

		termFreq, rest, err := readUvarint(data)
		if err != nil {
			return nil, err
		}
		data = rest

		positionCount, rest, err := readUvarint(data)
		if err != nil {
			return nil, err
		}
		data = rest

		if docIDDelta == 0 || docIDDelta > uint64(^uint32(0))-uint64(previousDocID) {
			return nil, fmt.Errorf("%w: invalid document ID delta", ErrMalformedData)
		}

		docID := previousDocID + uint32(docIDDelta)

		if termFreq == 0 || termFreq > uint64(^uint32(0)) {
			return nil, fmt.Errorf("%w: invalid term frequency", ErrMalformedData)
		}

		if positionCount != termFreq {
			return nil, fmt.Errorf("%w: term frequency and position count differ", ErrMalformedData)
		}

		// at least one byte
		if positionCount > uint64(len(data)) {
			return nil, fmt.Errorf("%w: impossible position count", ErrMalformedData)
		}

		positions := make([]uint32, 0, int(positionCount))
		var previousPosition uint32

		for j := uint64(0); j < positionCount; j++ {
			delta, rest, err := readUvarint(data)
			if err != nil {
				return nil, err
			}
			data = rest

			if delta > uint64(^uint32(0))-uint64(previousPosition) {
				return nil, fmt.Errorf("%w: position overflow", ErrMalformedData)
			}

			position := previousPosition + uint32(delta)

			if j > 0 && position <= previousPosition {
				return nil, fmt.Errorf("%w: positions are not increasing", ErrMalformedData)
			}

			positions = append(positions, position)
			previousPosition = position
		}

		postings = append(postings, Posting{
			DocID:     docID,
			TermFreq:  uint32(termFreq),
			Positions: positions,
		})

		previousDocID = docID
	}

	if len(data) != 0 {
		return nil, ErrTrailingData
	}

	return postings, nil
}

func readUvarint(data []byte) (uint64, []byte, error) {
	value, n := binary.Uvarint(data)

	if n == 0 {
		return 0, nil, fmt.Errorf("%w: truncated varint", ErrMalformedData)
	}

	if n < 0 {
		return 0, nil, fmt.Errorf("%w: overflowing varint", ErrMalformedData)
	}

	return value, data[n:], nil
}
