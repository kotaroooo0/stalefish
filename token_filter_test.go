package stalefish

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLowercaseFilter_Filter(t *testing.T) {
	tests := []struct {
		tokenStream TokenStream
		want        TokenStream
	}{
		{
			tokenStream: TokenStream{Tokens: []Token{{Term: "Hoge"}, {Term: "fuGA"}, {Term: "PIYO"}}},
			want:        TokenStream{Tokens: []Token{{Term: "hoge"}, {Term: "fuga"}, {Term: "piyo"}}},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tokenStream = %v, want = %v", tt.tokenStream, tt.want), func(t *testing.T) {
			f := LowercaseFilter{}
			if got := f.Filter(tt.tokenStream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LowercaseFilter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopWordFilter_Filter(t *testing.T) {
	tests := []struct {
		stopWords   []string
		tokenStream TokenStream
		want        TokenStream
	}{
		{
			stopWords:   []string{"hoge"},
			tokenStream: TokenStream{Tokens: []Token{{Term: "hoge"}, {Term: "fuga"}, {Term: "piyo"}}},
			want:        TokenStream{Tokens: []Token{{Term: "fuga"}, {Term: "piyo"}}},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("stopWords = %v, tokenStream = %v, want = %v", tt.stopWords, tt.tokenStream, tt.want), func(t *testing.T) {
			f := StopWordFilter{
				stopWords: tt.stopWords,
			}
			if got := f.Filter(tt.tokenStream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StopWordFilter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStemmerFilter_Filter(t *testing.T) {
	tests := []struct {
		tokenStream TokenStream
		want        TokenStream
	}{
		{
			tokenStream: TokenStream{Tokens: []Token{{Term: "pens"}, {Term: "came"}}},
			want:        TokenStream{Tokens: []Token{{Term: "pen"}, {Term: "came"}}},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tokenStream = %v, want = %v", tt.tokenStream, tt.want), func(t *testing.T) {
			f := StemmerFilter{}
			if got := f.Filter(tt.tokenStream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StemmerFilter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRomajiReadingformFilter_Filter(t *testing.T) {
	tests := []struct {
		tokenStream TokenStream
		want        TokenStream
	}{
		{
			tokenStream: TokenStream{Tokens: []Token{{Term: "おっ早う！", Kana: "おはよう"}, {Term: "チョット！", Kana: "ちょっと"}}},
			want:        TokenStream{Tokens: []Token{{Term: "ohayo", Kana: "おはよう"}, {Term: "chotto", Kana: "ちょっと"}}},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tokenStream = %v, want = %v", tt.tokenStream, tt.want), func(t *testing.T) {
			f := RomajiReadingformFilter{}
			if got := f.Filter(tt.tokenStream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RomajiReadingformFilter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKanaReadingformFilter_Filter(t *testing.T) {
	tests := []struct {
		tokenStream TokenStream
		want        TokenStream
	}{
		{
			tokenStream: TokenStream{Tokens: []Token{{Term: "おっ早う！", Kana: "おはよう"}, {Term: "チョット！", Kana: "ちょっと"}}},
			want:        TokenStream{Tokens: []Token{{Term: "おはよう", Kana: "おはよう"}, {Term: "ちょっと", Kana: "ちょっと"}}},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tokenStream = %v, want = %v", tt.tokenStream, tt.want), func(t *testing.T) {

			f := KanaReadingformFilter{}
			if got := f.Filter(tt.tokenStream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KanaReadingformFilter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
