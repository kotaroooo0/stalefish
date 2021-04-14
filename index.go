package stalefish

// 転置インデックス
// TokenIDー>転置リストのマップ
type InvertedIndex map[TokenID]InvertedIndexValue

type TokenID uint64

type Token struct {
	ID   TokenID `db:"id"`
	Term string  `db:"term"`
	Kana string  `db:"kana"`
}

type TokenOption func(*Token)

func NewToken(term string, options ...TokenOption) Token {
	token := Token{Term: term}
	for _, option := range options {
		option(&token)
	}
	return token
}

func setKana(kana string) TokenOption {
	return func(s *Token) {
		s.Kana = kana
	}
}

type TokenStream struct {
	Tokens []Token
}

func NewTokenStream(tokens []Token) *TokenStream {
	return &TokenStream{
		Tokens: tokens,
	}
}

func (ts *TokenStream) size() int {
	return len(ts.Tokens)
}

// 転置リスト
type InvertedIndexValue struct {
	Token          Token
	PostingList    *Postings // トークンを含むポスティングスリスト
	DocsCount      uint64    // トークンを含む文書数
	PositionsCount uint64    // 全文書内でのトークンの出現数
}

func NewInvertedIndexValue(token Token, pl *Postings, docsCount, positionsCount uint64) InvertedIndexValue {
	return InvertedIndexValue{
		Token:          token,
		PostingList:    pl,
		DocsCount:      docsCount,
		PositionsCount: positionsCount,
	}
}

// 転置リストのスライス
type InvertedIndexValues []InvertedIndexValue

// ポスティングリスト。文書IDのリンクリスト
type Postings struct {
	DocumentID     DocumentID // 文書のID
	Positions      []uint64   // 文書中の位置情報
	PositionsCount uint64     // 文書中の位置情報の数
	Next           *Postings
}

func NewPostings(documentID DocumentID, positions []uint64, positionsCount uint64, next *Postings) *Postings {
	return &Postings{
		DocumentID:     documentID,
		Positions:      positions,
		PositionsCount: positionsCount,
		Next:           next,
	}
}

func (p *Postings) push(e *Postings) {
	e.Next = p.Next
	p.Next = e
}
