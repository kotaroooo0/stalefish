package stalefish

import (
	"strings"

	"github.com/kljensen/snowball/english"
	"github.com/kotaroooo0/gojaconv/jaconv"
)

type CharType int

const (
	Kana   CharType = iota + 1 // カナ
	Romaji                     // ローマ字
)

func (c CharType) String() string {
	switch c {
	case Kana:
		return "Kana"
	case Romaji:
		return "Romaji"
	default:
		return "Unknown"
	}
}

type TokenFilter interface {
	Filter(*TokenStream) *TokenStream
}

type LowercaseFilter struct{}

func NewLowercaseFilter() LowercaseFilter {
	return LowercaseFilter{}
}

func (f LowercaseFilter) Filter(tokenStream *TokenStream) *TokenStream {
	r := make([]Token, tokenStream.size())
	for i, token := range tokenStream.Tokens {
		lower := strings.ToLower(token.Term)
		r[i] = NewToken(lower, setKana(token.Kana))
	}
	return NewTokenStream(r)
}

type StopWordFilter struct{}

func NewStopWordFilter() StopWordFilter {
	return StopWordFilter{}
}

func (f StopWordFilter) Filter(tokenStream *TokenStream) *TokenStream {
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

func NewStemmerFilter() StemmerFilter {
	return StemmerFilter{}
}

func (f StemmerFilter) Filter(tokenStream *TokenStream) *TokenStream {
	r := make([]Token, tokenStream.size())
	for i, token := range tokenStream.Tokens {
		stemmed := english.Stem(token.Term, false)
		r[i] = NewToken(stemmed, setKana(token.Kana))
	}
	return NewTokenStream(r)
}

type ReadingformFilter struct {
	charType CharType
}

func NewReadingformFilter(charType CharType) ReadingformFilter {
	return ReadingformFilter{
		charType: charType,
	}
}

func (f ReadingformFilter) Filter(tokenStream *TokenStream) *TokenStream {
	// ローマ字に指定されていたらローマ字に変換する、それ以外ではカナに変換する
	if f.charType == Romaji {
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
