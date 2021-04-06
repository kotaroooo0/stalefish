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
		s = c.filter(s)
	}
	tokenStream := a.tokenizer.tokenize(s)
	for _, f := range a.tokenFilters {
		tokenStream = f.filter(tokenStream)
	}
	return tokenStream
}
