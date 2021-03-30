package stalefish

// 転置インデックス
// TokenIDー>転置リストのマップ
type InvertedIndexMap map[TokenID]InvertedIndexValue

type TokenID int

type Kind int

const (
	Term   Kind = iota // トークンのオリジナル
	Kana               // トークンのカナ
	Romaji             // トークンのローマ字
)

type Token struct {
	ID     TokenID `db:"id"`
	Term   string  `db:"term"`
	Kana   string  `db:"kana"`
	Romaji string  `db:"romaji"`
}

type TokenOption func(*Token)

func NewToken(term string, options ...TokenOption) Token {
	token := Token{Term: term}
	for _, option := range options {
		option(&token)
	}
	return token
}

func SetKana(kana string) TokenOption {
	return func(s *Token) {
		s.Kana = kana
	}
}

func SetRomaji(romaji string) TokenOption {
	return func(s *Token) {
		s.Romaji = romaji
	}
}

type TokenStream struct {
	Tokens   []Token
	Selected Kind
}

func NewTokenStream(tokens []Token, selected Kind) *TokenStream {
	return &TokenStream{
		Tokens:   tokens,
		Selected: selected,
	}
}

func (ts *TokenStream) size() int {
	return len(ts.Tokens)
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
