package stalefish

type DocumentID uint

// TODO: リフレクションを利用して任意のフィールドをもつ構造体として扱いたい
type Document struct {
	ID   DocumentID
	Body string `db:"body"`
}

func NewDocument(body string) Document {
	return Document{
		Body: body,
	}
}
