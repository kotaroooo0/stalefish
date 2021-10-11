package stalefish

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestIndexer_AddDocument(t *testing.T) {
	tests := []struct {
		doc      Document
		expected InvertedIndex
	}{
		{
			doc: Document{ID: 2, Body: "aa bb cc aa"},
			expected: InvertedIndex(
				map[TokenID]PostingList{
					TokenID(0): {NewPostings(DocumentID(1), []uint64{0}, NewPostings(DocumentID(2), []uint64{0, 3}, NewPostings(DocumentID(3), []uint64{1}, nil)))},
					TokenID(1): {NewPostings(DocumentID(1), []uint64{1}, NewPostings(DocumentID(2), []uint64{1}, NewPostings(DocumentID(3), []uint64{2}, nil)))},
					TokenID(2): {NewPostings(DocumentID(1), []uint64{2}, NewPostings(DocumentID(2), []uint64{2}, nil))},
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("doc = %v, expected = %v", tt.doc, tt.expected), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockStorage := NewMockStorage(mockCtrl)

			// Given
			i := &Indexer{
				Storage:       mockStorage,
				Analyzer:      NewAnalyzer([]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}),
				InvertedIndex: make(InvertedIndex),
			}
			invertedIndex := InvertedIndex(
				map[TokenID]PostingList{
					TokenID(0): {NewPostings(DocumentID(1), []uint64{0}, NewPostings(DocumentID(3), []uint64{1}, nil))},
					TokenID(1): {NewPostings(DocumentID(1), []uint64{1}, NewPostings(DocumentID(3), []uint64{2}, nil))},
					TokenID(2): {NewPostings(DocumentID(1), []uint64{2}, nil)},
				},
			)
			mockStorage.EXPECT().AddToken(gomock.Any()).Return(nil).AnyTimes()
			mockStorage.EXPECT().AddDocument(tt.doc).Return(tt.doc.ID, nil).Times(1)
			mockStorage.EXPECT().GetTokenByTerm("aa").Return(Token{ID: 0, Term: "aa"}, nil).Times(2)
			mockStorage.EXPECT().GetTokenByTerm("bb").Return(Token{ID: 1, Term: "bb"}, nil).Times(1)
			mockStorage.EXPECT().GetTokenByTerm("cc").Return(Token{ID: 2, Term: "cc"}, nil).Times(1)
			mockStorage.EXPECT().GetInvertedIndexByTokenIDs([]TokenID{0, 1, 2}).Return(invertedIndex, nil).Times(1)
			mockStorage.EXPECT().UpsertInvertedIndex(invertedIndex).Times(1)

			// When
			if err := i.AddDocument(tt.doc); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIndexer_UpdateMemoryInvertedIndexByDocument(t *testing.T) {
	cases := []struct {
		doc      Document
		expected InvertedIndex
	}{
		{
			doc: Document{ID: 1, Body: "aa bb cc aa"},
			expected: InvertedIndex{
				0: PostingList{
					Postings: NewPostings(1, []uint64{0, 3}, nil),
				},
				1: PostingList{
					Postings: NewPostings(1, []uint64{1}, nil),
				},
				2: PostingList{
					Postings: NewPostings(1, []uint64{2}, nil),
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("doc = %v, expected = %v", tt.doc, tt.expected), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockStorage := NewMockStorage(mockCtrl)

			// Given
			indexer := Indexer{
				Storage:       mockStorage,
				Analyzer:      Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
				InvertedIndex: InvertedIndex{},
			}
			mockStorage.EXPECT().AddToken(gomock.Any()).Return(nil).AnyTimes()
			mockStorage.EXPECT().GetTokenByTerm("aa").Return(Token{ID: 0, Term: "aa"}, nil).Times(2)
			mockStorage.EXPECT().GetTokenByTerm("bb").Return(Token{ID: 1, Term: "bb"}, nil).Times(1)
			mockStorage.EXPECT().GetTokenByTerm("cc").Return(Token{ID: 2, Term: "cc"}, nil).Times(1)

			// When
			if err := indexer.updateMemoryInvertedIndexByDocument(tt.doc); err != nil {
				t.Error(err)
			}

			// Then
			if diff := cmp.Diff(indexer.InvertedIndex, tt.expected); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})

	}
}

func TestIndexer_UpdateMemoryInvertedIndexByToken(t *testing.T) {
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
					Postings: &Postings{DocumentID: 1, Positions: []uint64{1}, Next: nil},
				},
				TokenID(3): PostingList{
					Postings: &Postings{DocumentID: 1, Positions: []uint64{1}, Next: nil},
				},
				TokenID(4): PostingList{
					Postings: &Postings{DocumentID: 2, Positions: []uint64{1}, Next: nil},
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
					Postings: &Postings{DocumentID: 1, Positions: []uint64{1, 99}, Next: nil},
				},
				TokenID(4): PostingList{
					Postings: &Postings{DocumentID: 2, Positions: []uint64{1}, Next: nil},
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
					Postings: &Postings{DocumentID: 1, Positions: []uint64{1}, Next: nil},
				},
				TokenID(4): PostingList{
					Postings: &Postings{DocumentID: 1, Positions: []uint64{99}, Next: &Postings{DocumentID: 2, Positions: []uint64{1}, Next: nil}},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("docID = %v, token = %v, pos = %v, expected = %v", tt.docID, tt.token, tt.pos, tt.expected), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockStorage := NewMockStorage(mockCtrl)

			// Given
			mockStorage.EXPECT().AddToken(gomock.Any()).Return(nil).AnyTimes()
			mockStorage.EXPECT().GetTokenByTerm("ab").Return(Token{ID: 2, Term: "ab"}, nil).AnyTimes()
			mockStorage.EXPECT().GetTokenByTerm("abc").Return(Token{ID: 3, Term: "abc"}, nil).AnyTimes()
			mockStorage.EXPECT().GetTokenByTerm("abcd").Return(Token{ID: 4, Term: "abcd"}, nil).AnyTimes()
			indexer := Indexer{
				Storage:  mockStorage,
				Analyzer: Analyzer{[]CharFilter{}, NewStandardTokenizer(), []TokenFilter{}},
				InvertedIndex: InvertedIndex{
					TokenID(3): PostingList{
						Postings: &Postings{DocumentID: 1, Positions: []uint64{1}, Next: nil},
					},
					TokenID(4): PostingList{
						Postings: &Postings{DocumentID: 2, Positions: []uint64{1}, Next: nil},
					},
				},
			}

			// When
			if err := indexer.updateMemoryPostingListByToken(tt.docID, tt.token, tt.pos); err != nil {
				t.Error(err)
			}

			// Then
			if diff := cmp.Diff(indexer.InvertedIndex, tt.expected); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}

func TestIndexer_Merge(t *testing.T) {
	cases := []struct {
		memoryInvertedIndex  PostingList
		storageInvertedIndex PostingList
		expected             PostingList
	}{
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
			storageInvertedIndex: PostingList{
				Postings: nil,
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: nil,
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0}, NewPostings(3, []uint64{0}, NewPostings(4, []uint64{3}, nil))),
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(2, []uint64{1, 2}, NewPostings(4, []uint64{3}, NewPostings(5, []uint64{12}, nil))),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0}, NewPostings(2, []uint64{1, 2}, NewPostings(3, []uint64{0}, NewPostings(4, []uint64{3}, NewPostings(5, []uint64{12}, nil))))),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(3, []uint64{0}, NewPostings(4, []uint64{0}, NewPostings(5, []uint64{3}, nil))),
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{1, 2}, NewPostings(2, []uint64{3}, nil)),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{1, 2}, NewPostings(2, []uint64{3}, NewPostings(3, []uint64{0}, NewPostings(4, []uint64{0}, NewPostings(5, []uint64{3}, nil))))),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0, 4}, nil),
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(3, []uint64{0, 1}, nil),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0, 4}, NewPostings(3, []uint64{0, 1}, nil)),
			},
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("memoryInvertedIndex = %v, storageInvertedIndex = %v, expected = %v", tt.memoryInvertedIndex, tt.storageInvertedIndex, tt.expected), func(t *testing.T) {
			merged := merge(tt.memoryInvertedIndex, tt.storageInvertedIndex)
			if diff := cmp.Diff(merged, tt.expected); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}
