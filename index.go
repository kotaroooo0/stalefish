package stalefish

// 転置インデックス
// TokenIDー>転置リストのマップ
type InvertedIndex map[TokenID]PostingList

// 転置リスト
type PostingList struct {
	Postings       *Postings // トークンを含むポスティングスリスト
	DocsCount      uint64    // トークンを含む文書数
	PositionsCount uint64    // 全文書内でのトークンの出現数
}

func NewPostingList(pl *Postings, docsCount, positionsCount uint64) PostingList {
	return PostingList{
		Postings:       pl,
		DocsCount:      docsCount,
		PositionsCount: positionsCount,
	}
}

func (i PostingList) Merge(target PostingList) (PostingList, error) {
	merged := PostingList{
		Postings:       nil,
		PositionsCount: 0,
		DocsCount:      0,
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

	for c := merged.Postings; c != nil; c = c.Next {
		merged.DocsCount += 1
		merged.PositionsCount += c.PositionsCount
	}
	return merged, nil
}

// ポスティング(文書IDのリンクリスト)
type Postings struct {
	DocumentID     DocumentID // 文書のID
	Positions      []uint64   // 文書中の位置情報
	PositionsCount uint64     // 文書中の位置情報の数
	Next           *Postings  // 次のポスティングへのポインタ
}

func NewPostings(documentID DocumentID, positions []uint64, positionsCount uint64, next *Postings) *Postings {
	return &Postings{
		DocumentID:     documentID,
		Positions:      positions,
		PositionsCount: positionsCount,
		Next:           next,
	}
}

func (p *Postings) Push(e *Postings) {
	e.Next = p.Next
	p.Next = e
}
