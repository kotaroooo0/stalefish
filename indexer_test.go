package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// モック
type TestStorage struct {
	Storage
}

func (s TestStorage) AddToken(token Token) (TokenID, error) {
	return TokenID(0), nil
}

func (s TestStorage) GetTokenByTerm(term string) (Token, error) {
	return Token{
		ID:   TokenID(len(term)),
		Term: term,
	}, nil
}

func TestIndexerAddDocument(t *testing.T) {
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
		t.Errorf("%+v\n", err)
	}

	doc2 := NewDocument("ee ff gg hh ii jj kk")
	err = indexer.AddDocument(doc2)
	if err != nil {
		t.Errorf("%+v\n", err)
	}

	doc3 := NewDocument("aa aa bb bb jj kk ll oo nn bb vv rr tt uu yy qq")
	err = indexer.AddDocument(doc3)
	if err != nil {
		t.Errorf("%+v\n", err)
	}
	token, err := storage.GetTokenByTerm("aa")
	if err != nil {
		t.Error(err)
	}
	actual, err := storage.GetInvertedIndexByTokenID(token.ID)
	if err != nil {
		t.Error(err)
	}
	expected := InvertedIndexValue{
		Token:          Token{ID: 1, Term: "aa"},
		PostingList:    NewPostings(1, []int{0, 4}, 2, NewPostings(3, []int{0, 1}, 2, nil)),
		DocsCount:      2,
		PositionsCount: 4,
	}
	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("Diff: (-got +want)\n%s", diff)
	}
}

