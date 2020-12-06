package stalefish

type Analyzer struct {
	CharFilters  []CharFilter
	Tokenizer    Tokenizer
	TokenFilters []TokenFilter
}

func NewAnalyzer(charFilters []CharFilter, tokenizer Tokenizer, tokenFilters []TokenFilter) Analyzer {
	return Analyzer{
		CharFilters:  charFilters,
		Tokenizer:    tokenizer,
		TokenFilters: tokenFilters,
	}
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
