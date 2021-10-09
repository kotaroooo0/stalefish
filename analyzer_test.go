package stalefish

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		analyzer Analyzer
		text     string
		tokens   *TokenStream
	}{
		{
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
			text:     "",
			tokens:   NewTokenStream([]Token{}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
			text:     "a",
			tokens: NewTokenStream([]Token{
				NewToken("a"),
			}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
			text:     "small wild,cat!",
			tokens: NewTokenStream([]Token{
				NewToken("small"),
				NewToken("wild"),
				NewToken("cat"),
			}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewLowercaseFilter()}},
			text:     "I am BIG",
			tokens: NewTokenStream([]Token{
				NewToken("i"),
				NewToken("am"),
				NewToken("big"),
			}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewStopWordFilter([]string{"a"})}},
			text:     "how a Big",
			tokens: NewTokenStream([]Token{
				NewToken("how"),
				NewToken("Big"),
			}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewStemmerFilter()}},
			text:     "Long pens",
			tokens: NewTokenStream([]Token{
				NewToken("long"),
				NewToken("pen"),
			}),
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("analyzer = %v, text = %v, tokens = %v", tt.analyzer, tt.text, tt.tokens), func(t *testing.T) {
			if diff := cmp.Diff(tt.tokens, tt.analyzer.Analyze(tt.text)); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}
