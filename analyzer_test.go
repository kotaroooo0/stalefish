package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func TestAnalyze(t *testing.T) {
	kagome, err := NewKagome()
	if err != nil {
		t.Error("error: fail to initialize kagome tokenizer")
	}
	morphologicalTokenizer := NewMorphologicalTokenizer(kagome, tokenizer.Search)

	cases := []struct {
		analyzer Analyzer
		text     string
		tokens   *TokenStream
	}{
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			text:     "",
			tokens:   NewTokenStream([]Token{}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			text:     "a",
			tokens: NewTokenStream([]Token{
				NewToken("a"),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			text:     "small wild,cat!",
			tokens: NewTokenStream([]Token{
				NewToken("small"),
				NewToken("wild"),
				NewToken("cat"),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{LowercaseFilter{}}},
			text:     "I am BIG",
			tokens: NewTokenStream([]Token{
				NewToken("i"),
				NewToken("am"),
				NewToken("big"),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{StopWordFilter{}}},
			text:     "how a Big",
			tokens: NewTokenStream([]Token{
				NewToken("how"),
				NewToken("Big"),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{StemmerFilter{}}},
			text:     "Long pens",
			tokens: NewTokenStream([]Token{
				NewToken("long"),
				NewToken("pen"),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, morphologicalTokenizer, []TokenFilter{}},
			text:     "今日は天気が良い",
			tokens: NewTokenStream([]Token{
				NewToken("今日", SetKana("キョウ")),
				NewToken("は", SetKana("ハ")),
				NewToken("天気", SetKana("テンキ")),
				NewToken("が", SetKana("ガ")),
				NewToken("良い", SetKana("ヨイ")),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, morphologicalTokenizer, []TokenFilter{NewReadingformFilter(Kana)}},
			text:     "今日は天気が良い",
			tokens: NewTokenStream([]Token{
				NewToken("今日", SetKana("キョウ")),
				NewToken("は", SetKana("ハ")),
				NewToken("天気", SetKana("テンキ")),
				NewToken("が", SetKana("ガ")),
				NewToken("良い", SetKana("ヨイ")),
			}, 0),
		},
		{
			analyzer: Analyzer{[]CharFilter{}, morphologicalTokenizer, []TokenFilter{NewReadingformFilter(Romaji)}},
			text:     "今日は天気が良い",
			tokens: NewTokenStream([]Token{
				NewToken("今日", SetRomaji("kyo")),
				NewToken("は", SetRomaji("ha")),
				NewToken("天気", SetRomaji("tenki")),
				NewToken("が", SetRomaji("ga")),
				NewToken("良い", SetRomaji("yoi")),
			}, 0),
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tt.tokens, tt.analyzer.Analyze(tt.text)); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
