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

func (a Analyzer) Analyze(s string) *TokenStream {
	for _, c := range a.CharFilters {
		s = c.Filter(s)
	}
	tokenStream := a.Tokenizer.Tokenize(s)
	for _, f := range a.TokenFilters {
		tokenStream = f.Filter(tokenStream)
	}
	return tokenStream
}
