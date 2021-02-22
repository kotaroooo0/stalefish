package stalefish

import (
	"strings"
	"unicode"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

type Tokenizer interface {
	Tokenize(string) *TokenStream
}

type StandardTokenizer struct{}

func (t StandardTokenizer) Tokenize(s string) *TokenStream {
	terms := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	tokens := make([]Token, 0)
	for _, term := range terms {
		tokens = append(tokens, NewToken(term, ""))
	}
	return NewTokenStream(tokens)
}

type MorphologicalTokenizer struct {
	Kagome tokenizer.Tokenizer
	Mode   tokenizer.TokenizeMode
}

func (t MorphologicalTokenizer) Tokenize(s string) *TokenStream {
	kagomeTokens := t.Kagome.Analyze(s, t.Mode)
	tokens := make([]Token, 0)
	for _, token := range kagomeTokens {
		features := token.Features()
		if features[1] == "空白" {
			continue
		}
		kana := token.Surface
		if len(features) >= 8 {
			kana = features[7]
		}
		tokens = append(tokens, NewToken(token.Surface, kana))
	}
	return NewTokenStream(tokens)
}

func NewMorphologicalTokenizer(kagome tokenizer.Tokenizer, mode tokenizer.TokenizeMode) *MorphologicalTokenizer {
	return &MorphologicalTokenizer{
		Kagome: kagome,
		Mode:   mode,
	}
}
