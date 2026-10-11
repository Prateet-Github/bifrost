package codec

type Posting struct {
	DocID     uint32
	TermFreq  uint32
	Positions []uint32
}
