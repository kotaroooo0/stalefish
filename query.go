package stalefish

type Query interface {
	Searcher() (Searcher, error)
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

func (q *PhraseQuery) Searcher() (Searcher, error) {
	tokens := q.Analyzer.Analyze(q.Phrase)
	s, err := NewPhraseSearcher(tokens)
	if err != nil {
		return nil, err
	}
	return s, nil
}
