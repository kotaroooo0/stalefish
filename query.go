package stalefish

type MatchAllQuery struct{}

func NewMatchAllQuery() MatchAllQuery {
	return MatchAllQuery{}
}

func (q MatchAllQuery) Searcher(storage Storage) Searcher {
	return NewMatchAllSearcher(storage)
}

type MatchQuery struct {
	keyword  string
	logic    Logic
	analyzer Analyzer
	sorter   Sorter
}

func NewMatchQuery(keyword string, logic Logic, analyzer Analyzer, sorter Sorter) MatchQuery {
	return MatchQuery{
		keyword:  keyword,
		logic:    logic,
		analyzer: analyzer,
		sorter:   sorter,
	}
}

func (q MatchQuery) Searcher(storage Storage) Searcher {
	tokenStream := q.analyzer.Analyze(q.keyword)
	return NewMatchSearcher(tokenStream, q.logic, storage, q.sorter)

}

type PhraseQuery struct {
	phrase   string
	analyzer Analyzer
	sorter   Sorter
}

func NewPhraseQuery(phrase string, analyzer Analyzer, sorter Sorter) PhraseQuery {
	return PhraseQuery{
		phrase:   phrase,
		analyzer: analyzer,
		sorter:   sorter,
	}
}

func (q PhraseQuery) Searcher(storage Storage) Searcher {
	terms := q.analyzer.Analyze(q.phrase)
	return NewPhraseSearcher(terms, storage, q.sorter)
}
