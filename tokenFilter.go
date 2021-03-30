package stalefish

import (
	"strings"

	"github.com/kljensen/snowball/english"
	"github.com/kotaroooo0/gojaconv/jaconv"
)

type TokenFilter interface {
	Filter(*TokenStream) *TokenStream
}

type LowercaseFilter struct{}

func (f LowercaseFilter) Filter(tokenStream *TokenStream) *TokenStream {
	r := make([]Token, tokenStream.size())
	for i, token := range tokenStream.Tokens {
		lower := strings.ToLower(token.Term)
		r[i] = NewToken(lower, SetKana(token.Kana))
	}
	return NewTokenStream(r, tokenStream.Selected)
}

type StopWordFilter struct{}

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
	return NewTokenStream(r, tokenStream.Selected)
}

type StemmerFilter struct{}

func (f StemmerFilter) Filter(tokenStream *TokenStream) *TokenStream {
	r := make([]Token, tokenStream.size())
	for i, token := range tokenStream.Tokens {
		stemmed := english.Stem(token.Term, false)
		r[i] = NewToken(stemmed, SetKana(token.Kana))
	}
	return NewTokenStream(r, tokenStream.Selected)
}

type ReadingformFilter struct {
	Selected Kind
}

func NewReadingformFilter(kind Kind) ReadingformFilter {
	return ReadingformFilter{
		Selected: kind,
	}
}

func (f ReadingformFilter) Filter(tokenStream *TokenStream) *TokenStream {
	// ローマ字に指定されていたらローマ字に変換する、それ以外ではカナに変換する
	if f.Selected == Romaji {
		tokenStream.Selected = Romaji
		for i, token := range tokenStream.Tokens {
			token.Romaji = jaconv.ToHebon(jaconv.KatakanaToHiragana(token.Kana))
			tokenStream.Tokens[i] = token
		}
		return tokenStream
	}

	// カナはTokenizerで既に変換されているのでTokenStreamの変数にセットする
	tokenStream.Selected = Kana
	return tokenStream
}
