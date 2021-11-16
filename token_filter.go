package stalefish

import (
	"strings"

	"github.com/kljensen/snowball/english"
	"github.com/kotaroooo0/gojaconv/jaconv"
)

type TokenFilter interface {
	Filter(TokenStream) TokenStream
}

type LowercaseFilter struct{}

func NewLowercaseFilter() LowercaseFilter {
	return LowercaseFilter{}
}

func (f LowercaseFilter) Filter(tokenStream TokenStream) TokenStream {
	r := make([]Token, tokenStream.Size())
	for i, token := range tokenStream.Tokens {
		lower := strings.ToLower(token.Term)
		r[i] = NewToken(lower)
	}
	return NewTokenStream(r)
}

type StopWordFilter struct {
	stopWords []string
}

func NewStopWordFilter(stopWords []string) StopWordFilter {
	return StopWordFilter{
		stopWords: stopWords,
	}
}

func (f StopWordFilter) Filter(tokenStream TokenStream) TokenStream {
	stopwords := make(map[string]struct{})
	for _, w := range f.stopWords {
		stopwords[w] = struct{}{}
	}
	r := make([]Token, 0, tokenStream.Size())
	for _, token := range tokenStream.Tokens {
		if _, ok := stopwords[token.Term]; !ok {
			r = append(r, token)
		}
	}
	return NewTokenStream(r)
}

type StemmerFilter struct{}

func NewStemmerFilter() StemmerFilter {
	return StemmerFilter{}
}

func (f StemmerFilter) Filter(tokenStream TokenStream) TokenStream {
	r := make([]Token, tokenStream.Size())
	for i, token := range tokenStream.Tokens {
		stemmed := english.Stem(token.Term, false)
		r[i] = NewToken(stemmed)
	}
	return NewTokenStream(r)
}

type RomajiReadingformFilter struct{}

func NewRomajiReadingformFilter() RomajiReadingformFilter {
	return RomajiReadingformFilter{}
}

func (f RomajiReadingformFilter) Filter(tokenStream TokenStream) TokenStream {
	for i, token := range tokenStream.Tokens {
		tokenStream.Tokens[i].Term = jaconv.ToHebon(jaconv.KatakanaToHiragana(token.Kana))
	}
	return tokenStream

}

type KanaReadingformFilter struct{}

func NewKanaReadingformFilter() KanaReadingformFilter {
	return KanaReadingformFilter{}
}

func (f KanaReadingformFilter) Filter(tokenStream TokenStream) TokenStream {
	// カナはTokenizerで既に変換されているのでTokenStreamの変数にセットする
	for i := range tokenStream.Tokens {
		tokenStream.Tokens[i].Term = tokenStream.Tokens[i].Kana
	}
	return tokenStream
}
