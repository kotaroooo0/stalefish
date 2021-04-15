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
	truncateTableAll(db)

	storage := NewStorageRdbImpl(db)
	analyzer := NewAnalyzer([]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewLowercaseFilter(), NewStopWordFilter()})
	indexer := NewIndexer(storage, analyzer)

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

func TestMatchSearch(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)

	storage := NewStorageRdbImpl(db)
	analyzer := NewAnalyzer([]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewLowercaseFilter(), NewStopWordFilter()})
	indexer := NewIndexer(storage, analyzer)

	doc1 := NewDocument("aa bb tt")
	if err = indexer.AddDocument(doc1); err != nil {
		t.Error(err)
	}
	doc2 := NewDocument("ee ff")
	if err = indexer.AddDocument(doc2); err != nil {
		t.Error(err)
	}
	doc3 := NewDocument("aa bb gg")
	if err = indexer.AddDocument(doc3); err != nil {
		t.Error(err)
	}
	doc4 := NewDocument("cc dd")
	if err = indexer.AddDocument(doc4); err != nil {
		t.Error(err)
	}
	doc1.ID = 1
	doc2.ID = 2
	doc3.ID = 3
	doc4.ID = 4

	cases := []struct {
		terms        *TokenStream
		logic        Logic
		expectedDocs []Document
	}{
		{
			terms: NewTokenStream(
				[]Token{NewToken("aa"), NewToken("bb")},
			),
			logic:        AND,
			expectedDocs: []Document{doc1, doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("ee"), NewToken("cc")},
			),
			logic:        OR,
			expectedDocs: []Document{doc2, doc4},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("aa"), NewToken("tt"), NewToken("dd")},
			),
			logic:        OR,
			expectedDocs: []Document{doc1, doc3, doc4},
		},
	}

	for _, tt := range cases {
		matchSearcher := NewMatchSearcher(tt.terms, tt.logic, storage)
		actualDocs, err := matchSearcher.Search()
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(tt.expectedDocs, actualDocs); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestPhraseSearch(t *testing.T) {
	// TODO: デバッグ用にplayground的に使う、ちゃんとしたテストも書きたい
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)

	storage := NewStorageRdbImpl(db)
	analyzer := NewAnalyzer([]CharFilter{}, NewStandardTokenizer(), []TokenFilter{NewLowercaseFilter(), NewStopWordFilter()})
	indexer := NewIndexer(storage, analyzer)

	doc1 := NewDocument("aa bb cc")
	if err = indexer.AddDocument(doc1); err != nil {
		t.Error(err)
	}
	doc2 := NewDocument("ee ff gg")
	if err = indexer.AddDocument(doc2); err != nil {
		t.Error(err)
	}
	doc3 := NewDocument("jj kk ll aa bb")
	if err = indexer.AddDocument(doc3); err != nil {
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
				[]Token{NewToken("aa"), NewToken("bb")},
			),
			expectedDocs: []Document{doc1, doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("ff"), NewToken("gg")},
			),
			expectedDocs: []Document{doc2},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("kk"), NewToken("ll"), NewToken("aa")},
			),
			expectedDocs: []Document{doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("aa")},
			), expectedDocs: []Document{doc1, doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("ll")},
			), expectedDocs: []Document{doc3},
		},
		{
			terms:        NewTokenStream([]Token{}),
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
