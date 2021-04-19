package stalefish

// 転置インデックス
// TokenIDー>転置リストのマップ
type InvertedIndex map[TokenID]PostingList

func NewInvertedIndex(m map[TokenID]PostingList) InvertedIndex {
	return InvertedIndex(m)
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

func (i PostingList) Merge(target PostingList) PostingList {
	if i.Postings == nil {
		return target
	}
	if target.Postings == nil {
		return i
	}

	merged := PostingList{
		Postings: nil,
	}
	var smaller, larger *Postings
	if i.Postings.DocumentID <= target.Postings.DocumentID {
		merged.Postings = i.Postings
		smaller, larger = i.Postings, target.Postings
	} else {
		merged.Postings = target.Postings
		smaller, larger = target.Postings, i.Postings
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
	return merged
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

func (p *Postings) Push(e *Postings) {
	e.Next = p.Next
	p.Next = e
}
