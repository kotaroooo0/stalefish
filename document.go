package stalefish

type DocumentID uint64

type Document struct {
	ID         DocumentID
	Body       string `db:"body"`
	TokenCount int    `db:"token_count"`
}

func NewDocument(body string) Document {
	return Document{
		Body: body,
	}
}
