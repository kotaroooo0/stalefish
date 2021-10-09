package stalefish

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestMatchAllSearch(t *testing.T) {
	// Mock
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStorage := NewMockStorage(mockCtrl)

	// Given
	doc1 := Document{ID: 1, Body: "aa bb"}
	doc2 := Document{ID: 2, Body: "cc dd"}
	doc3 := Document{ID: 3, Body: "ee ff"}
	docs := []Document{doc1, doc2, doc3}
	mockStorage.EXPECT().GetAllDocuments().Return(docs, nil)

	// When
	matchAllSearcher := NewMatchAllSearcher(mockStorage)
	matchedDocs, err := matchAllSearcher.Search()
	if err != nil {
		t.Fatal(err)
	}

	// Then
	if diff := cmp.Diff(matchedDocs, []Document{doc1, doc2, doc3}); diff != "" {
		t.Errorf("Diff: (-got +want)\n%s", diff)
	}
}

func TestMatchSearch(t *testing.T) {
	doc1 := Document{ID: 1, Body: "aa bb cc"}
	doc2 := Document{ID: 2, Body: "dd ee"}
	doc3 := Document{ID: 3, Body: "ff aa bb"}

	cases := []struct {
		terms        *TokenStream
		logic        Logic
		expectedDocs []Document
	}{
		{
			terms: NewTokenStream(
				[]Token{NewToken("dd")},
			),
			logic:        AND,
			expectedDocs: []Document{doc2},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("aa"), NewToken("bb")},
			),
			logic:        AND,
			expectedDocs: []Document{doc1, doc3},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("aa"), NewToken("dd")},
			),
			logic:        AND,
			expectedDocs: []Document{},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("dd")},
			),
			logic:        OR,
			expectedDocs: []Document{doc2},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("cc"), NewToken("dd")},
			),
			logic:        OR,
			expectedDocs: []Document{doc1, doc2},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("aa"), NewToken("ff")},
			),
			logic:        OR,
			expectedDocs: []Document{doc1, doc3},
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("terms = %v, logic = %v, expectedDocs = %v", tt.terms, tt.logic, tt.expectedDocs), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockStorage := NewMockStorage(mockCtrl)

			// Given
			termToId := map[string]int{"aa": 0, "bb": 1, "cc": 2, "dd": 3, "ee": 4, "ff": 5}
			invertedIndex := InvertedIndex(
				map[TokenID]PostingList{
					TokenID(0): {NewPostings(DocumentID(1), []uint64{0}, NewPostings(DocumentID(3), []uint64{1}, nil))},
					TokenID(1): {NewPostings(DocumentID(1), []uint64{1}, NewPostings(DocumentID(3), []uint64{2}, nil))},
					TokenID(2): {NewPostings(DocumentID(1), []uint64{2}, nil)},
					TokenID(3): {NewPostings(DocumentID(2), []uint64{0}, nil)},
					TokenID(4): {NewPostings(DocumentID(2), []uint64{1}, nil)},
					TokenID(5): {NewPostings(DocumentID(3), []uint64{0}, nil)},
				},
			)

			tokens := []Token{}
			for _, t := range tt.terms.Tokens {
				tId, ok := termToId[t.Term]
				if !ok {
					continue
				}
				tokens = append(tokens, Token{
					ID:   TokenID(tId),
					Term: t.Term,
				})
			}
			mockStorage.EXPECT().GetTokensByTerms(tt.terms.Terms()).Return(tokens, nil)

			ids := make([]TokenID, tt.terms.Size())
			filteredInvertedIndex := make(InvertedIndex)
			for i, t := range tt.terms.Tokens {
				tId := TokenID(termToId[t.Term])
				ids[i] = tId
				filteredInvertedIndex[tId] = invertedIndex[tId]
			}
			mockStorage.EXPECT().GetInvertedIndexByTokenIDs(ids).Return(filteredInvertedIndex, nil)

			docIDs := make([]DocumentID, len(tt.expectedDocs))
			for i, doc := range tt.expectedDocs {
				docIDs[i] = doc.ID
			}
			mockStorage.EXPECT().GetDocuments(docIDs).Return(tt.expectedDocs, nil)

			// When
			matchSearcher := NewMatchSearcher(tt.terms, tt.logic, mockStorage)
			actualDocs, err := matchSearcher.Search()
			if err != nil {
				t.Fatal(err)
			}

			// Then
			if diff := cmp.Diff(tt.expectedDocs, actualDocs); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}

func TestPhraseSearch(t *testing.T) {
	doc1 := Document{ID: 1, Body: "aa bb cc"}
	doc2 := Document{ID: 2, Body: "dd ee"}
	doc3 := Document{ID: 3, Body: "ff aa bb"}

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
				[]Token{NewToken("dd"), NewToken("ee")},
			),
			expectedDocs: []Document{doc2},
		},
		{
			terms: NewTokenStream(
				[]Token{NewToken("ff"), NewToken("aa"), NewToken("bb")},
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
				[]Token{NewToken("ff")},
			), expectedDocs: []Document{doc3},
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("terms = %v, expectedDocs = %v", tt.terms, tt.expectedDocs), func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockStorage := NewMockStorage(mockCtrl)

			// Given
			termToId := map[string]int{"aa": 0, "bb": 1, "cc": 2, "dd": 3, "ee": 4, "ff": 5}
			invertedIndex := InvertedIndex(
				map[TokenID]PostingList{
					TokenID(0): {NewPostings(DocumentID(1), []uint64{0}, NewPostings(DocumentID(3), []uint64{1}, nil))},
					TokenID(1): {NewPostings(DocumentID(1), []uint64{1}, NewPostings(DocumentID(3), []uint64{2}, nil))},
					TokenID(2): {NewPostings(DocumentID(1), []uint64{2}, nil)},
					TokenID(3): {NewPostings(DocumentID(2), []uint64{0}, nil)},
					TokenID(4): {NewPostings(DocumentID(2), []uint64{1}, nil)},
					TokenID(5): {NewPostings(DocumentID(3), []uint64{0}, nil)},
				},
			)

			tokens := []Token{}
			for _, t := range tt.terms.Tokens {
				tId, ok := termToId[t.Term]
				if !ok {
					continue
				}
				tokens = append(tokens, Token{
					ID:   TokenID(tId),
					Term: t.Term,
				})
			}
			mockStorage.EXPECT().GetTokensByTerms(tt.terms.Terms()).Return(tokens, nil)

			ids := make([]TokenID, tt.terms.Size())
			filteredInvertedIndex := make(InvertedIndex)
			for i, t := range tt.terms.Tokens {
				tId := TokenID(termToId[t.Term])
				ids[i] = tId
				filteredInvertedIndex[tId] = invertedIndex[tId]
			}
			mockStorage.EXPECT().GetInvertedIndexByTokenIDs(ids).Return(filteredInvertedIndex, nil)

			docIDs := make([]DocumentID, len(tt.expectedDocs))
			for i, doc := range tt.expectedDocs {
				docIDs[i] = doc.ID
			}
			mockStorage.EXPECT().GetDocuments(docIDs).Return(tt.expectedDocs, nil)

			// When
			phraseSearcher := NewPhraseSearcher(tt.terms, mockStorage)
			actualDocs, err := phraseSearcher.Search()
			if err != nil {
				t.Fatal(err)
			}

			// Then
			if diff := cmp.Diff(tt.expectedDocs, actualDocs); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}
