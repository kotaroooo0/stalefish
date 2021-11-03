package stalefish

import (
	"fmt"
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

func insertDocuments(db *sqlx.DB, docs []Document) error {
	for _, doc := range docs {
		if _, err := db.NamedExec(`insert into documents (body, token_count) values (:body, :token_count)`, map[string]interface{}{"body": doc.Body, "token_count": doc.TokenCount}); err != nil {
			return err
		}
	}
	return nil
}

func insertTokens(db *sqlx.DB, tokens []Token) error {
	for _, token := range tokens {
		if _, err := db.NamedExec(`insert into tokens (term) values (:term)`, map[string]interface{}{"term": token.Term}); err != nil {
			return err
		}
	}
	return nil
}

func insertInvertedIndex(db *sqlx.DB, invertedIndex InvertedIndex) error {
	encoded, err := encode(invertedIndex)
	if err != nil {
		return err
	}
	for _, v := range encoded {
		if _, err := db.NamedExec(
			`insert into inverted_indexes (token_id, posting_list)
			values (:token_id, :posting_list)
			on duplicate key update posting_list = :posting_list`, v); err != nil {
			return err
		}
	}
	return nil
}

func TestStorageRdbImpl_CountDocuments(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	if err := insertDocuments(db, []Document{
		{Body: "TestGetAllDocuments1", TokenCount: 1},
		{Body: "TestGetAllDocuments2", TokenCount: 2},
		{Body: "TestGetAllDocuments3", TokenCount: 3},
		{Body: "TestGetAllDocuments4", TokenCount: 3},
		{Body: "TestGetAllDocuments5", TokenCount: 3},
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		expected int
	}{
		{
			expected: 5,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("expected = %v", tt.expected), func(t *testing.T) {
			s := &StorageRdbImpl{
				DB: db,
			}
			got, err := s.CountDocuments()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.expected {
				t.Errorf("StorageRdbImpl.CountDocuments() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestGetAllDocuments(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	if err := insertDocuments(db, []Document{
		{Body: "TestGetAllDocuments1", TokenCount: 1},
		{Body: "TestGetAllDocuments2", TokenCount: 2},
		{Body: "TestGetAllDocuments3", TokenCount: 3},
	}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		expected []Document
	}{
		{
			expected: []Document{
				{ID: 1, Body: "TestGetAllDocuments1", TokenCount: 1},
				{ID: 2, Body: "TestGetAllDocuments2", TokenCount: 2},
				{ID: 3, Body: "TestGetAllDocuments3", TokenCount: 3},
			},
		},
	}
	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		t.Run(fmt.Sprintf("expected = %v", tt.expected), func(t *testing.T) {
			docs, err := storage.GetAllDocuments()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(docs, tt.expected); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
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
	if err := insertDocuments(db, []Document{
		{Body: "TestGetDocuments1", TokenCount: 4},
		{Body: "TestGetDocuments2", TokenCount: 5},
		{Body: "TestGetDocuments3", TokenCount: 6},
	}); err != nil {
		t.Fatal(err)
	}

	expectedDoc1 := Document{ID: 1, Body: "TestGetDocuments1", TokenCount: 4}
	expectedDoc2 := Document{ID: 2, Body: "TestGetDocuments2", TokenCount: 5}
	expectedDoc3 := Document{ID: 3, Body: "TestGetDocuments3", TokenCount: 6}

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
		t.Run(fmt.Sprintf("ids = %v, expected = %v", tt.ids, tt.expected), func(t *testing.T) {
			docs, err := storage.GetDocuments(tt.ids)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(docs, tt.expected); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
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
			doc:      Document{Body: "TestAddDocument1", TokenCount: 3},
			expected: 1,
		},
		{
			doc:      Document{Body: "TestAddDocument2", TokenCount: 6},
			expected: 2,
		},
		{
			doc:      Document{Body: "TestAddDocument3", TokenCount: 9},
			expected: 3,
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		t.Run(fmt.Sprintf("doc = %v, expected = %v", tt.doc, tt.expected), func(t *testing.T) {
			id, err := storage.AddDocument(tt.doc)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(id, tt.expected); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}

func TestStorageRdbImpl_AddToken(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		token Token
	}
	tests := []struct {
		token   Token
		want    TokenID
		wantErr bool
	}{
		{
			token:   NewToken("TestAddToken1"),
			want:    1,
			wantErr: false,
		},
		{
			token:   NewToken("TestAddToken2"),
			want:    2,
			wantErr: false,
		},
		{
			token:   NewToken("TestAddToken2"),
			want:    0,
			wantErr: true,
		},
	}
	db, err := NewTestDBClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := truncateTableAll(db); err != nil {
		t.Fatal(err)
	}
	storage := NewStorageRdbImpl(db)
	for _, tt := range tests {
		t.Run(fmt.Sprintf("token = %v", tt.token), func(t *testing.T) {
			got, err := storage.AddToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("StorageRdbImpl.AddToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StorageRdbImpl.AddToken() = %v, want %v", got, tt.want)
			}
		})
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
	if err := insertTokens(db, []Token{
		NewToken("term1"),
		NewToken("term2"),
	}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		term     string
		expected *Token
	}{
		{
			term:     "term1",
			expected: &Token{ID: 1, Term: "term1"},
		},
		{
			term:     "term2",
			expected: &Token{ID: 2, Term: "term2"},
		},
		{
			term:     "term3",
			expected: nil,
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		t.Run(fmt.Sprintf("term = %v, expected = %v", tt.term, tt.expected), func(t *testing.T) {
			token, err := storage.GetTokenByTerm(tt.term)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(token, tt.expected); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
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
	if err := insertTokens(db, []Token{
		NewToken("term1"),
		NewToken("term2"),
	}); err != nil {
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
		t.Run(fmt.Sprintf("terms = %v, expected = %v", tt.terms, tt.expected), func(t *testing.T) {
			tokens, err := storage.GetTokensByTerms(tt.terms)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tokens, tt.expected); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
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
	invertedIndex := NewInvertedIndex(
		map[TokenID]PostingList{
			TokenID(777): NewPostingList(
				NewPostings(1, []uint64{1, 2, 3, 4}, NewPostings(100, []uint64{11, 22}, NewPostings(250, []uint64{11, 15, 22}, nil))),
			),
		},
	)
	if err := insertInvertedIndex(db, invertedIndex); err != nil {
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
						NewPostings(1, []uint64{1, 2, 3, 4}, NewPostings(100, []uint64{11, 22}, NewPostings(250, []uint64{11, 15, 22}, nil))),
					),
				},
			),
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		t.Run(fmt.Sprintf("tokenIDs = %v, expected = %v", tt.tokenIDs, tt.expected), func(t *testing.T) {
			actual, err := storage.GetInvertedIndexByTokenIDs(tt.tokenIDs)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(actual, tt.expected); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
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
	invertedIndex := NewInvertedIndex(
		map[TokenID]PostingList{
			TokenID(777): NewPostingList(
				NewPostings(1, []uint64{1, 2, 3, 4}, NewPostings(3, []uint64{11, 22}, NewPostings(5, []uint64{11, 15, 22}, nil))),
			),
			TokenID(888): NewPostingList(
				NewPostings(4, []uint64{3, 4}, NewPostings(634, []uint64{11, 22, 444}, NewPostings(421421, []uint64{11, 22}, nil))),
			),
		},
	)

	cases := []struct {
		invertedIndex InvertedIndex
		err           error
	}{
		{
			invertedIndex: invertedIndex,
			err:           nil,
		},
	}

	storage := NewStorageRdbImpl(db)
	for _, tt := range cases {
		t.Run(fmt.Sprintf("invertedIndex = %v, err = %v", tt.invertedIndex, tt.err), func(t *testing.T) {
			if err := storage.UpsertInvertedIndex(tt.invertedIndex); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(err, tt.err); diff != "" {
				t.Fatalf("Diff: (-got +want)\n%s", diff)
			}
		})
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
	var root *Postings = NewPostings(DocumentID(0), randUint64Slice(), nil)
	var p *Postings = root
	for i := 0; i < 5000; i++ {
		p.Next = NewPostings(DocumentID(i*10), randUint64Slice(), nil)
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
