package stalefish

import (
	"strings"
	"unicode"

	"github.com/kotaroooo0/stalefish/morphology"
)

type Tokenizer interface {
	tokenize(string) *TokenStream
}

type StandardTokenizer struct{}

func NewStandardTokenizer() StandardTokenizer {
	return StandardTokenizer{}
}

func (t StandardTokenizer) tokenize(s string) *TokenStream {
	terms := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	tokens := make([]Token, 0)
	for _, term := range terms {
		tokens = append(tokens, NewToken(term))
	}
	return NewTokenStream(tokens, Term)
}

type MorphologicalTokenizer struct {
	morphology morphology.Morphology
}

func (t MorphologicalTokenizer) tokenize(s string) *TokenStream {
	mTokens := t.morphology.Analyze(s)
	tokens := make([]Token, len(mTokens))
	for i, t := range mTokens {
		tokens[i] = NewToken(t.Term, SetKana(t.Kana))
	}
	return NewTokenStream(tokens, Term)
}

func NewMorphologicalTokenizer(morphology morphology.Morphology) MorphologicalTokenizer {
	return MorphologicalTokenizer{
		morphology: morphology,
	}
}
