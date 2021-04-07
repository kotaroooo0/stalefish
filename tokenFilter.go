package stalefish

import (
	"strings"

	"github.com/kljensen/snowball/english"
	"github.com/kotaroooo0/gojaconv/jaconv"
)

type TokenFilter interface {
	filter(*TokenStream) *TokenStream
}

type LowercaseFilter struct{}

func NewLowercaseFilter() *LowercaseFilter {
	return &LowercaseFilter{}
}

func (f LowercaseFilter) filter(tokenStream *TokenStream) *TokenStream {
	r := make([]Token, tokenStream.size())
	for i, token := range tokenStream.Tokens {
		lower := strings.ToLower(token.Term)
		r[i] = NewToken(lower, setKana(token.Kana))
	}
	return NewTokenStream(r)
}

type StopWordFilter struct{}

func NewStopWordFilter() *StopWordFilter {
	return &StopWordFilter{}
}

func (f StopWordFilter) filter(tokenStream *TokenStream) *TokenStream {
	var stopwords = map[string]struct{}{
		"a": {}, "and": {}, "be": {}, "have": {}, "i": {},
		"in": {}, "of": {}, "that": {}, "the": {}, "to": {},
	}
	r := make([]Token, 0, tokenStream.size())
	for _, token := range tokenStream.Tokens {
		if _, ok := stopwords[token.Term]; !ok {
			r = append(r, token)
		}
	}
	return NewTokenStream(r)
}

type StemmerFilter struct{}

func NewStemmerFilter() *StemmerFilter {
	return &StemmerFilter{}
}

func (f StemmerFilter) filter(tokenStream *TokenStream) *TokenStream {
	r := make([]Token, tokenStream.size())
	for i, token := range tokenStream.Tokens {
		stemmed := english.Stem(token.Term, false)
		r[i] = NewToken(stemmed, setKana(token.Kana))
	}
	return NewTokenStream(r)
}

type ReadingformFilter struct {
	selected Kind
}

func NewReadingformFilter(kind Kind) ReadingformFilter {
	return ReadingformFilter{
		selected: kind,
	}
}

func (f ReadingformFilter) filter(tokenStream *TokenStream) *TokenStream {
	// ローマ字に指定されていたらローマ字に変換する、それ以外ではカナに変換する
	if f.selected == Romaji {
		for i, token := range tokenStream.Tokens {
			tokenStream.Tokens[i].Term = jaconv.ToHebon(jaconv.KatakanaToHiragana(token.Kana))
		}
		return tokenStream
	}

	// カナはTokenizerで既に変換されているのでTokenStreamの変数にセットする
	for i := range tokenStream.Tokens {
		tokenStream.Tokens[i].Term = tokenStream.Tokens[i].Kana
	}
	return tokenStream
}
