package stalefish

type Storage interface {
	GetAllDocuments() ([]Document, error)                          // 全てのドキュメントを返す
	GetDocuments([]DocumentID) ([]Document, error)                 // 複数IDから複数ドキュメントを返す
	AddDocument(Document) (DocumentID, error)                      // ドキュメントを挿入する。挿入したドキュメントのIDを返す。
	AddToken(Token) (TokenID, error)                               // トークンを挿入する。挿入したトークンのIDを返す。
	GetTokenByTerm(string) (Token, error)                          // 語句からトークンを取得する
	GetTokensByTerms([]string) ([]Token, error)                    // 複数の語句から複数トークンを取得する
	GetInvertedIndexByTokenID(TokenID) (PostingList, error)        // トークンIDからポスティングリストを取得する
	GetInvertedIndexesByTokenIDs([]TokenID) ([]PostingList, error) // 複数トークンIDから複数のポスティングリストを取得する
	UpsertInvertedIndex(TokenID, PostingList) error                // 転置リストを更新する
}
