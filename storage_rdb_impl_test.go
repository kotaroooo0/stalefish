package stalefish

import (
	"math/rand"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"

	"github.com/jmoiron/sqlx"
)

// TODO: DBを叩きに行ってるいるのでモックする
func NewTestDBClient() (*sqlx.DB, error) {
	config := NewDBConfig("root", "password", "127.0.0.1", "3306", "stalefish")
	return NewDBClient(config)
}

func truncateTableAll(db *sqlx.DB) {
	db.Exec("truncate table documents")
	db.Exec("truncate table tokens")
	db.Exec("truncate table inverted_indexes")
}

func TestGetDocuments(t *testing.T) {
	// TODO: before()作って共通化
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)

	storage := NewStorageRdbImpl(db)

	// TODO: テストデータ生成方法
	doc1 := NewDocument("body1")
	doc2 := NewDocument("body2")
	doc3 := NewDocument("body3")
	storage.AddDocument(doc1)
	storage.AddDocument(doc2)
	storage.AddDocument(doc3)
	expectedDoc1 := Document{
		ID:   1,
		Body: "body1",
	}
	expectedDoc2 := Document{
		ID:   2,
		Body: "body2",
	}
	expectedDoc3 := Document{
		ID:   3,
		Body: "body3",
	}

	cases := []struct {
		ids  []DocumentID
		docs []Document
	}{
		{
			ids:  []DocumentID{1, 2, 3},
			docs: []Document{expectedDoc1, expectedDoc2, expectedDoc3},
		},
		{
			ids:  []DocumentID{3},
			docs: []Document{expectedDoc3},
		},
		{
			ids:  []DocumentID{1, 3},
			docs: []Document{expectedDoc1, expectedDoc3},
		},
	}

	for _, tt := range cases {
		docs, err := storage.GetDocuments(tt.ids)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(docs, tt.docs); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestAddDocument(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)
	storage := NewStorageRdbImpl(db)

	cases := []struct {
		doc Document
		id  DocumentID
	}{
		{
			doc: NewDocument("body1"),
			id:  1,
		},
		{
			doc: NewDocument("body2"),
			id:  2,
		},
		{
			doc: NewDocument("body3"),
			id:  3,
		},
		{
			doc: NewDocument("body4"),
			id:  4,
		},
	}

	for _, tt := range cases {
		id, err := storage.AddDocument(tt.doc)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(id, tt.id); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestAddToken(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)
	storage := NewStorageRdbImpl(db)

	cases := []struct {
		token Token
		id    TokenID
	}{
		{
			token: NewToken("token1"),
			id:    1,
		},
		{
			token: NewToken("token2"),
			id:    2,
		},
		{
			token: NewToken("token3"),
			id:    3,
		},
		{
			token: NewToken("token4"),
			id:    4,
		},
	}

	for _, tt := range cases {
		id, err := storage.AddToken(tt.token)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(id, tt.id); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestGetTokenByTerm(t *testing.T) {
	// NOTE:ここから実装
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)
	storage := NewStorageRdbImpl(db)

	// TODO: テストデータ生成方法
	token1 := NewToken("term1")
	token2 := NewToken("term2")
	storage.AddToken(token1)
	storage.AddToken(token2)
	expectedToken1 := Token{
		ID:   1,
		Term: "term1",
	}
	expectedToken2 := Token{
		ID:   2,
		Term: "term2",
	}

	cases := []struct {
		term  string
		token Token
	}{
		{
			term:  "term1",
			token: expectedToken1,
		},
		{
			term:  "term2",
			token: expectedToken2,
		},
		{
			term:  "term3",
			token: Token{},
		},
	}

	for _, tt := range cases {
		token, err := storage.GetTokenByTerm(tt.term)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(token, tt.token); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestUpsertInvertedIndex(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)
	storage := NewStorageRdbImpl(db)

	inverted := InvertedIndexValue{
		Token:          Token{ID: 12, Term: "hoge"},
		PostingList:    NewPostings(1, []uint64{1, 2, 3, 4}, 4, NewPostings(3, []uint64{11, 22}, 2, nil)),
		DocsCount:      123,
		PositionsCount: 11,
	}

	cases := []struct {
		invertedIndex InvertedIndexValue
		err           error
	}{
		{
			invertedIndex: inverted,
			err:           nil,
		},
	}

	for _, tt := range cases {
		err := storage.UpsertInvertedIndex(tt.invertedIndex)
		if diff := cmp.Diff(err, tt.err); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestGetInvertedIndexByTokenID(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)

	storage := NewStorageRdbImpl(db)

	token := Token{ID: 1, Term: "hoge"}
	inverted := NewInvertedIndexValue(
		token,
		NewPostings(1, []uint64{1, 2, 3, 4}, 4, NewPostings(3, []uint64{11, 22}, 2, NewPostings(5, []uint64{11, 15, 22}, 3, nil))),
		123,
		11,
	)
	err = storage.UpsertInvertedIndex(inverted)
	if err != nil {
		t.Error(err)
	}
	_, err = storage.AddToken(token)
	if err != nil {
		t.Error(err)
	}

	cases := []struct {
		tokenID            TokenID
		invertedIndexValue InvertedIndexValue
	}{
		{
			tokenID: TokenID(1),
			invertedIndexValue: NewInvertedIndexValue(
				token,
				NewPostings(1, []uint64{1, 2, 3, 4}, 4, NewPostings(3, []uint64{11, 22}, 2, NewPostings(5, []uint64{11, 15, 22}, 3, nil))),
				123,
				11),
		},
	}

	for _, tt := range cases {
		actual, err := storage.GetInvertedIndexByTokenID(tt.tokenID)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(actual, tt.invertedIndexValue); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func TestCompressedIndex(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	truncateTableAll(db)

	storage := NewStorageRdbImpl(db)
	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, 100)
	for i := 0; i < 3000; i++ {
		sem <- struct{}{}

		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()
			token := Token{ID: TokenID(id), Term: "hoge"}
			inverted := NewInvertedIndexValue(
				token,
				createHeavyPostingList(),
				1,
				2,
			)
			err = storage.UpsertInvertedIndex(inverted)
			if err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
}

func createHeavyPostingList() *Postings {
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
