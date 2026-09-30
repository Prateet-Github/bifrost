package stemmer

import "strings"

type PorterStemmer struct{}

func NewPorterStemmer() *PorterStemmer {
	return &PorterStemmer{}
}

func (s *PorterStemmer) Stem(word string) string {
	word = strings.ToLower(word)

	if len(word) <= 2 {
		return word
	}

	return word
}
