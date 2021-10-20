package stalefish

import (
	"fmt"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/kotaroooo0/stalefish/morphology"
)

func TestMorphologicalTokenizerTokenize(t *testing.T) {
	cases := []struct {
		text     string
		expected TokenStream
	}{
		{
			text: "今日は天気が良い",
			expected: TokenStream{
				Tokens: []Token{
					{Term: "今日", Kana: "キョウ"},
					{Term: "は", Kana: "ハ"},
					{Term: "天気", Kana: "テンキ"},
					{Term: "が", Kana: "ガ"},
					{Term: "良い", Kana: "ヨイ"},
				},
			},
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

func TestNgramTokenizer_Tokenize(t *testing.T) {
	tests := []struct {
		n        int
		text     string
		expected TokenStream
	}{
		{
			n:        1,
			text:     "hogefuga",
			expected: TokenStream{Tokens: []Token{{Term: "h"}, {Term: "o"}, {Term: "g"}, {Term: "e"}, {Term: "f"}, {Term: "u"}, {Term: "g"}, {Term: "a"}}},
		},
		{
			n:        2,
			text:     "hogefuga",
			expected: TokenStream{Tokens: []Token{{Term: "ho"}, {Term: "og"}, {Term: "ge"}, {Term: "ef"}, {Term: "fu"}, {Term: "ug"}, {Term: "ga"}}},
		},
		{
			n:        3,
			text:     "hogefuga",
			expected: TokenStream{Tokens: []Token{{Term: "hog"}, {Term: "oge"}, {Term: "gef"}, {Term: "efu"}, {Term: "fug"}, {Term: "uga"}}},
		},
		{
			n:        1,
			text:     "日本昔ばなし",
			expected: TokenStream{Tokens: []Token{{Term: "日"}, {Term: "本"}, {Term: "昔"}, {Term: "ば"}, {Term: "な"}, {Term: "し"}}},
		},
		{
			n:        2,
			text:     "日本昔ばなし",
			expected: TokenStream{Tokens: []Token{{Term: "日本"}, {Term: "本昔"}, {Term: "昔ば"}, {Term: "ばな"}, {Term: "なし"}}},
		},
		{
			n:        6,
			text:     "日本昔ばなし",
			expected: TokenStream{Tokens: []Token{{Term: "日本昔ばなし"}}},
		},
		{
			n:        7,
			text:     "日本昔ばなし",
			expected: TokenStream{Tokens: []Token{}},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("n = %v, text = %v, expected = %v", tt.n, tt.text, tt.expected), func(t *testing.T) {
			tr := &NgramTokenizer{
				n: tt.n,
			}
			if got := tr.Tokenize(tt.text); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NgramTokenizer.Tokenize() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
