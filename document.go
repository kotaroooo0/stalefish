package stalefish

type DocumentID uint64

// TODO: リフレクションを利用して任意のフィールドをもつ構造体として扱いたい
type Document struct {
	ID         DocumentID
	Body       string `db:"body"`
	TokenCount int    `db:"token_count"`
}

func NewDocument(body string, tokenCount int) Document {
	return Document{
		Body:       body,
		TokenCount: tokenCount,
	}
}
