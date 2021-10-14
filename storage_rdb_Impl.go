package stalefish

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewDBClient(dbConfig *DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbConfig.User, dbConfig.Password, dbConfig.Addr, dbConfig.Port, dbConfig.DB),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type StorageRdbImpl struct {
	DB *sqlx.DB
}

func NewStorageRdbImpl(db *sqlx.DB) *StorageRdbImpl {
	return &StorageRdbImpl{
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

func (s *StorageRdbImpl) CountDocuments() (int, error) {
	var count int
	row := s.DB.QueryRow(`select count(*) from documents`)
	if err := row.Scan(&count); err != nil {
		return -1, err
	}
	return count, nil
}

func (s *StorageRdbImpl) GetAllDocuments() ([]Document, error) {
	var docs []Document
	if err := s.DB.Select(&docs, `select * from documents`); err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *StorageRdbImpl) GetDocuments(ids []DocumentID) ([]Document, error) {
	if len(ids) == 0 {
		return []Document{}, nil
	}
	intDocIDs := make([]int, len(ids))
	for i, id := range ids {
		intDocIDs[i] = int(id)
	}

	sql, params, err := sqlx.In(`select * from documents where id in (?)`, intDocIDs)
	if err != nil {
		return nil, err
	}
	var docs []Document
	if err = s.DB.Select(&docs, sql, params...); err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *StorageRdbImpl) AddDocument(doc Document) (DocumentID, error) {
	res, err := s.DB.NamedExec(`insert into documents (body, token_count) values (:body, :token_count)`,
		map[string]interface{}{
			"body":        doc.Body,
			"token_count": doc.TokenCount,
		})
	if err != nil {
		return 0, err
	}

	insertedID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return DocumentID(insertedID), nil
}

func (s *StorageRdbImpl) AddToken(token Token) error {
	res, err := s.DB.NamedExec(`insert into tokens (term) values (:term)`,
		map[string]interface{}{
			"term": token.Term,
		},
	)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return nil
			}
		}
		return err
	}

	if _, err := res.LastInsertId(); err != nil {
		return err
	}
	return nil
}

func (s *StorageRdbImpl) GetTokenByTerm(term string) (Token, error) {
	var token Token
	if err := s.DB.Get(&token, `select * from tokens where term = ?`, term); err != nil {
		if err != sql.ErrNoRows {
			return Token{}, err
		}
		return Token{}, nil
	}
	return token, nil
}

func (s *StorageRdbImpl) GetTokensByTerms(terms []string) ([]Token, error) {
	if len(terms) == 0 {
		return []Token{}, nil
	}

	query, args, err := sqlx.In(`select * from tokens where term in (?) order by field (term, ?)`, terms, terms)
	if err != nil {
		return nil, err
	}

	var tokens []Token
	if err := s.DB.Select(&tokens, query, args...); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *StorageRdbImpl) GetInvertedIndexByTokenIDs(ids []TokenID) (InvertedIndex, error) {
	if len(ids) == 0 {
		return InvertedIndex{}, nil
	}
	var encoded []EncodedInvertedIndex

	query, args, err := sqlx.In(
		`select
       		token_id,
			posting_list
		from
			inverted_indexes
		where
			token_id in (?)`, ids)
	if err != nil {
		return nil, err
	}
	if err = s.DB.Select(&encoded, query, args...); err != nil {
		return nil, err
	}
	return decode(encoded)
}

func (s *StorageRdbImpl) UpsertInvertedIndex(inverted InvertedIndex) error {
	encoded, err := encode(inverted)
	if err != nil {
		return err
	}

	for _, v := range encoded {
		_, err := s.DB.NamedExec(
			`insert into inverted_indexes (token_id, posting_list)
			values (:token_id, :posting_list)
			on duplicate key update posting_list = :posting_list`, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func encode(invertedIndex InvertedIndex) ([]EncodedInvertedIndex, error) {
	encoded := make([]EncodedInvertedIndex, 0)
	for k, v := range invertedIndex {
		// 差分を取る
		var p *Postings = v.Postings
		var beforeDocumentID DocumentID = 0
		for p != nil {
			p.DocumentID -= beforeDocumentID
			beforeDocumentID = p.DocumentID + beforeDocumentID
			p = p.Next
		}

		// Gobでシリアライズ&圧縮
		plBuf := bytes.NewBuffer(nil)
		if err := gob.NewEncoder(plBuf).Encode(v.Postings); err != nil {
			return nil, err
		}
		encoded = append(encoded, NewEncodedInvertedIndex(k, plBuf.Bytes()))
	}
	return encoded, nil
}

type EncodedInvertedIndex struct {
	TokenID     TokenID `db:"token_id"`     // トークンID
	PostingList []byte  `db:"posting_list"` // トークンを含むポスティングスリスト
}

func NewEncodedInvertedIndex(id TokenID, pl []byte) EncodedInvertedIndex {
	return EncodedInvertedIndex{
		TokenID:     id,
		PostingList: pl,
	}
}

func decode(e []EncodedInvertedIndex) (InvertedIndex, error) {
	m := make(map[TokenID]PostingList)
	for _, encoded := range e {
		// Gobでデシリアライズ
		p := &Postings{}
		ret := bytes.NewBuffer(encoded.PostingList)
		if err := gob.NewDecoder(ret).Decode(p); err != nil {
			return nil, err
		}
		pl := NewPostingList(p)

		// 差分から本来のIx｀Dへ変換
		var c *Postings = pl.Postings
		var beforeDocumentID DocumentID = 0
		for c != nil {
			c.DocumentID += beforeDocumentID
			beforeDocumentID = c.DocumentID
			c = c.Next
		}
		m[encoded.TokenID] = pl
	}
	return m, nil
}
