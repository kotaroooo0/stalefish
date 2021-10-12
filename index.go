package stalefish

import "sort"

// 転置インデックス
// TokenIDー>転置リストのマップ
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

// 転置リスト
type PostingList struct {
	Postings *Postings // トークンを含むポスティングスリスト
}

func NewPostingList(pl *Postings) PostingList {
	return PostingList{
		Postings: pl,
	}
}

func (p PostingList) Size() int {
	size := 0
	ps := p.Postings
	for ps != nil {
		ps = ps.Next
		size++
	}
	return size
}

func (p PostingList) AppearanceCountInDocument(docID DocumentID) int {
	count := 0
	ps := p.Postings
	for ps != nil {
		if ps.DocumentID == docID {
			count = len(ps.Positions)
			break
		}
		ps = ps.Next
	}
	return count
}

// ポスティング(文書IDのリンクリスト)
type Postings struct {
	DocumentID DocumentID // 文書のID
	Positions  []uint64   // 文書中の位置情報
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
