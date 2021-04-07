package stalefish

type MatchAllQuery struct{}

func NewMatchAllQuery() *MatchAllQuery {
	return &MatchAllQuery{}
}

func (q *MatchAllQuery) Searcher(storage Storage) Searcher {
	return NewMatchAllSearcher(storage)
}

type MatchQuery struct {
	text     string
	logic    Logic
	analyzer Analyzer
}

func NewMatchQuery(text string, logic Logic, analyzer Analyzer) *MatchQuery {
	return &MatchQuery{
		text:     text,
		logic:    logic,
		analyzer: analyzer,
	}
}

func (q *MatchQuery) Searcher(storage Storage) Searcher {
	tokenStream := q.analyzer.Analyze(q.text)
	return NewMatchSearcher(tokenStream, q.logic, storage)

}

type PhraseQuery struct {
	phrase   string
	analyzer Analyzer
}

func NewPhraseQuery(phrase string, analyzer Analyzer) *PhraseQuery {
	return &PhraseQuery{
		phrase:   phrase,
		analyzer: analyzer,
	}
}

func (q *PhraseQuery) Searcher(storage Storage) Searcher {
	terms := q.analyzer.Analyze(q.phrase)
	return NewPhraseSearcher(terms, storage)
}
