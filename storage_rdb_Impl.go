package stalefish

import (
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type StorageRdbImpl struct {
	DB *sqlx.DB
}

func NewStorageRdbImpl(db *sqlx.DB) StorageRdbImpl {
	return StorageRdbImpl{
		DB: db,
	}
}

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

func NewDBConfig(user, password, addr, port, db string) *DBConfig {
	return &DBConfig{
		User:     user,
		Password: password,
		Addr:     addr,
		Port:     port,
		DB:       db,
	}
}

type DBConfig struct {
	User     string
	Password string
	Addr     string
	Port     string
	DB       string
}

func (s StorageRdbImpl) GetAllDocuments() ([]Document, error) {
	var docs []Document
	err := s.DB.Select(&docs, `select * from documents`)
	return docs, err
}

func (s StorageRdbImpl) GetDocuments(ids []DocumentID) ([]Document, error) {
	intDocIDs := make([]int, len(ids))
	for i, id := range ids {
		intDocIDs[i] = int(id)
	}

	sql, params, err := sqlx.In(`select * from documents where id in (?)`, intDocIDs)
	if err != nil {
		return nil, err
	}
	var docs []Document
	err = s.DB.Select(&docs, sql, params...)
	return docs, err
}

func (s StorageRdbImpl) AddDocument(doc Document) (DocumentID, error) {
	res, err := s.DB.NamedExec(`insert into documents (body) values (:body)`,
		map[string]interface{}{
			"body": doc.Body,
		})
	if err != nil {
		return -1, err
	}

	insertedID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return DocumentID(insertedID), err
}

// TODO: 同じトークンでもIDがインクリメントされ、IDがとびとびになる
func (s StorageRdbImpl) AddToken(token Token) (TokenID, error) {
	res, err := s.DB.NamedExec(`insert into tokens (term) values (:term)`,
		map[string]interface{}{
			"term": token.Term,
		})
	if err != nil {
		return -1, err
	}
	insertedID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return TokenID(insertedID), err
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
		return Token{}, fmt.Errorf("error: two or more hits(inconsistent match result)")
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
	return err
}

func (s StorageRdbImpl) GetInvertedIndexByTokenID(tokenID TokenID) (InvertedIndexValue, error) {
	// TODO: LEFT JOINのがいいかも(?)
	var invertedIndexDtos []invertedIndexDto
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
		where token_id = ?`, int(tokenID))
	if err != nil {
		return InvertedIndexValue{}, err
	}

	switch len(invertedIndexDtos) {
	case 0:
		return InvertedIndexValue{}, nil
	case 1:
		return dtoToInvertedIndexValue(invertedIndexDtos[0]), nil
	default:
		return InvertedIndexValue{}, fmt.Errorf("error: two or more hits(inconsistent match result)")
	}
}

func postingsToJson(p *postings) ([]byte, error) {
	list := make([]posting, 0)
	for p != nil {
		list = append(list, newPosting(p.documentId, p.positions, p.positionsCount))
		p = p.next
	}
	return json.Marshal(list)
}

func listToPostings(list []posting) *postings {
	var p *postings = newPostings(list[0].documentId, list[0].positions, list[0].positionCount, nil)
	var root *postings = p
	for i, l := range list {
		if i == 0 {
			continue
		}
		p.next = newPostings(l.documentId, l.positions, l.positionCount, nil)
		p = p.next
	}
	return root
}

func dtoToInvertedIndexValue(dto invertedIndexDto) InvertedIndexValue {
	return InvertedIndexValue{
		Token:          dto.token,
		PostingList:    listToPostings(dto.postingList),
		DocsCount:      dto.docsCount,
		PositionsCount: dto.positionsCount,
	}
}

// 転置リスト
type invertedIndexDto struct {
	token          Token     `db:"token"`
	postingList    []posting `db:"posting_list"`    // トークンを含むポスティングスリスト
	docsCount      int       `db:"docs_count"`      // トークンを含む文書数
	positionsCount int       `db:"positions_count"` // 全文書内でのトークンの出現数
}

type posting struct {
	documentId    DocumentID
	positions     []int
	positionCount int
}

func newPosting(documentId DocumentID, positions []int, positionCount int) posting {
	return posting{
		documentId:    documentId,
		positions:     positions,
		positionCount: positionCount,
	}
}
