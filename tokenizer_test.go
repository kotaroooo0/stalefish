package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kotaroooo0/stalefish/morphology"
)

type kagomeMock struct {
}

func NewKagomeMock() *kagomeMock {
	return &kagomeMock{}
}

func (k *kagomeMock) Analyze(text string) []morphology.MorphologyToken {
	return []morphology.MorphologyToken{
		morphology.NewMorphologyToken("今日", "キョウ"),
		morphology.NewMorphologyToken("は", "ハ"),
		morphology.NewMorphologyToken("天気", "テンキ"),
		morphology.NewMorphologyToken("が", "ガ"),
		morphology.NewMorphologyToken("良い", "ヨイ"),
	}
}

func TestMorphologicalTokenizerTokenize(t *testing.T) {
	t.Parallel()
	tokenizer := NewMorphologicalTokenizer(NewKagomeMock())
	cases := []struct {
		text     string
		expected *TokenStream
	}{
		{
			text: "今日は天気が良い",
			expected: NewTokenStream([]Token{
				NewToken("今日", setKana("キョウ")),
				NewToken("は", setKana("ハ")),
				NewToken("天気", setKana("テンキ")),
				NewToken("が", setKana("ガ")),
				NewToken("良い", setKana("ヨイ")),
			}),
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tokenizer.Tokenize(tt.text), tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
