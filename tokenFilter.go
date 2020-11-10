package stalefish

import (
	"strings"

	"github.com/kljensen/snowball/english"
)

type TokenFilter interface {
	Filter([]string) []string
}

type LowercaseFilter struct{}

func (f LowercaseFilter) Filter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

type StopWordFilter struct{}

func (f StopWordFilter) Filter(tokens []string) []string {
	var stopwords = map[string]struct{}{
		"a": {}, "and": {}, "be": {}, "have": {}, "i": {},
		"in": {}, "of": {}, "that": {}, "the": {}, "to": {},
	}
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			r = append(r, token)
		}
	}
	return r
}

type StemmerFilter struct{}

func (f StemmerFilter) Filter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = english.Stem(token, false)
	}
	return r
}
