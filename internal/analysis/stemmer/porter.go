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

func isConsonant(word string, i int) bool {
	switch word[i] {
	case 'a', 'e', 'i', 'o', 'u':
		return false

	case 'y':
		if i == 0 {
			return true
		}

		return !isConsonant(word, i-1)

	default:
		return true
	}
}

func measure(word string) int {
	m := 0
	i := 0
	n := len(word)

	for i < n && isConsonant(word, i) {
		i++
	}

	for i < n {
		for i < n && !isConsonant(word, i) {
			i++
		}

		if i >= n {
			break
		}

		m++

		for i < n && isConsonant(word, i) {
			i++
		}
	}

	return m
}

func containsVowel(word string) bool {
	for i := 0; i < len(word); i++ {
		if !isConsonant(word, i) {
			return true
		}
	}

	return false
}

func endsWithDoubleConsonant(word string) bool {
	if len(word) < 2 {
		return false
	}

	last := len(word) - 1

	return word[last] == word[last-1] &&
		isConsonant(word, last)
}

func isCVC(word string, i int) bool {
	if i < 2 || !isConsonant(word, i) {
		return false
	}

	if isConsonant(word, i-1) || !isConsonant(word, i-2) {
		return false
	}

	switch word[i] {
	case 'w', 'x', 'y':
		return false
	}

	return true
}

func step1a(word string) string {
	switch {
	case strings.HasSuffix(word, "sses"):
		return strings.TrimSuffix(word, "sses") + "ss"

	case strings.HasSuffix(word, "ies"):
		return strings.TrimSuffix(word, "ies") + "i"

	case strings.HasSuffix(word, "ss"):
		return word

	case strings.HasSuffix(word, "s"):
		return strings.TrimSuffix(word, "s")

	default:
		return word
	}
}

func step1b(word string) string {
	if strings.HasSuffix(word, "eed") {
		stem := strings.TrimSuffix(word, "eed")

		if measure(stem) > 0 {
			return stem + "ee"
		}

		return word
	}

	if strings.HasSuffix(word, "ed") {
		stem := strings.TrimSuffix(word, "ed")

		if containsVowel(stem) {
			return finishStep1b(stem)
		}

		return word
	}

	if strings.HasSuffix(word, "ing") {
		stem := strings.TrimSuffix(word, "ing")

		if containsVowel(stem) {
			return finishStep1b(stem)
		}

		return word
	}

	return word
}

func finishStep1b(word string) string {
	switch {
	case strings.HasSuffix(word, "at"):
		return word + "e"

	case strings.HasSuffix(word, "bl"):
		return word + "e"

	case strings.HasSuffix(word, "iz"):
		return word + "e"
	}

	if endsWithDoubleConsonant(word) {
		last := word[len(word)-1]

		switch last {
		case 'l', 's', 'z':
			return word

		default:
			return word[:len(word)-1]
		}
	}

	if measure(word) == 1 && isCVC(word, len(word)-1) {
		return word + "e"
	}

	return word
}

func step1c(word string) string {
	if !strings.HasSuffix(word, "y") {
		return word
	}

	stem := word[:len(word)-1]

	if containsVowel(stem) {
		return stem + "i"
	}

	return word
}
