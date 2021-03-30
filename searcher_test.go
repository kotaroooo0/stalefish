package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMatchAllSearch(t *testing.T) {
	// TODO: デバッグ用にplayground的に使う、ちゃんとしたテストも書きたい
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	db.Exec("truncate table documents")
	db.Exec("truncate table tokens")
	db.Exec("truncate table inverted_indexes")

	storage := NewStorageRdbImpl(db)
	analyzer := NewAnalyzer([]CharFilter{}, StandardTokenizer{}, []TokenFilter{LowercaseFilter{}, StopWordFilter{}})
	indexer := NewIndexer(storage, analyzer, make(InvertedIndexMap))

	doc1 := NewDocument("aa bb cc dd aa bb")
	err = indexer.AddDocument(doc1)
	if err != nil {
		t.Error(err)
	}
	doc2 := NewDocument("ee ff gg hh ii jj kk")
	err = indexer.AddDocument(doc2)
	if err != nil {
		t.Error(err)
	}
	doc3 := NewDocument("aa aa bb bb jj kk ll oo nn bb vv rr tt uu yy qq")
	err = indexer.AddDocument(doc3)
	if err != nil {
		t.Error(err)
	}

	matchAllSearcher := NewMatchAllSearcher(storage)
	docs, err := matchAllSearcher.Search()
	if err != nil {
		t.Error(err)
	}
	doc1.ID = 1
	doc2.ID = 2
	doc3.ID = 3
	if diff := cmp.Diff(docs, []Document{doc1, doc2, doc3}); diff != "" {
		t.Errorf("Diff: (-got +want)\n%s", diff)
	}
}

func TestPhraseSearch(t *testing.T) {
	// TODO: デバッグ用にplayground的に使う、ちゃんとしたテストも書きたい
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	db.Exec("truncate table documents")
	db.Exec("truncate table tokens")
	db.Exec("truncate table inverted_indexes")

	storage := NewStorageRdbImpl(db)
	analyzer := NewAnalyzer([]CharFilter{}, StandardTokenizer{}, []TokenFilter{LowercaseFilter{}, StopWordFilter{}})
	indexer := NewIndexer(storage, analyzer, make(InvertedIndexMap))

	doc1 := NewDocument("aa bb cc dd aa bb")
	err = indexer.AddDocument(doc1)
	if err != nil {
		t.Error(err)
	}
	doc2 := NewDocument("ee ff gg hh ii jj kk")
	err = indexer.AddDocument(doc2)
	if err != nil {
		t.Error(err)
	}
	doc3 := NewDocument("aa aa bb bb jj kk ll oo nn bb vv rr tt uu yy qq")
	err = indexer.AddDocument(doc3)
	if err != nil {
		t.Error(err)
	}
	doc1.ID = 1
	doc2.ID = 2
	doc3.ID = 3

	cases := []struct {
		terms        *TokenStream
		expectedDocs []Document
	}{
		{
			terms: NewTokenStream(
				[]Token{
					NewToken("aa"),
					NewToken("bb"),
				},
				Term,
			),
			expectedDocs: []Document{doc1, doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{
					NewToken("tt"),
					NewToken("uu"),
				},
				Term,
			),
			expectedDocs: []Document{doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{
					NewToken("aa"),
				},
				Term,
			), expectedDocs: []Document{doc1, doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{
					NewToken("ff"),
				},
				Term,
			), expectedDocs: []Document{doc2},
		},
		{
			terms:        NewTokenStream([]Token{}, Term),
			expectedDocs: []Document{},
		},
	}

	for _, tt := range cases {
		phraseSearcher := NewPhraseSearcher(tt.terms, storage)
		actualDocs, err := phraseSearcher.Search()
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(tt.expectedDocs, actualDocs); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
