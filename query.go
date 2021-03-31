package stalefish

type MatchAllQuery struct{}

func NewMatchAllQuery() *MatchAllQuery {
	return &MatchAllQuery{}
}

func (q *MatchAllQuery) Searcher(storage Storage) Searcher {
	return NewMatchAllSearcher(storage)
}

type MatchQuery struct {
	Text     string
	Logic    Logic
	Analyzer Analyzer
}

func NewMatchQuery(text string, analyzer Analyzer) *MatchQuery {
	return &MatchQuery{
		Text:     text,
		Analyzer: analyzer,
	}
}

func (q *MatchQuery) Searcher(storage Storage) Searcher {
	tokenStream := q.Analyzer.Analyze(q.Text)
	return NewMatchSearcher(tokenStream, q.Logic, storage)

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
