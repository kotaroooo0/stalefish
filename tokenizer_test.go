package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func TestMorphologicalTokenizerTokenize(t *testing.T) {
	kagome, err := NewKagome()
	if err != nil {
		t.Error("error: fail to initialize kagome tokenizer")
	}
	tok := NewMorphologicalTokenizer(kagome, tokenizer.Search)
	cases := []struct {
		sentence string
		expected *TokenStream
	}{
		{
			sentence: "Ishiuchi Maruyama",
			expected: NewTokenStream([]Token{
				NewToken("Ishiuchi", SetKana("Ishiuchi")),
				NewToken("Maruyama", SetKana("Maruyama")),
			}, Term),
		},
		{
			sentence: "石打丸山スキー場",
			expected: NewTokenStream([]Token{
				NewToken("石打丸山スキー場", SetKana("イシウチマルヤマスキージョウ")),
			}, Term),
		},
		{
			sentence: "石打丸山",
			expected: NewTokenStream([]Token{
				NewToken("石打丸山", SetKana("イシウチマルヤマ")),
			}, Term),
		},
		{
			sentence: "いしうちまるやま",
			expected: NewTokenStream([]Token{
				NewToken("い", SetKana("イ")),
				NewToken("し", SetKana("シ")),
				NewToken("うち", SetKana("ウチ")),
				NewToken("まるや", SetKana("マルヤ")),
				NewToken("ま", SetKana("マ")),
			}, Term),
		},
		{
			sentence: "イシウチ",
			expected: NewTokenStream([]Token{
				NewToken("イシウチ", SetKana("イシウチ")),
			}, Term),
		},
		{
			sentence: "白馬",
			expected: NewTokenStream([]Token{
				NewToken("白馬", SetKana("ハクバ")),
			}, Term),
		},
		{
			sentence: "白馬47",
			expected: NewTokenStream([]Token{
				NewToken("白馬", SetKana("ハクバ")),
				NewToken("47", SetKana("47")),
			}, Term),
		},
		{
			sentence: "琵琶湖バレイ",
			expected: NewTokenStream([]Token{
				NewToken("琵琶湖", SetKana("ビワコ")),
				NewToken("バレイ", SetKana("バレイ")),
			}, Term),
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tok.Tokenize(tt.sentence), tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
