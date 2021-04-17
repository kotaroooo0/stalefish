package stalefish

type Storage interface {
	GetAllDocuments() ([]Document, error)
	GetDocuments([]DocumentID) ([]Document, error)                 // 複数IDから複数ドキュメントを返す
	AddDocument(Document) (DocumentID, error)                      // ドキュメントを挿入する。挿入したドキュメントのIDを返す。
	AddToken(Token) (TokenID, error)                               // トークンを挿入する。挿入したトークンのIDを返す。
	GetTokenByTerm(string) (Token, error)                          // 語句からトークンを取得する
	GetInvertedIndexByTokenID(TokenID) (InvertedIndexValue, error) // トークンIDから転置リストを取得する
	UpsertInvertedIndex(TokenID, InvertedIndexValue) error         // 転置リストを更新する
}
