package stalefish

import (
	"fmt"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestTfIdfSorter_Sort(t *testing.T) {
	docs := []Document{
		{ID: 1, Body: "りんご　みかん", TokenCount: 2},
		{ID: 2, Body: "りんご　りんご　みかん", TokenCount: 3},
		{ID: 3, Body: "りんご　りんご　みかん　みかん　みかん", TokenCount: 5},
	}
	invertedIndex := map[TokenID]PostingList{
		1: {NewPostings(1, []uint64{0}, NewPostings(2, []uint64{0, 1}, NewPostings(3, []uint64{0, 1}, nil)))},
		2: {NewPostings(1, []uint64{1}, NewPostings(2, []uint64{2}, NewPostings(3, []uint64{2, 3, 4}, nil)))},
	}
	tests := []struct {
		docs          []Document
		invertedIndex InvertedIndex
		tokens        []Token
		expected      []Document
	}{
		{
			docs:          docs,
			invertedIndex: invertedIndex,
			tokens:        []Token{{ID: 1, Term: "りんご"}},
			expected: []Document{
				{ID: 2, Body: "りんご　りんご　みかん", TokenCount: 3},
				{ID: 1, Body: "りんご　みかん", TokenCount: 2},
				{ID: 3, Body: "りんご　りんご　みかん　みかん　みかん", TokenCount: 5},
			},
		},
		{
			docs:          docs,
			invertedIndex: invertedIndex,
			tokens:        []Token{{ID: 2, Term: "みかん"}},
			expected: []Document{
				{ID: 3, Body: "りんご　りんご　みかん　みかん　みかん", TokenCount: 5},
				{ID: 1, Body: "りんご　みかん", TokenCount: 2},
				{ID: 2, Body: "りんご　りんご　みかん", TokenCount: 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("docs = %v, invertedIndex = %v, tokens = %v, expected = %v", tt.docs, tt.invertedIndex, tt.tokens, tt.expected), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockStorage := NewMockStorage(mockCtrl)

			// Then
			s := &TfIdfSorter{
				Storage: mockStorage,
			}
			mockStorage.EXPECT().CountDocuments().Return(3, nil)

			// When
			got, err := s.Sort(tt.docs, tt.invertedIndex, tt.tokens)
			if err != nil {
				t.Fatal(err)
			}

			// Then
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("TfIdfSorter.Sort() = %v, want %v", got, tt.expected)
			}
		})
	}
}
