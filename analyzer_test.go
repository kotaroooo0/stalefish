package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()

	morphologicalTokenizer := NewMorphologicalTokenizer(NewKagomeMock())
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
			analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewStopWordFilter()}},
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
		{
			analyzer: Analyzer{[]CharFilter{}, morphologicalTokenizer, []TokenFilter{}},
			text:     "今日は天気が良い",
			tokens: NewTokenStream([]Token{
				NewToken("今日", setKana("キョウ")),
				NewToken("は", setKana("ハ")),
				NewToken("天気", setKana("テンキ")),
				NewToken("が", setKana("ガ")),
				NewToken("良い", setKana("ヨイ")),
			}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, morphologicalTokenizer, []TokenFilter{NewReadingformFilter(Kana)}},
			text:     "今日は天気が良い",
			tokens: NewTokenStream([]Token{
				NewToken("キョウ", setKana("キョウ")),
				NewToken("ハ", setKana("ハ")),
				NewToken("テンキ", setKana("テンキ")),
				NewToken("ガ", setKana("ガ")),
				NewToken("ヨイ", setKana("ヨイ")),
			}),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, morphologicalTokenizer, []TokenFilter{NewReadingformFilter(Romaji)}},
			text:     "今日は天気が良い",
			tokens: NewTokenStream([]Token{
				NewToken("kyo", setKana("キョウ")),
				NewToken("ha", setKana("ハ")),
				NewToken("tenki", setKana("テンキ")),
				NewToken("ga", setKana("ガ")),
				NewToken("yoi", setKana("ヨイ")),
			}),
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tt.tokens, tt.analyzer.Analyze(tt.text)); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
