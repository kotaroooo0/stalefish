package stalefish

import (
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

func TestGetDocuments(t *testing.T) {
	// TODO: before()作って共通化
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	db.Exec("truncate table documents")
	storage := NewStorageRdbImpl(db)

	// TODO: テストデータ生成方法
	doc1 := NewDocument("title1")
	doc2 := NewDocument("title2")
	doc3 := NewDocument("title3")
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

func TestStorageAddDocument(t *testing.T) {
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	db.Exec("truncate table documents")
	storage := NewStorageRdbImpl(db)

	cases := []struct {
		doc Document
		id  DocumentID
	}{
		{
			doc: NewDocument("title1"),
			id:  1,
		},
		{
			doc: NewDocument("title2"),
			id:  2,
		},
		{
			doc: NewDocument("title3"),
			id:  3,
		},
		{
			doc: NewDocument("title4"),
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

func TestGetTokenByTerm(t *testing.T) {
	// NOTE:ここから実装
	db, err := NewTestDBClient()
	if err != nil {
		t.Error(err)
	}
	db.Exec("truncate table tokens")
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
	db.Exec("truncate table inverted_indexes")
	storage := NewStorageRdbImpl(db)

	p1 := Posting{
		DocumentID:     1,
		Positions:      []int{1, 2, 3, 4},
		PositionsCount: 4,
	}
	p2 := Posting{
		DocumentID:     3,
		Positions:      []int{11, 22},
		PositionsCount: 2,
	}
	inverted := InvertedIndexValue{
		Token:          Token{ID: 12, Term: "hoge"},
		PostingList:    []Posting{p1, p2},
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
	db.Exec("truncate table tokens")
	db.Exec("truncate table inverted_indexes")
	storage := NewStorageRdbImpl(db)

	p1 := Posting{
		DocumentID:     1,
		Positions:      []int{1, 2, 3, 4},
		PositionsCount: 4,
	}
	p2 := Posting{
		DocumentID:     3,
		Positions:      []int{11, 22},
		PositionsCount: 2,
	}
	token := Token{ID: 1, Term: "hoge"}
	inverted := InvertedIndexValue{
		Token:          token,
		PostingList:    []Posting{p1, p2},
		DocsCount:      123,
		PositionsCount: 11,
	}

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
			tokenID:            TokenID(1),
			invertedIndexValue: inverted,
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
