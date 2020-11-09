package stalefish

import (
	"strings"
	"unicode"
)

type Tokenizer interface {
	Tokenize(string) []string
}

type StandardTokenizer struct{}

func (t StandardTokenizer) Tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}
