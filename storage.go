package stalefish

type Storage interface {
	GetAllDocuments() ([]Document, error)                        // 全てのドキュメントを返す
	GetDocuments([]DocumentID) ([]Document, error)               // 複数IDから複数ドキュメントを返す
	AddDocument(Document) (DocumentID, error)                    // ドキュメントを挿入する。挿入したドキュメントのIDを返す。
	AddToken(Token) (TokenID, error)                             // トークンを挿入する。挿入したトークンのIDを返す。
	GetTokenByTerm(string) (Token, error)                        // 語句からトークンを取得する
	GetTokensByTerms([]string) ([]Token, error)                  // 複数の語句から複数トークンを取得する
	GetInvertedIndexByTokenIDs([]TokenID) (InvertedIndex, error) // 複数トークンIDから転置インデックスを取得する
	UpsertInvertedIndex(InvertedIndex) error                     // 転置リストを更新する
}
