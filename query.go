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
	sorter   Sorter
}

func NewMatchQuery(text string, logic Logic, analyzer Analyzer, sorter Sorter) *MatchQuery {
	return &MatchQuery{
		text:     text,
		logic:    logic,
		analyzer: analyzer,
		sorter:   sorter,
	}
}

func (q *MatchQuery) Searcher(storage Storage) Searcher {
	tokenStream := q.analyzer.Analyze(q.text)
	return NewMatchSearcher(tokenStream, q.logic, storage, q.sorter)

}

type PhraseQuery struct {
	phrase   string
	analyzer Analyzer
	sorter   Sorter
}

func NewPhraseQuery(phrase string, analyzer Analyzer, sorter Sorter) *PhraseQuery {
	return &PhraseQuery{
		phrase:   phrase,
		analyzer: analyzer,
		sorter:   sorter,
	}
}

func (q *PhraseQuery) Searcher(storage Storage) Searcher {
	terms := q.analyzer.Analyze(q.phrase)
	return NewPhraseSearcher(terms, storage, q.sorter)
}
