package stalefish

// 転置インデックス
// TokenIDー>転置リストのマップ
type InvertedIndexMap map[TokenID]InvertedIndexValue

type TokenID int

type Token struct {
	ID   TokenID `db:"id"`
	Term string  `db:"term"`
	Kana string  `db:"kana"`
}

func NewToken(term string) Token {
	return Token{
		Term: term,
	}
}

// 転置リスト
type InvertedIndexValue struct {
	Token          Token       `db:"token"`
	PostingList    PostingList `db:"posting_list"`    // トークンを含むポスティングスリスト
	DocsCount      int         `db:"docs_count"`      // トークンを含む文書数
	PositionsCount int         `db:"positions_count"` // 全文書内でのトークンの出現数
}

// 転置リストのスライス
type InvertedIndexValues []InvertedIndexValue

// ポスティングリスト。文書IDのリンクリスト
type PostingList []Posting

type Posting struct {
	DocumentID     DocumentID // 文書のID
	Positions      []int      // 文書中の位置情報
	PositionsCount int        // 文書中の位置情報の数
}
