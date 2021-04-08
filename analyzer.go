package stalefish

type Analyzer struct {
	charFilters  []CharFilter
	tokenizer    Tokenizer
	tokenFilters []TokenFilter
}

func NewAnalyzer(charFilters []CharFilter, tokenizer Tokenizer, tokenFilters []TokenFilter) Analyzer {
	return Analyzer{
		charFilters:  charFilters,
		tokenizer:    tokenizer,
		tokenFilters: tokenFilters,
	}
}

func (a Analyzer) Analyze(s string) *TokenStream {
	for _, c := range a.charFilters {
		s = c.Filter(s)
	}
	tokenStream := a.tokenizer.Tokenize(s)
	for _, f := range a.tokenFilters {
		tokenStream = f.Filter(tokenStream)
	}
	return tokenStream
}
