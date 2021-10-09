package morphology

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		text     string
		expected []MorphologyToken
	}{
		{
			text: "今日は天気が良い",
			expected: []MorphologyToken{
				NewMorphologyToken("今日", "キョウ"),
				NewMorphologyToken("は", "ハ"),
				NewMorphologyToken("天気", "テンキ"),
				NewMorphologyToken("が", "ガ"),
				NewMorphologyToken("良い", "ヨイ"),
			},
		},
		{
			text: "白馬へ滑りにいきたい",
			expected: []MorphologyToken{
				NewMorphologyToken("白馬", "ハクバ"),
				NewMorphologyToken("へ", "ヘ"),
				NewMorphologyToken("滑り", "スベリ"),
				NewMorphologyToken("に", "ニ"),
				NewMorphologyToken("いき", "イキ"),
				NewMorphologyToken("たい", "タイ"),
			},
		},
		{
			text: "Ishiuchi Maruyama",
			expected: []MorphologyToken{
				NewMorphologyToken("Ishiuchi", "Ishiuchi"),
				NewMorphologyToken("Maruyama", "Maruyama"),
			},
		},
		{
			text: "石打丸山スキー場",
			expected: []MorphologyToken{
				NewMorphologyToken("石打丸山スキー場", "イシウチマルヤマスキージョウ"),
			},
		},
		{
			text: "石打丸山",
			expected: []MorphologyToken{
				NewMorphologyToken("石打丸山", "イシウチマルヤマ"),
			},
		},
		{
			text: "いしうちまるやま",
			expected: []MorphologyToken{
				NewMorphologyToken("い", "イ"),
				NewMorphologyToken("し", "シ"),
				NewMorphologyToken("うち", "ウチ"),
				NewMorphologyToken("まるや", "マルヤ"),
				NewMorphologyToken("ま", "マ"),
			},
		},
		{
			text: "イシウチ",
			expected: []MorphologyToken{
				NewMorphologyToken("イシウチ", "イシウチ"),
			},
		},
		{
			text: "白馬",
			expected: []MorphologyToken{
				NewMorphologyToken("白馬", "ハクバ"),
			},
		},
		{
			text: "白馬47",
			expected: []MorphologyToken{
				NewMorphologyToken("白馬", "ハクバ"),
				NewMorphologyToken("47", "47"),
			},
		},
		{
			text: "琵琶湖バレイ",
			expected: []MorphologyToken{
				NewMorphologyToken("琵琶湖", "ビワコ"),
				NewMorphologyToken("バレイ", "バレイ"),
			},
		},
	}

	kagome, err := NewKagome()
	if err != nil {
		t.Error("error: fail to initialize kagome tokenizer")
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("text = %v, expected = %v", tt.text, tt.expected), func(t *testing.T) {
			if diff := cmp.Diff(tt.expected, kagome.Analyze(tt.text)); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}
