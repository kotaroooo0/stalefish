package main

type Analyzer struct {
	Tokenizer   Tokenizer
	CharFilters []CharFilter
	Filters     []Filter
}

func (a Analyzer) Analyze(s string) []string {
	for _, c := range a.CharFilters {
		s = c.Filter(s)
	}
	tokens := a.Tokenizer.Tokenize(s)
	for _, f := range a.Filters {
		tokens = f.Filter(tokens)
	}
	return tokens
}