func TestUpdateMemoryInvertedIndexByDocument(t *testing.T) {
	cases := []struct {
		doc    Document
		output InvertedIndex
	}{
		{
			doc: Document{ID: 1, Body: "int string uint string string"},
			output: InvertedIndex{
				3: InvertedIndexValue{
					Token:          Token{ID: 3, Term: "int"},
					PostingList:    NewPostings(1, []int{0}, 1, nil),
					DocsCount:      1,
					PositionsCount: 1,
				},
				6: InvertedIndexValue{
					Token:          Token{ID: 6, Term: "string"},
					PostingList:    NewPostings(1, []int{1, 3, 4}, 3, nil),
					DocsCount:      1,
					PositionsCount: 3,
				},
				4: InvertedIndexValue{
					Token:          Token{ID: 4, Term: "uint"},
					PostingList:    NewPostings(1, []int{2}, 1, nil),
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		},
	}

	for _, tt := range cases {
		indexer := Indexer{
			Storage:       TestStorage{},
			Analyzer:      Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
			InvertedIndex: InvertedIndex{},
		}
		if err := indexer.updateMemoryInvertedIndexByDocument(tt.doc); err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(indexer.InvertedIndex, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestUpdateMemoryInvertedIndexByToken(t *testing.T) {
	cases := []struct {
		docID    DocumentID
		token    Token
		pos      int
		expected InvertedIndex
	}{
		{
			// 対応するinvertedIndexValueがない
			docID: 1,
			token: NewToken("ab"),
			pos:   1,
			expected: InvertedIndex{
				TokenID(2): InvertedIndexValue{
					Token:          Token{ID: 2, Term: "ab", Kana: ""},
					PostingList:    &Postings{DocumentID: 1, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(3): InvertedIndexValue{
					Token:          Token{ID: 3, Term: "abc", Kana: ""},
					PostingList:    &Postings{DocumentID: 1, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(4): InvertedIndexValue{
					Token:          Token{ID: 4, Term: "abcd", Kana: ""},
					PostingList:    &Postings{DocumentID: 2, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		},
		{
			// 既に対象ドキュメントのポスティングが存在する
			docID: 1,
			token: NewToken("abc"),
			pos:   99,
			expected: InvertedIndex{
				TokenID(3): InvertedIndexValue{
					Token:          Token{ID: 3, Term: "abc", Kana: ""},
					PostingList:    &Postings{DocumentID: 1, Positions: []int{1, 99}, PositionsCount: 2, Next: nil},
					DocsCount:      1,
					PositionsCount: 2,
				},
				TokenID(4): InvertedIndexValue{
					Token:          Token{ID: 4, Term: "abcd", Kana: ""},
					PostingList:    &Postings{DocumentID: 2, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		},
		{
			// まだ対象ドキュメントのポスティングが存在しない
			docID: 1,
			token: NewToken("abcd"),
			pos:   99,
			expected: InvertedIndex{
				TokenID(3): InvertedIndexValue{
					Token:          Token{ID: 3, Term: "abc", Kana: ""},
					PostingList:    &Postings{DocumentID: 1, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(4): InvertedIndexValue{
					Token:          Token{ID: 4, Term: "abcd", Kana: ""},
					PostingList:    &Postings{DocumentID: 1, Positions: []int{99}, PositionsCount: 1, Next: &Postings{DocumentID: 2, Positions: []int{1}, PositionsCount: 1, Next: nil}},
					DocsCount:      2,
					PositionsCount: 2,
				},
			},
		},
	}

	for _, tt := range cases {
		indexer := Indexer{
			Storage:  TestStorage{},
			Analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
			InvertedIndex: InvertedIndex{
				TokenID(3): InvertedIndexValue{
					Token:          Token{ID: 3, Term: "abc", Kana: ""},
					PostingList:    &Postings{DocumentID: 1, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(4): InvertedIndexValue{
					Token:          Token{ID: 4, Term: "abcd", Kana: ""},
					PostingList:    &Postings{DocumentID: 2, Positions: []int{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		}
		if err := indexer.updateMemoryInvertedIndexByToken(tt.docID, tt.token, tt.pos); err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(indexer.InvertedIndex, tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestMerge(t *testing.T) {
	cases := []struct {
		memoryInvertedIndex  InvertedIndexValue
		storageInvertedIndex InvertedIndexValue
		output               InvertedIndexValue
	}{
		{
			memoryInvertedIndex: InvertedIndexValue{
				Token:          Token{ID: 3, Term: "int"},
				PostingList:    NewPostings(1, []int{0}, 1, NewPostings(3, []int{0}, 1, NewPostings(4, []int{3}, 1, nil))),
				DocsCount:      3,
				PositionsCount: 3,
			},
			storageInvertedIndex: InvertedIndexValue{
				Token:          Token{ID: 3, Term: "int"},
				PostingList:    NewPostings(2, []int{1, 2}, 2, NewPostings(4, []int{3}, 1, NewPostings(5, []int{12}, 1, nil))),
				DocsCount:      3,
				PositionsCount: 4,
			},
			output: InvertedIndexValue{
				Token:          Token{ID: 3, Term: "int"},
				PostingList:    NewPostings(1, []int{0}, 1, NewPostings(2, []int{1, 2}, 2, NewPostings(3, []int{0}, 1, NewPostings(4, []int{3}, 1, NewPostings(5, []int{12}, 1, nil))))),
				DocsCount:      5,
				PositionsCount: 6,
			},
		},
		{
			memoryInvertedIndex: InvertedIndexValue{
				Token:          Token{ID: 3, Term: "int"},
				PostingList:    NewPostings(3, []int{0}, 1, NewPostings(4, []int{0}, 1, NewPostings(5, []int{3}, 1, nil))),
				DocsCount:      3,
				PositionsCount: 3,
			},
			storageInvertedIndex: InvertedIndexValue{
				Token:          Token{ID: 3, Term: "int"},
				PostingList:    NewPostings(1, []int{1, 2}, 2, NewPostings(2, []int{3}, 1, nil)),
				DocsCount:      2,
				PositionsCount: 3,
			},
			output: InvertedIndexValue{
				Token:          Token{ID: 3, Term: "int"},
				PostingList:    NewPostings(1, []int{1, 2}, 2, NewPostings(2, []int{3}, 1, NewPostings(3, []int{0}, 1, NewPostings(4, []int{0}, 1, NewPostings(5, []int{3}, 1, nil))))),
				DocsCount:      5,
				PositionsCount: 6,
			},
		},
		{
			memoryInvertedIndex: InvertedIndexValue{
				Token:          Token{ID: 1, Term: "int"},
				PostingList:    NewPostings(1, []int{0, 4}, 2, nil),
				DocsCount:      1,
				PositionsCount: 2,
			},
			storageInvertedIndex: InvertedIndexValue{
				Token:          Token{ID: 1, Term: "int"},
				PostingList:    NewPostings(3, []int{0, 1}, 2, nil),
				DocsCount:      1,
				PositionsCount: 2,
			},
			output: InvertedIndexValue{
				Token:          Token{ID: 1, Term: "int"},
				PostingList:    NewPostings(1, []int{0, 4}, 2, NewPostings(3, []int{0, 1}, 2, nil)),
				DocsCount:      2,
				PositionsCount: 4,
			},
		},
	}
	for _, tt := range cases {
		merged, err := merge(tt.memoryInvertedIndex, tt.storageInvertedIndex)
		if err != nil {
			t.Error("error: merge failed")
		}
		if diff := cmp.Diff(merged, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
