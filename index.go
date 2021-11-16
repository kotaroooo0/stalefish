package stalefish

import "sort"

// 転置インデックス
// TokenIDー>ポスティングリストのマップ
type InvertedIndex map[TokenID]PostingList

func NewInvertedIndex(m map[TokenID]PostingList) InvertedIndex {
	return InvertedIndex(m)
}

func (ii InvertedIndex) TokenIDs() []TokenID {
	ids := []TokenID{}
	for i := range ii {
		ids = append(ids, i)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// ポスティングリスト
type PostingList struct {
	Postings *Postings // トークンごとのポスティングリスト
}

func NewPostingList(pl *Postings) PostingList {
	return PostingList{
		Postings: pl,
	}
}

func (p PostingList) Size() int {
	size := 0
	pp := p.Postings
	for pp != nil {
		pp = pp.Next
		size++
	}
	return size
}

func (p PostingList) AppearanceCountInDocument(docID DocumentID) int {
	count := 0
	pp := p.Postings
	for pp != nil {
		if pp.DocumentID == docID {
			count = len(pp.Positions)
			break
		}
		pp = pp.Next
	}
	return count
}

// ポスティング（ドキュメントID等を含むリンクリスト）
type Postings struct {
	DocumentID DocumentID // ドキュメントのID
	Positions  []uint64   // ドキュメント中での位置情報
	Next       *Postings  // 次のポスティングへのポインタ
}

func NewPostings(documentID DocumentID, positions []uint64, next *Postings) *Postings {
	return &Postings{
		DocumentID: documentID,
		Positions:  positions,
		Next:       next,
	}
}

func (p *Postings) PushBack(e *Postings) {
	e.Next = p.Next
	p.Next = e
}
