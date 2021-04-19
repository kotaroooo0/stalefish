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
	actual, err := storage.GetInvertedIndexByTokenIDs([]TokenID{token.ID})
	if err != nil {
		t.Error(err)
	}
	expected := NewInvertedIndex(
		map[TokenID]PostingList{
			token.ID: NewPostingList(
				NewPostings(1, []uint64{0, 4}, 2, NewPostings(3, []uint64{0, 1}, 2, nil)), 2, 4),
		},
	)
	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("Diff: (-got +want)\n%s", diff)
	}
}

func TestUpdateMemoryInvertedIndexByDocument(t *testing.T) {
	cases := []struct {
		doc      Document
		expected InvertedIndex
	}{
		{
			doc: Document{ID: 1, Body: "ho fug piyo fug fug"},
			expected: InvertedIndex{
				2: PostingList{
					Postings:       NewPostings(1, []uint64{0}, 1, nil),
					DocsCount:      1,
					PositionsCount: 1,
				},
				3: PostingList{
					Postings:       NewPostings(1, []uint64{1, 3, 4}, 3, nil),
					DocsCount:      1,
					PositionsCount: 3,
				},
				4: PostingList{
					Postings:       NewPostings(1, []uint64{2}, 1, nil),
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
		if diff := cmp.Diff(indexer.InvertedIndex, tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestUpdateMemoryInvertedIndexByToken(t *testing.T) {
	cases := []struct {
		docID    DocumentID
		token    Token
		pos      uint64
		expected InvertedIndex
	}{
		{
			// 対応するポスティングリストがない
			docID: 1,
			token: NewToken("ab"),
			pos:   1,
			expected: InvertedIndex{
				TokenID(2): PostingList{
					Postings:       &Postings{DocumentID: 1, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(3): PostingList{
					Postings:       &Postings{DocumentID: 1, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(4): PostingList{
					Postings:       &Postings{DocumentID: 2, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
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
				TokenID(3): PostingList{
					Postings:       &Postings{DocumentID: 1, Positions: []uint64{1, 99}, PositionsCount: 2, Next: nil},
					DocsCount:      1,
					PositionsCount: 2,
				},
				TokenID(4): PostingList{
					Postings:       &Postings{DocumentID: 2, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
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
				TokenID(3): PostingList{
					Postings:       &Postings{DocumentID: 1, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(4): PostingList{
					Postings:       &Postings{DocumentID: 1, Positions: []uint64{99}, PositionsCount: 1, Next: &Postings{DocumentID: 2, Positions: []uint64{1}, PositionsCount: 1, Next: nil}},
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
				TokenID(3): PostingList{
					Postings:       &Postings{DocumentID: 1, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
				TokenID(4): PostingList{
					Postings:       &Postings{DocumentID: 2, Positions: []uint64{1}, PositionsCount: 1, Next: nil},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		}
		if err := indexer.updateMemoryPostingListByToken(tt.docID, tt.token, tt.pos); err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(indexer.InvertedIndex, tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
