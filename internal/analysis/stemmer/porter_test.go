package stemmer

import (
	"testing"

	"github.com/Prateet-Github/bifrost/internal/analysis/tokenizer"
)

func TestMeasure(t *testing.T) {
	tests := []struct {
		word     string
		expected int
	}{
		{"tr", 0},
		{"tree", 0},
		{"trouble", 1},
		{"troubles", 2},
		{"oats", 1},
		{"trees", 1},
		{"private", 2},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := measure(tt.word)

			if got != tt.expected {
				t.Errorf(
					"measure(%q) = %d, want %d",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestEndsWithDoubleConsonant(t *testing.T) {
	tests := []struct {
		word     string
		expected bool
	}{
		{"fall", true},
		{"pass", true},
		{"miss", true},
		{"buzz", true},
		{"sing", false},
		{"cat", false},
		{"a", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := endsWithDoubleConsonant(tt.word)

			if got != tt.expected {
				t.Errorf(
					"endsWithDoubleConsonant(%q) = %v, want %v",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestIsCVC(t *testing.T) {
	tests := []struct {
		word     string
		index    int
		expected bool
	}{
		{"hop", 2, true},
		{"cap", 2, true},
		{"bat", 2, true},
		{"box", 2, false},
		{"toy", 2, false},
		{"snow", 3, false},
		{"at", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := isCVC(tt.word, tt.index)

			if got != tt.expected {
				t.Errorf(
					"isCVC(%q, %d) = %v, want %v",
					tt.word,
					tt.index,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep1a(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"caresses", "caress"},
		{"ponies", "poni"},
		{"ties", "ti"},
		{"cats", "cat"},
		{"feed", "feed"},
		{"agreed", "agreed"},
		{"plastered", "plastered"},
		{"glass", "glass"},
		{"press", "press"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step1a(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step1a(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep1b(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"feed", "feed"},
		{"agreed", "agree"},
		{"plastered", "plaster"},
		{"bled", "bled"},
		{"motoring", "motor"},
		{"sing", "sing"},
		{"conflated", "conflate"},
		{"troubled", "trouble"},
		{"hopping", "hop"},
		{"tanned", "tan"},
		{"falling", "fall"},
		{"hissing", "hiss"},
		{"fizzed", "fizz"},
		{"failing", "fail"},
		{"filing", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step1b(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step1b(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep1c(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"happy", "happi"},
		{"sky", "sky"},
		{"cry", "cry"},
		{"toy", "toi"},
		{"boy", "boi"},
		{"say", "sai"},
		{"by", "by"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step1c(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step1c(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestReplaceSuffix(t *testing.T) {
	tests := []struct {
		word        string
		suffix      string
		replacement string
		expected    string
	}{
		{"relational", "ational", "ate", "relate"},
		{"conditional", "tional", "tion", "condition"},
		{"digitizer", "izer", "ize", "digitize"},
		{"happy", "ness", "ful", "happy"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := replaceSuffix(
				tt.word,
				tt.suffix,
				tt.replacement,
			)

			if got != tt.expected {
				t.Errorf(
					"replaceSuffix(%q, %q, %q) = %q, want %q",
					tt.word,
					tt.suffix,
					tt.replacement,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep2(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"relational", "relate"},
		{"conditional", "condition"},
		{"rational", "rational"},
		{"valenci", "valence"},
		{"hesitanci", "hesitance"},
		{"digitizer", "digitize"},
		{"conformabli", "conformable"},
		{"radicalli", "radical"},
		{"differentli", "different"},
		{"vileli", "vile"},
		{"analogousli", "analogous"},
		{"vietnamization", "vietnamize"},
		{"predication", "predicate"},
		{"operator", "operate"},
		{"feudalism", "feudal"},
		{"decisiveness", "decisive"},
		{"hopefulness", "hopeful"},
		{"callousness", "callous"},
		{"formaliti", "formal"},
		{"sensitiviti", "sensitive"},
		{"sensibiliti", "sensible"},
		{"triplicate", "triplicate"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step2(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step2(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep3(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"triplicate", "triplic"},
		{"formative", "form"},
		{"formalize", "formal"},
		{"electriciti", "electric"},
		{"electrical", "electric"},
		{"hopeful", "hope"},
		{"goodness", "good"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step3(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step3(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep4(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"revival", "reviv"},
		{"allowance", "allow"},
		{"inference", "infer"},
		{"airliner", "airlin"},
		{"gyroscopic", "gyroscop"},
		{"adjustable", "adjust"},
		{"defensible", "defens"},
		{"irritant", "irrit"},
		{"replacement", "replac"},
		{"adjustment", "adjust"},
		{"dependent", "depend"},
		{"homologou", "homolog"},
		{"communism", "commun"},
		{"activate", "activ"},
		{"angulariti", "angular"},
		{"homologous", "homolog"},
		{"effective", "effect"},
		{"bowdlerize", "bowdler"},
		{"adoption", "adopt"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step4(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step4(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep5a(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"probate", "probat"},
		{"rate", "rate"},
		{"cease", "ceas"},
		{"controll", "controll"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step5a(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step5a(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStep5b(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"controll", "control"},
		{"roll", "roll"},
		{"fall", "fall"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := step5b(tt.word)

			if got != tt.expected {
				t.Errorf(
					"step5b(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStem(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"caresses", "caress"},
		{"ponies", "poni"},
		{"ties", "ti"},
		{"cats", "cat"},
		{"feed", "feed"},
		{"agreed", "agre"},
		{"plastered", "plaster"},
		{"motoring", "motor"},
		{"conflated", "conflat"},
		{"troubled", "troubl"},
		{"hopping", "hop"},
		{"tanned", "tan"},
		{"happy", "happi"},
		{"relational", "relat"},
		{"conditional", "condit"},
		{"triplicate", "triplic"},
		{"formative", "form"},
		{"revival", "reviv"},
		{"allowance", "allow"},
		{"probate", "probat"},
		{"controll", "control"},
	}

	stemmer := NewPorterStemmer()

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := stemmer.Stem(tt.word)

			if got != tt.expected {
				t.Errorf(
					"Stem(%q) = %q, want %q",
					tt.word,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestStemTokenPreservesMetadata(t *testing.T) {
	stemmer := NewPorterStemmer()

	token := tokenizer.Token{
		Text:        "running",
		Position:    5,
		StartOffset: 20,
		EndOffset:   27,
	}

	result := stemmer.StemToken(token)

	if result.Text != "run" {
		t.Errorf("Text = %q, want %q", result.Text, "run")
	}

	if result.Position != 5 {
		t.Errorf("Position = %d, want 5", result.Position)
	}

	if result.StartOffset != 20 {
		t.Errorf("StartOffset = %d, want 20", result.StartOffset)
	}

	if result.EndOffset != 27 {
		t.Errorf("EndOffset = %d, want 27", result.EndOffset)
	}
}
