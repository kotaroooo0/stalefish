package stalefish

import (
	"bytes"
	"encoding/gob"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

func NewDBClient(dbConfig *DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbConfig.User, dbConfig.Password, dbConfig.Addr, dbConfig.Port, dbConfig.DB),
	)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return db, nil
}

type StorageRdbImpl struct {
	DB *sqlx.DB
}

func NewStorageRdbImpl(db *sqlx.DB) StorageRdbImpl {
	return StorageRdbImpl{
		DB: db,
	}
}

type DBConfig struct {
	User     string
	Password string
	Addr     string
	Port     string
	DB       string
}

func NewDBConfig(user, password, addr, port, db string) *DBConfig {
	return &DBConfig{
		User:     user,
		Password: password,
		Addr:     addr,
		Port:     port,
		DB:       db,
	}
}

func (s StorageRdbImpl) GetAllDocuments() ([]Document, error) {
	var docs []Document
	if err := s.DB.Select(&docs, `select * from documents`); err != nil {
		return nil, errors.New(err.Error())
	}
	return docs, nil
}

func (s StorageRdbImpl) GetDocuments(ids []DocumentID) ([]Document, error) {
	intDocIDs := make([]int, len(ids))
	for i, id := range ids {
		intDocIDs[i] = int(id)
	}

	sql, params, err := sqlx.In(`select * from documents where id in (?)`, intDocIDs)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var docs []Document
	if err = s.DB.Select(&docs, sql, params...); err != nil {
		return nil, errors.New(err.Error())
	}
	return docs, nil
}

func (s StorageRdbImpl) AddDocument(doc Document) (DocumentID, error) {
	res, err := s.DB.NamedExec(`insert into documents (body) values (:body)`,
		map[string]interface{}{
			"body": doc.Body,
		})
	if err != nil {
		return 0, errors.New(err.Error())
	}

	insertedID, err := res.LastInsertId()
	if err != nil {
		return 0, errors.New(err.Error())
	}
	return DocumentID(insertedID), nil
}

// TODO: 同じトークンでもIDがインクリメントされ、IDがとびとびになる
func (s StorageRdbImpl) AddToken(token Token) (TokenID, error) {
	res, err := s.DB.NamedExec(`insert into tokens (term) values (:term)`,
		map[string]interface{}{
			"term": token.Term,
		})
	if err != nil {
		return 0, errors.New(err.Error())
	}
	insertedID, err := res.LastInsertId()
	if err != nil {
		return 0, errors.New(err.Error())
	}
	return TokenID(insertedID), nil
}

func (s StorageRdbImpl) GetTokenByTerm(term string) (Token, error) {
	var tokens []Token
	err := s.DB.Select(&tokens, `select * from tokens where term = ?`, term)
	if err != nil {
		return Token{}, err
	}
	switch len(tokens) {
	case 0:
		return Token{}, nil
	case 1:
		return tokens[0], nil
	default:
		return Token{}, errors.New("error: two or more hits(inconsistent match result)")
	}
}

func (s StorageRdbImpl) UpsertInvertedIndex(tokenID TokenID, invertedIndexValue InvertedIndexValue) error {
	encoded, err := invertedIndexValue.encode()
	if err != nil {
		return errors.New(err.Error())
	}
	_, err = s.DB.NamedExec(
		`insert into inverted_indexes (token_id, posting_list, docs_count, positions_count)
		values (:token_id, :posting_list, :docs_count, :positions_count)
		on duplicate key update posting_list = :posting_list, docs_count = :docs_count, positions_count = :positions_count`,
		map[string]interface{}{
			"token_id":        tokenID,
			"posting_list":    encoded.PostingList,
			"docs_count":      encoded.DocsCount,
			"positions_count": encoded.PositionsCount,
		})
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (s StorageRdbImpl) GetInvertedIndexByTokenID(tokenID TokenID) (InvertedIndexValue, error) {
	var encodedInvertedIndexs []EncodedInvertedIndex

	err := s.DB.Select(&encodedInvertedIndexs,
		`select
			posting_list,
			docs_count,
			positions_count
		from
			inverted_indexes
		where
			token_id = ?`, int(tokenID))
	if err != nil {
		return InvertedIndexValue{}, errors.New(err.Error())
	}

	switch len(encodedInvertedIndexs) {
	case 0:
		return InvertedIndexValue{}, nil
	case 1:
		return encodedInvertedIndexs[0].decode()
	default:
		return InvertedIndexValue{}, errors.New("error: two or more hits(inconsistent match result)")
	}
}

func (i InvertedIndexValue) encode() (EncodedInvertedIndex, error) {
	// 差分を取る
	var p *Postings = i.PostingList
	var beforeDocumentID DocumentID = 0
	for p != nil {
		p.DocumentID -= beforeDocumentID
		beforeDocumentID = p.DocumentID + beforeDocumentID
		p = p.Next
	}

	// Gobでシリアライズ&圧縮
	plBuf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(plBuf).Encode(i.PostingList); err != nil {
		return EncodedInvertedIndex{}, errors.New(err.Error())
	}
	return NewEncodedInvertedIndex(plBuf.Bytes(), i.DocsCount, i.PositionsCount), nil
}

type EncodedInvertedIndex struct {
	PostingList    []byte `db:"posting_list"`    // トークンを含むポスティングスリスト
	DocsCount      uint64 `db:"docs_count"`      // トークンを含む文書数
	PositionsCount uint64 `db:"positions_count"` // 全文書内でのトークンの出現数
}

func NewEncodedInvertedIndex(pl []byte, docsCount, positionsCount uint64) EncodedInvertedIndex {
	return EncodedInvertedIndex{
		PostingList:    pl,
		DocsCount:      docsCount,
		PositionsCount: positionsCount,
	}
}

func (e EncodedInvertedIndex) decode() (InvertedIndexValue, error) {
	// Gobでデシリアライズ
	pl := &Postings{}
	ret := bytes.NewBuffer(e.PostingList)
	if err := gob.NewDecoder(ret).Decode(pl); err != nil {
		return InvertedIndexValue{}, errors.New(err.Error())
	}
	inverted := NewInvertedIndexValue(pl, e.DocsCount, e.PositionsCount)

	// 差分から本来のIDへ変換
	var c *Postings = inverted.PostingList
	var beforeDocumentID DocumentID = 0
	for c != nil {
		c.DocumentID += beforeDocumentID
		beforeDocumentID = c.DocumentID
		c = c.Next
	}
	return inverted, nil
}
