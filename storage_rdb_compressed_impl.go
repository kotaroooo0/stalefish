package stalefish

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// ストレージで圧縮した転置リストを扱う実装
type StorageRdbCompressedImpl struct {
	StorageRdbImpl
}

func NewStorageRdbCompressedImpl(db *sqlx.DB) StorageRdbCompressedImpl {
	return StorageRdbCompressedImpl{
		StorageRdbImpl: StorageRdbImpl{
			DB: db,
		},
	}
}

func (s StorageRdbCompressedImpl) UpsertInvertedIndex(invertedIndexValue InvertedIndexValue) error {
	// TODO: 実装
	return nil
}

func (s StorageRdbCompressedImpl) GetInvertedIndexByTokenID(tokenID TokenID) (InvertedIndexValue, error) {
	// TODO: 実装
	return InvertedIndexValue{}, nil
}
