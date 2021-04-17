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
	PostingList    *Postings // トークンを含むポスティングスリスト
	DocsCount      uint64    // トークンを含む文書数
	PositionsCount uint64    // 全文書内でのトークンの出現数
}

func NewInvertedIndexValue(pl *Postings, docsCount, positionsCount uint64) InvertedIndexValue {
	return InvertedIndexValue{
		PostingList:    pl,
		DocsCount:      docsCount,
		PositionsCount: positionsCount,
	}
}

func (i InvertedIndexValue) Merge(target InvertedIndexValue) (InvertedIndexValue, error) {
	merged := InvertedIndexValue{
		PostingList:    nil,
		PositionsCount: 0,
		DocsCount:      0,
	}

	var smaller, larger *Postings
	if i.PostingList.DocumentID <= target.PostingList.DocumentID {
		merged.PostingList = i.PostingList
		smaller, larger = i.PostingList, target.PostingList
	} else {
		merged.PostingList = target.PostingList
		smaller, larger = target.PostingList, i.PostingList
	}

	for larger != nil {
		if smaller.Next == nil {
			smaller.Next = larger
			break
		}

		if smaller.Next.DocumentID < larger.DocumentID {
			smaller = smaller.Next
		} else if smaller.Next.DocumentID > larger.DocumentID {
			largerNext, smallerNext := larger.Next, smaller.Next
			smaller.Next, larger.Next = larger, smallerNext
			smaller = larger
			larger = largerNext
		} else if smaller.Next.DocumentID == larger.DocumentID {
			smaller, larger = smaller.Next, larger.Next
		}
	}

	for c := merged.PostingList; c != nil; c = c.Next {
		merged.DocsCount += 1
		merged.PositionsCount += c.PositionsCount
	}
	return merged, nil
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
