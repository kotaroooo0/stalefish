package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		analyzer Analyzer
		text     string
		tokens   []string
	}{
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			text:     "",
			tokens:   []string{},
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			text:     "a",
			tokens:   []string{"a"},
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			text:     "small wild,cat!",
			tokens:   []string{"small", "wild", "cat"},
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{LowercaseFilter{}}},
			text:     "I am BIG",
			tokens:   []string{"i", "am", "big"},
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{StopWordFilter{}}},
			text:     "how a Big",
			tokens:   []string{"how", "Big"},
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{StemmerFilter{}}},
			text:     "Long pens",
			tokens:   []string{"long", "pen"},
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tt.tokens, tt.analyzer.Analyze(tt.text)); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
