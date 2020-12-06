package stalefish

type Query interface {
	Searcher() Searcher
}

type MatchAllQuery struct{}

func (q *MatchAllQuery) Searcher(storage Storage) Searcher {
	return NewMatchAllSearcher(storage)
}

type PhraseQuery struct {
	Phrase   string
	Analyzer Analyzer
}

func NewPhraseQuery(phrase string, analyzer Analyzer) *PhraseQuery {
	return &PhraseQuery{
		Phrase:   phrase,
		Analyzer: analyzer,
	}
}

func (q *PhraseQuery) Searcher(storage Storage) Searcher {
	terms := q.Analyzer.Analyze(q.Phrase)
	return NewPhraseSearcher(terms, storage)
}
