package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestStorage struct {
	Storage
}

func (s TestStorage) GetTokenID(token string) TokenID {
	return TokenID(len(token))
}
func TestTextToPostingList(t *testing.T) {
	cases := []struct {
		docID  DocumentID
		text   string
		output InvertIndexHash
	}{
		{
			docID: 1,
			text:  "int string uint string string",
			output: InvertIndexHash{
				3: InvertedIndexValue{
					TokenID: 3,
					Token:   "int",
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
					TokenID: 6,
					Token:   "string",
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
					TokenID: 4,
					Token:   "uint",
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
			Storage:         TestStorage{},
			Analyzer:        Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			InvertIndexHash: InvertIndexHash{},
		}
		indexer.textToPostingLists(tt.docID, tt.text)
		// pp.Println(indexer.InvertIndexHash)
		if diff := cmp.Diff(indexer.InvertIndexHash, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestTokenToPostingList(t *testing.T) {
	cases := []struct {
		docID  DocumentID
		token  string
		pos    int
		output InvertIndexHash
	}{
		{
			docID: 1,
			token: "abc",
			pos:   1,
			output: InvertIndexHash{
				3: InvertedIndexValue{
					TokenID: 3,
					Token:   "abc",
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
			output: InvertIndexHash{
				4: InvertedIndexValue{
					TokenID: 4,
					Token:   "abcd",
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
			Storage:         TestStorage{},
			Analyzer:        Analyzer{[]CharFilter{}, StandardTokenizer{}, []TokenFilter{}},
			InvertIndexHash: InvertIndexHash{},
		}
		indexer.tokenToPostingList(tt.docID, tt.token, tt.pos)
		// pp.Println(indexer.InvertIndexHash)
		if diff := cmp.Diff(indexer.InvertIndexHash, tt.output); diff != "" {
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
				TokenID: 3,
				Token:   "int",
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
				TokenID: 3,
				Token:   "int",
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
				TokenID: 3,
				Token:   "int",
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
		merged, err := mergeInvertedIndex(tt.memoryInvertedIndex, tt.storageInvertedIndex)
		if err != nil {
			t.Error("error: merge failed")
		}
		if diff := cmp.Diff(merged, tt.output); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
