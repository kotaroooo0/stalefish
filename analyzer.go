package stalefish

type Analyzer struct {
	CharFilters  []CharFilter
	Tokenizer    Tokenizer
	TokenFilters []TokenFilter
}

func (a Analyzer) Analyze(s string) []string {
	for _, c := range a.CharFilters {
		s = c.Filter(s)
	}
	tokens := a.Tokenizer.Tokenize(s)
	for _, f := range a.TokenFilters {
		tokens = f.Filter(tokens)
	}
	return tokens
}
