package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	ipaneologd "github.com/ikawaha/kagome-dict-ipa-neologd"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func TestMorphologicalTokenizerTokenize(t *testing.T) {
	kagome, err := tokenizer.New(ipaneologd.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		t.Error("error: fail to initialize kagome tokenizer")
	}
	tok := NewMorphologicalTokenizer(*kagome, tokenizer.Search)
	cases := []struct {
		sentence string
		expected *TokenStream
	}{
		{
			sentence: "Ishiuchi Maruyama",
			expected: NewTokenStream([]Token{
				NewToken("Ishiuchi", "Ishiuchi"),
				NewToken("Maruyama", "Maruyama"),
			}),
		},
		{
			sentence: "石打丸山スキー場",
			expected: NewTokenStream([]Token{
				NewToken("石打丸山スキー場", "イシウチマルヤマスキージョウ"),
			}),
		},
		{
			sentence: "石打丸山",
			expected: NewTokenStream([]Token{
				NewToken("石打丸山", "イシウチマルヤマ"),
			}),
		},
		{
			sentence: "いしうちまるやま",
			expected: NewTokenStream([]Token{
				NewToken("い", "イ"),
				NewToken("し", "シ"),
				NewToken("うち", "ウチ"),
				NewToken("まるや", "マルヤ"),
				NewToken("ま", "マ"),
			}),
		},
		{
			sentence: "イシウチ",
			expected: NewTokenStream([]Token{
				NewToken("イシウチ", "イシウチ"),
			}),
		},
		{
			sentence: "白馬",
			expected: NewTokenStream([]Token{
				NewToken("白馬", "ハクバ"),
			}),
		},
		{
			sentence: "白馬47",
			expected: NewTokenStream([]Token{
				NewToken("白馬", "ハクバ"),
				NewToken("47", "47"),
			}),
		},
		{
			sentence: "琵琶湖バレイ",
			expected: NewTokenStream([]Token{
				NewToken("琵琶湖", "ビワコ"),
				NewToken("バレイ", "バレイ"),
			}),
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tok.Tokenize(tt.sentence), tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
