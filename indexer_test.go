package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k0kubun/pp"
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
	pp.Println(indexer.InvertedIndexMap)
}

func TestUpdateMemoryInvertedIndexByText(t *testing.T) {
	cases := []struct {
		docID  DocumentID
		text   string
		output InvertedIndexMap
	}{
		{
			docID: 1,
			text:  "int string uint string string",
			output: InvertedIndexMap{
				3: InvertedIndexValue{
					Token: Token{ID: 3, Term: "int"},
					PostingList: PostingList{
						Posting{
							DocumentID: 1,
							Positions: []int{
								0,
							},
							PositionsCount: 1,
						},
					},
					DocsCount:      1,
					PositionsCount: 1,
				},
				6: InvertedIndexValue{
					Token: Token{ID: 6, Term: "string"},
					PostingList: PostingList{
						Posting{
							DocumentID: 1,
							Positions: []int{
								1,
								3,
								4,
							},
							PositionsCount: 3,
						},
					},
					DocsCount:      1,
					PositionsCount: 1,
				},
				4: InvertedIndexValue{
					Token: Token{ID: 4, Term: "uint"},
					PostingList: PostingList{
						Posting{
							DocumentID: 1,
							Positions: []int{
								2,
							},
							PositionsCount: 1,
						},
					},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		},
	}

	for _, tt := range cases {
		indexer := Indexer{
			Storage:          TestStorage{},
			Analyzer:         Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			InvertedIndexMap: InvertedIndexMap{},
		}
		indexer.UpdateMemoryInvertedIndexByText(tt.docID, tt.text)
		// pp.Println(indexer.InvertedIndexMap)
		if diff := cmp.Diff(indexer.InvertedIndexMap, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestUpdateMemoryInvertedIndexByToken(t *testing.T) {
	cases := []struct {
		docID  DocumentID
		token  string
		pos    int
		output InvertedIndexMap
	}{
		{
			docID: 1,
			token: "abc",
			pos:   1,
			output: InvertedIndexMap{
				3: InvertedIndexValue{
					Token: Token{ID: 3, Term: "abc"},
					PostingList: PostingList{
						Posting{
							DocumentID: 1,
							Positions: []int{
								1,
							},
							PositionsCount: 1,
						},
					},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		},
		{
			docID: 1,
			token: "abcd",
			pos:   2,
			output: InvertedIndexMap{
				4: InvertedIndexValue{
					Token: Token{ID: 4, Term: "abcd"},
					PostingList: PostingList{
						Posting{
							DocumentID: 1,
							Positions: []int{
								2,
							},
							PositionsCount: 1,
						},
					},
					DocsCount:      1,
					PositionsCount: 1,
				},
			},
		},
	}

	for _, tt := range cases {
		indexer := Indexer{
			Storage:          TestStorage{},
			Analyzer:         Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			InvertedIndexMap: InvertedIndexMap{},
		}
		indexer.UpdateMemoryInvertedIndexByToken(tt.docID, tt.token, tt.pos)
		// pp.Println(indexer.InvertedIndexMap)
		if diff := cmp.Diff(indexer.InvertedIndexMap, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestMergeInvertedIndex(t *testing.T) {
	cases := []struct {
		memoryInvertedIndex  InvertedIndexValue
		storageInvertedIndex InvertedIndexValue
		output               InvertedIndexValue
	}{
		{
			memoryInvertedIndex: InvertedIndexValue{
				Token: Token{ID: 3, Term: "int"},
				PostingList: PostingList{
					Posting{
						DocumentID: 1,
						Positions: []int{
							0,
						},
						PositionsCount: 1,
					},
					Posting{
						DocumentID: 3,
						Positions: []int{
							0,
						},
						PositionsCount: 1,
					},
					Posting{
						DocumentID: 4,
						Positions: []int{
							3,
						},
						PositionsCount: 1,
					},
				},
				DocsCount:      3,
				PositionsCount: 3,
			},
			storageInvertedIndex: InvertedIndexValue{
				Token: Token{ID: 3, Term: "int"},
				PostingList: PostingList{
					Posting{
						DocumentID: 2,
						Positions: []int{
							1, 2,
						},
						PositionsCount: 2,
					},
					Posting{
						DocumentID: 4,
						Positions: []int{
							3,
						},
						PositionsCount: 1,
					},
					Posting{
						DocumentID: 5,
						Positions: []int{
							12,
						},
						PositionsCount: 1,
					},
				},
				DocsCount:      3,
				PositionsCount: 4,
			},
			output: InvertedIndexValue{
				Token: Token{ID: 3, Term: "int"},
				PostingList: PostingList{
					Posting{
						DocumentID: 1,
						Positions: []int{
							0,
						},
						PositionsCount: 1,
					},
					Posting{
						DocumentID: 2,
						Positions: []int{
							1,
							2,
						},
						PositionsCount: 2,
					},
					Posting{
						DocumentID: 3,
						Positions: []int{
							0,
						},
						PositionsCount: 1,
					},
					Posting{
						DocumentID: 4,
						Positions: []int{
							3,
						},
						PositionsCount: 1,
					},
					Posting{
						DocumentID: 5,
						Positions: []int{
							12,
						},
						PositionsCount: 1,
					},
				},
				DocsCount:      5,
				PositionsCount: 6,
			},
		},
	}

	for _, tt := range cases {
		merged, err := MergeInvertedIndex(tt.memoryInvertedIndex, tt.storageInvertedIndex)
		if err != nil {
			t.Error("error: merge failed")
		}
		if diff := cmp.Diff(merged, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
