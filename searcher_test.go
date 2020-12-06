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
		terms        []string
		expectedDocs []Document
	}{
		{
			terms:        []string{"aa", "bb"},
			expectedDocs: []Document{doc1, doc3},
		},
		{
			terms:        []string{"tt", "uu"},
			expectedDocs: []Document{doc3},
		},
		{
			terms:        []string{"aa"},
			expectedDocs: []Document{doc1, doc3},
		},
		{
			terms:        []string{"ff"},
			expectedDocs: []Document{doc2},
		},
		{
			terms:        []string{},
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
