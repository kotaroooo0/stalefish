package stalefish

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/kotaroooo0/stalefish/morphology"
)

func TestMorphologicalTokenizerTokenize(t *testing.T) {
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
		t.Run(fmt.Sprintf("text = %v, expected = %v", tt.text, tt.expected), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockMorphology := NewMockMorphology(mockCtrl)

			// Given
			tokenizer := NewMorphologicalTokenizer(mockMorphology)
			mockMorphology.EXPECT().Analyze(tt.text).Return([]morphology.MorphologyToken{
				morphology.NewMorphologyToken("今日", "キョウ"),
				morphology.NewMorphologyToken("は", "ハ"),
				morphology.NewMorphologyToken("天気", "テンキ"),
				morphology.NewMorphologyToken("が", "ガ"),
				morphology.NewMorphologyToken("良い", "ヨイ"),
			})

			// When
			actual := tokenizer.Tokenize(tt.text)

			// Then
			if diff := cmp.Diff(actual, tt.expected); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}
