package normalizer

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) Normalize(token tokenizer.Token) tokenizer.Token {
	token.Text = strings.ToLower(token.Text)
	token.Text = foldAccents(token.Text)

	return token
}

func foldAccents(text string) string {
	decomposed := norm.NFD.String(text)

	var builder strings.Builder

	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		builder.WriteRune(r)
	}

	return builder.String()
}
