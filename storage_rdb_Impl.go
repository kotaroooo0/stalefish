package stalefish

import (
	"encoding/json"
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

func (s StorageRdbImpl) UpsertInvertedIndex(invertedIndexValue InvertedIndexValue) error {
	postingsJson, err := postingsToJson(invertedIndexValue.PostingList)
	if err != nil {
		return err
	}

	_, err = s.DB.NamedExec(
		`insert into inverted_indexes (token_id, posting_list, docs_count, positions_count)
		values (:token_id, :posting_list, :docs_count, :positions_count)
		on duplicate key update posting_list = :posting_list,docs_count = :docs_count,positions_count = :positions_count`,
		map[string]interface{}{
			"token_id":        invertedIndexValue.Token.ID,
			"posting_list":    postingsJson,
			"docs_count":      invertedIndexValue.DocsCount,
			"positions_count": invertedIndexValue.PositionsCount,
		},
	)
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (s StorageRdbImpl) GetInvertedIndexByTokenID(tokenID TokenID) (InvertedIndexValue, error) {
	// TODO: LEFT JOINのがいいかも(?)
	var invertedIndexDtos []InvertedIndexDto
	err := s.DB.Select(&invertedIndexDtos,
		`select
			tokens.id as "token.id",
			tokens.term as "token.term",
			inverted_indexes.posting_list as posting_list,
			inverted_indexes.docs_count as docs_count,
			inverted_indexes.positions_count as positions_count
		from
			inverted_indexes
		join
			tokens on inverted_indexes.token_id = tokens.id
		where
			token_id = ?`, int(tokenID))
	if err != nil {
		return InvertedIndexValue{}, errors.New(err.Error())
	}

	switch len(invertedIndexDtos) {
	case 0:
		return InvertedIndexValue{}, nil
	case 1:
		return dtoToInvertedIndexValue(invertedIndexDtos[0]), nil
	default:
		return InvertedIndexValue{}, errors.New("error: two or more hits(inconsistent match result)")
	}
}

func postingsToJson(p *Postings) ([]byte, error) {
	list := make([]Posting, 0)
	for p != nil {
		list = append(list, NewPosting(p.DocumentID, p.Positions, p.PositionsCount))
		p = p.Next
	}
	return json.Marshal(list)
}

func listToPostings(list []Posting) *Postings {
	var p *Postings = NewPostings(list[0].DocumentID, list[0].Positions, list[0].PositionCount, nil)
	var root *Postings = p
	for i, l := range list {
		if i == 0 {
			continue
		}
		p.Next = NewPostings(l.DocumentID, l.Positions, l.PositionCount, nil)
		p = p.Next
	}
	return root
}

func dtoToInvertedIndexValue(dto InvertedIndexDto) InvertedIndexValue {
	return InvertedIndexValue{
		Token:          dto.Token,
		PostingList:    listToPostings(dto.PostingList),
		DocsCount:      dto.DocsCount,
		PositionsCount: dto.PositionsCount,
	}
}

// 転置リスト
type InvertedIndexDto struct {
	Token          Token       `db:"token"`
	PostingList    PostingList `db:"posting_list"`    // トークンを含むポスティングスリスト
	DocsCount      uint64      `db:"docs_count"`      // トークンを含む文書数
	PositionsCount uint64      `db:"positions_count"` // 全文書内でのトークンの出現数
}

type Posting struct {
	DocumentID    DocumentID
	Positions     []uint64
	PositionCount uint64
}

type PostingList []Posting

func NewPosting(DocumentID DocumentID, positions []uint64, positionCount uint64) Posting {
	return Posting{
		DocumentID:    DocumentID,
		Positions:     positions,
		PositionCount: positionCount,
	}
}

func (pl *PostingList) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &pl)
		return nil
	case string:
		json.Unmarshal([]byte(v), &pl)
		return nil
	default:
		return errors.New(fmt.Sprintf("unsupported type: %T", v))
	}
}
