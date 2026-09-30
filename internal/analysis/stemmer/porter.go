package stemmer

import (
	"strings"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

type PorterStemmer struct{}
type suffixRule struct {
	suffix      string
	replacement string
}

func NewPorterStemmer() *PorterStemmer {
	return &PorterStemmer{}
}

func (s *PorterStemmer) Stem(word string) string {
	word = strings.ToLower(word)

	if len(word) <= 2 {
		return word
	}

	word = step1a(word)
	word = step1b(word)
	word = step1c(word)

	word = step2(word)
	word = step3(word)
	word = step4(word)

	word = step5a(word)
	word = step5b(word)

	return word
}

func (s *PorterStemmer) StemToken(token tokenizer.Token) tokenizer.Token {
	token.Text = s.Stem(token.Text)
	return token
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

func replaceSuffix(word, suffix, replacement string) string {
	if !strings.HasSuffix(word, suffix) {
		return word
	}

	return strings.TrimSuffix(word, suffix) + replacement
}

var step2Rules = []suffixRule{
	{"ational", "ate"},
	{"tional", "tion"},
	{"enci", "ence"},
	{"anci", "ance"},
	{"izer", "ize"},
	{"bli", "ble"},
	{"alli", "al"},
	{"entli", "ent"},
	{"eli", "e"},
	{"ousli", "ous"},
	{"ization", "ize"},
	{"ation", "ate"},
	{"ator", "ate"},
	{"alism", "al"},
	{"iveness", "ive"},
	{"fulness", "ful"},
	{"ousness", "ous"},
	{"aliti", "al"},
	{"iviti", "ive"},
	{"biliti", "ble"},
	{"logi", "log"},
}

func step2(word string) string {
	for _, rule := range step2Rules {
		if !strings.HasSuffix(word, rule.suffix) {
			continue
		}

		stem := strings.TrimSuffix(word, rule.suffix)

		if measure(stem) > 0 {
			return stem + rule.replacement
		}

		return word
	}

	return word
}

var step3Rules = []suffixRule{
	{"icate", "ic"},
	{"ative", ""},
	{"alize", "al"},
	{"iciti", "ic"},
	{"ical", "ic"},
	{"ful", ""},
	{"ness", ""},
}

func step3(word string) string {
	for _, rule := range step3Rules {
		if !strings.HasSuffix(word, rule.suffix) {
			continue
		}

		stem := strings.TrimSuffix(word, rule.suffix)

		if measure(stem) > 0 {
			return stem + rule.replacement
		}

		return word
	}

	return word
}

var step4Suffixes = []string{
	"al",
	"ance",
	"ence",
	"er",
	"ic",
	"able",
	"ible",
	"ant",
	"ement",
	"ment",
	"ent",
	"ou",
	"ism",
	"ate",
	"iti",
	"ous",
	"ive",
	"ize",
}

func step4(word string) string {
	for _, suffix := range step4Suffixes {
		if !strings.HasSuffix(word, suffix) {
			continue
		}

		stem := strings.TrimSuffix(word, suffix)

		if measure(stem) > 1 {
			return stem
		}

		return word
	}

	if strings.HasSuffix(word, "ion") {
		stem := strings.TrimSuffix(word, "ion")

		if measure(stem) > 1 &&
			(strings.HasSuffix(stem, "s") ||
				strings.HasSuffix(stem, "t")) {
			return stem
		}
	}

	return word
}

func step5a(word string) string {
	if !strings.HasSuffix(word, "e") {
		return word
	}

	stem := strings.TrimSuffix(word, "e")
	m := measure(stem)

	if m > 1 {
		return stem
	}

	if m == 1 && !isCVC(stem, len(stem)-1) {
		return stem
	}

	return word
}

func step5b(word string) string {
	if !strings.HasSuffix(word, "ll") {
		return word
	}

	if measure(word) > 1 {
		return word[:len(word)-1]
	}

	return word
}
