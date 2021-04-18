package stalefish

import (
	"strings"
	"unicode"

	"github.com/kotaroooo0/stalefish/morphology"
)

type Tokenizer interface {
	Tokenize(string) *TokenStream
}

type StandardTokenizer struct{}

func NewStandardTokenizer() *StandardTokenizer {
	return &StandardTokenizer{}
}

func (t *StandardTokenizer) Tokenize(s string) *TokenStream {
	terms := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	tokens := make([]Token, len(terms))
	for i, term := range terms {
		tokens[i] = NewToken(term)
	}
	return NewTokenStream(tokens)
}

type MorphologicalTokenizer struct {
	morphology morphology.Morphology
}

func NewMorphologicalTokenizer(morphology morphology.Morphology) *MorphologicalTokenizer {
	return &MorphologicalTokenizer{
		morphology: morphology,
	}
}

func (t *MorphologicalTokenizer) Tokenize(s string) *TokenStream {
	mTokens := t.morphology.Analyze(s)
	tokens := make([]Token, len(mTokens))
	for i, t := range mTokens {
		tokens[i] = NewToken(t.Term, setKana(t.Kana))
	}
	return NewTokenStream(tokens)
}
