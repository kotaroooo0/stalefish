package stalefish

type Indexer struct {
	Storage       Storage       // 永続化層
	Analyzer      Analyzer      // 文章分割のためのアナライザ
	InvertedIndex InvertedIndex // 転置インデックス(メモリ上)
}

func NewIndexer(storage Storage, analyzer Analyzer) *Indexer {
	return &Indexer{
		Storage:       storage,
		Analyzer:      analyzer,
		InvertedIndex: make(InvertedIndex),
	}
}

const INDEX_SIZE_THRESHOLD = 0

// 1.ドキュメントからトークンを取り出す
// 2.トークンごとにポスティングリストを作って、それをメモリ上の転置インデックスに追加する
// 3.メモリ上の転置インデックスがある程度のサイズになったら、ストレージ上の転置インデックスにマージする
func (i *Indexer) AddDocument(doc Document) error {
	// ストレージにドキュメントを格納し、ドキュメントIDを取得
	docID, err := i.Storage.AddDocument(doc)
	if err != nil {
		return err
	}
	doc.ID = docID

	// ドキュメントからメモリ上の転置インデックスを更新
	if err := i.updateMemoryInvertedIndexByDocument(doc); err != nil {
		return err
	}

	// メモリ上の転置インデックスのサイズが閾値以下であれば、処理終了
	if len(i.InvertedIndex) < INDEX_SIZE_THRESHOLD {
		return nil
	}

	// マージ元の転置リストをストレージから読み出す
	storageInvertedIndex, err := i.Storage.GetInvertedIndexByTokenIDs(i.InvertedIndex.TokenIDs())
	if err != nil {
		return err
	}

	// メモリ上の転置インデックスとストレージ上の転置インデックスをマージする
	for tokenID, postingList := range i.InvertedIndex {
		i.InvertedIndex[tokenID] = merge(postingList, storageInvertedIndex[tokenID])
	}
	if err := i.Storage.UpsertInvertedIndex(i.InvertedIndex); err != nil {
		return err
	}

	// メモリの転置インデックスをリセット
	i.InvertedIndex = InvertedIndex{}
	return nil
}

// 文書からメモリ上の転置インデックスを更新する
func (i *Indexer) updateMemoryInvertedIndexByDocument(doc Document) error {
	tokens := i.Analyzer.Analyze(doc.Body)
	for pos, token := range tokens.Tokens {
		if err := i.updateMemoryPostingListByToken(doc.ID, token, uint64(pos)); err != nil {
			return err
		}
	}
	return nil
}

// トークンからメモリ上の転置インデックスを更新する
func (i *Indexer) updateMemoryPostingListByToken(docID DocumentID, token Token, pos uint64) error {
	// ストレージにIDの管理を任せる
	i.Storage.AddToken(NewToken(token.Term))
	token, err := i.Storage.GetTokenByTerm(token.Term)
	if err != nil {
		return err
	}

	postingList, ok := i.InvertedIndex[token.ID]
	// メモリ上にトークンに対応するポスティングリストがない時
	if !ok {
		i.InvertedIndex[token.ID] = PostingList{
			Postings: NewPostings(docID, []uint64{pos}, nil),
		}
		return nil
	}

	// ドキュメントに対応するポスティングが存在するかどうか
	// p == nilになる前にループ終了: 存在する
	// p == nilまでループが回る: 存在しない
	var p *Postings = postingList.Postings
	for p != nil && p.DocumentID != docID {
		p = p.Next
	}

	// 既に対象ドキュメントのポスティングが存在する時
	if p != nil {
		p.Positions = append(p.Positions, pos)
		i.InvertedIndex[token.ID] = postingList
		return nil
	}

	// まだ対象ドキュメントのポスティングが存在しない時
	// 1.追加されるポスティングのドキュメントIDが最小の時 or 2.追加されるポスティングのドキュメントIDが最小でない時
	// 1の時
	if docID < postingList.Postings.DocumentID {
		postingList.Postings = NewPostings(docID, []uint64{pos}, postingList.Postings)
		i.InvertedIndex[token.ID] = postingList
		return nil
	}
	// 2の時
	// ドキュメントIDが昇順になるように挿入する場所を探索
	var t *Postings = postingList.Postings
	for t.Next != nil && t.Next.DocumentID < docID {
		t = t.Next
	}
	t.PushBack(NewPostings(docID, []uint64{pos}, nil))
	i.InvertedIndex[token.ID] = postingList
	return nil
}

func merge(origin, target PostingList) PostingList {
	if origin.Postings == nil {
		return target
	}
	if target.Postings == nil {
		return origin
	}

	merged := PostingList{
		Postings: nil,
	}
	var smaller, larger *Postings
	if origin.Postings.DocumentID <= target.Postings.DocumentID {
		merged.Postings = origin.Postings
		smaller, larger = origin.Postings, target.Postings
	} else {
		merged.Postings = target.Postings
		smaller, larger = target.Postings, origin.Postings
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
