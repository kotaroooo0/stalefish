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
			analyzer: Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{}},
			text:     "",
			tokens:   []string{},
		},
		{
			analyzer: Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{}},
			text:     "a",
			tokens:   []string{"a"},
		},
		{
			analyzer: Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{}},
			text:     "small wild,cat!",
			tokens:   []string{"small", "wild", "cat"},
		},
		{
			analyzer: Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{LowercaseFilter{}}},
			text:     "I am BIG",
			tokens:   []string{"i", "am", "big"},
		},
		{
			analyzer: Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{StopWordFilter{}}},
			text:     "how a Big",
			tokens:   []string{"how", "Big"},
		},
		{
			analyzer: Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{StemmerFilter{}}},
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
