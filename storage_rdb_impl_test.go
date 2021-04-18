package stalefish

import (
	"math/rand"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"

	"github.com/jmoiron/sqlx"
)

func NewTestDBClient() (*sqlx.DB, error) {
	config := NewDBConfig("root", "password", "127.0.0.1", "3306", "stalefish_test")
	return NewDBClient(config)
}

func truncateTableAll(db *sqlx.DB) error {
	if _, err := db.Exec("truncate table documents"); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table tokens"); err != nil {
		return err
	}
	if _, err := db.Exec("truncate table inverted_indexes"); err != nil {
		return err
	}
	return nil
}

func initDocuments(db *sqlx.DB, docs []Document) error {
	storage := NewStorageRdbImpl(db)
	for _, doc := range docs {
		if _, err := storage.AddDocument(doc); err != nil {
			return err
		}
	}
	return nil
}

func initTokens(db *sqlx.DB) error {
	storage := NewStorageRdbImpl(db)

	tokens := []Token{
		NewToken("term1"),
		NewToken("term2"),
	}
	for _, token := range tokens {
		if _, err := storage.AddToken(token); err != nil {
			return err
		}
	}
	return nil
}

func TestGetAllDocuments(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	docs := []Document{
		NewDocument("body1"),
		NewDocument("body2"),
		NewDocument("body3"),
	}
	if err := initDocuments(db, docs); err != nil {
		t.Fatal(err)
	}

	expectedDoc1 := Document{ID: 1, Body: "body1"}
	expectedDoc2 := Document{ID: 2, Body: "body2"}
	expectedDoc3 := Document{ID: 3, Body: "body3"}

	cases := []struct {
		expected []Document
	}{
		{
			expected: []Document{expectedDoc1, expectedDoc2, expectedDoc3},
		},
	}
	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		docs, err := storage.GetAllDocuments()
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(docs, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestGetDocuments(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	docs := []Document{
		NewDocument("body1"),
		NewDocument("body2"),
		NewDocument("body3"),
	}
	if err := initDocuments(db, docs); err != nil {
		t.Fatal(err)
	}

	expectedDoc1 := Document{ID: 1, Body: "body1"}
	expectedDoc2 := Document{ID: 2, Body: "body2"}
	expectedDoc3 := Document{ID: 3, Body: "body3"}

	cases := []struct {
		ids      []DocumentID
		expected []Document
	}{
		{
			ids:      []DocumentID{1, 2, 3},
			expected: []Document{expectedDoc1, expectedDoc2, expectedDoc3},
		},
		{
			ids:      []DocumentID{3},
			expected: []Document{expectedDoc3},
		},
		{
			ids:      []DocumentID{1, 3},
			expected: []Document{expectedDoc1, expectedDoc3},
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		docs, err := storage.GetDocuments(tt.ids)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(docs, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestAddDocument(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		doc      Document
		expected DocumentID
	}{
		{
			doc:      NewDocument("body1"),
			expected: 1,
		},
		{
			doc:      NewDocument("body2"),
			expected: 2,
		},
		{
			doc:      NewDocument("body3"),
			expected: 3,
		},
		{
			doc:      NewDocument("body4"),
			expected: 4,
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		id, err := storage.AddDocument(tt.doc)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(id, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestAddToken(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		token    Token
		expected TokenID
	}{
		{
			token:    NewToken("token1"),
			expected: 1,
		},
		{
			token:    NewToken("token2"),
			expected: 2,
		},
		{
			token:    NewToken("token3"),
			expected: 3,
		},
		{
			token:    NewToken("token4"),
			expected: 4,
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		id, err := storage.AddToken(tt.token)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(id, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestGetTokenByTerm(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	if err := initTokens(db); err != nil {
		t.Fatal(err)
	}

	expectedToken1 := Token{ID: 1, Term: "term1"}
	expectedToken2 := Token{ID: 2, Term: "term2"}

	cases := []struct {
		term     string
		expected Token
	}{
		{
			term:     "term1",
			expected: expectedToken1,
		},
		{
			term:     "term2",
			expected: expectedToken2,
		},
		{
			term:     "term3",
			expected: Token{},
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		token, err := storage.GetTokenByTerm(tt.term)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(token, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestGetTokensByTerms(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	if err := initTokens(db); err != nil {
		t.Fatal(err)
	}

	expectedToken1 := Token{ID: 1, Term: "term1"}
	expectedToken2 := Token{ID: 2, Term: "term2"}

	cases := []struct {
		terms    []string
		expected []Token
	}{
		{
			terms:    []string{"term1"},
			expected: []Token{expectedToken1},
		},
		{
			terms:    []string{"term1", "term2"},
			expected: []Token{expectedToken1, expectedToken2},
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		tokens, err := storage.GetTokensByTerms(tt.terms)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(tokens, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestGetInvertedIndexByTokenIDs(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	inverted := NewInvertedIndex(
		map[TokenID]PostingList{
			TokenID(777): NewPostingList(
				NewPostings(1, []uint64{1, 2, 3, 4}, 4, NewPostings(100, []uint64{11, 22}, 2, NewPostings(250, []uint64{11, 15, 22}, 3, nil))),
				99,
				999),
		},
	)

	storage := NewStorageRdbImpl(db)
	if storage.UpsertInvertedIndex(inverted) != nil {
		t.Fatal(err)
	}

	cases := []struct {
		tokenIDs []TokenID
		expected InvertedIndex
	}{
		{
			tokenIDs: []TokenID{TokenID(777)},
			expected: NewInvertedIndex(
				map[TokenID]PostingList{
					TokenID(777): NewPostingList(
						NewPostings(1, []uint64{1, 2, 3, 4}, 4, NewPostings(100, []uint64{11, 22}, 2, NewPostings(250, []uint64{11, 15, 22}, 3, nil))),
						99,
						999),
				},
			),
		},
	}

	for _, tt := range cases {
		actual, err := storage.GetInvertedIndexByTokenIDs(tt.tokenIDs)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(actual, tt.expected); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestUpsertInvertedIndex(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	inverted := NewInvertedIndex(
		map[TokenID]PostingList{
			TokenID(777): NewPostingList(
				NewPostings(1, []uint64{1, 2, 3, 4}, 4, NewPostings(3, []uint64{11, 22}, 2, NewPostings(5, []uint64{11, 15, 22}, 3, nil))),
				99,
				999,
			),
			TokenID(888): NewPostingList(
				NewPostings(4, []uint64{3, 4}, 3, NewPostings(634, []uint64{11, 22, 444}, 3, NewPostings(421421, []uint64{11, 22}, 2, nil))),
				99,
				999,
			),
		},
	)

	cases := []struct {
		invertedIndex InvertedIndex
		err           error
	}{
		{
			invertedIndex: inverted,
			err:           nil,
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		if err := storage.UpsertInvertedIndex(tt.invertedIndex); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(err, tt.err); diff != "" {
			t.Fatalf("Diff: (-got +want)\n%s", diff)
		}
	}
}

// NOTE: 転置インデックスのサイズを計測するため
func TestCompressedIndex(t *testing.T) {
	// NOTE: テストを動かしたい時はコメントアウトする
	return

	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}

	storage := NewStorageRdbImpl(db)
	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, 100)
	for i := 0; i < 3000; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()
			inverted := NewInvertedIndex(map[TokenID]PostingList{
				TokenID(id): NewPostingList(
					createHeavyPostings(),
					1,
					2,
				),
			},
			)
			if err := storage.UpsertInvertedIndex(inverted); err != nil {
				t.Fatal(err)
			}
		}(i)
	}
	wg.Wait()
}

func createHeavyPostings() *Postings {
	var root *Postings = NewPostings(DocumentID(0), randUint64Slice(), 99, nil)
	var p *Postings = root
	for i := 0; i < 5000; i++ {
		p.Next = NewPostings(DocumentID(i*10), randUint64Slice(), 99, nil)
		p = p.Next
	}
	return root
}

func randUint64Slice() []uint64 {
	size := 3
	ret := make([]uint64, size)
	for i := 0; i < size; i++ {
		ret[i] = rand.Uint64()
	}
	return ret
}
