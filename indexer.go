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

// 0: 常にストレージへ保存
const INDEX_SIZE_THRESHOLD = 0

// 1.文書からトークンを取り出す
// 2.トークンごとにポスティングリストを作って、それをメモリ上の転置インデックスに追加する
// 3.メモリ上の転置インデックスがある程度のサイズになったら、ストレージ上の転置インデックスにマージする
func (i *Indexer) AddDocument(doc Document) error {
	// ストレージに文書を格納しIDを取得
	docID, err := i.Storage.AddDocument(doc)
	if err != nil {
		return err
	}
	doc.ID = docID

	// 文書から転置リストを構築
	if err := i.updateMemoryInvertedIndexByDocument(doc); err != nil {
		return err
	}

	ids := make([]TokenID, len(i.InvertedIndex))
	for i := range i.InvertedIndex {
		ids = append(ids, i)
	}

	if len(i.InvertedIndex) < INDEX_SIZE_THRESHOLD {
		return nil
	}

	// マージ元の転置リストをストレージから読み出す
	storageInvertedIndex, err := i.Storage.GetInvertedIndexByTokenIDs(ids)
	if err != nil {
		return err
	}

	// ストレージ上の転置インデックスにマージする
	for tokenID, postingList := range i.InvertedIndex {
		storagePostingList := storageInvertedIndex[tokenID]
		i.InvertedIndex[tokenID] = postingList.Merge(storagePostingList)
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
func (i *Indexer) updateMemoryPostingListByToken(docID DocumentID, term Token, pos uint64) error {
	// ストレージにIDの管理を任せる
	if _, err := i.Storage.AddToken(NewToken(term.Term)); err != nil {
		return err
	}
	token, err := i.Storage.GetTokenByTerm(term.Term)
	if err != nil {
		return err
	}

	postingList, ok := i.InvertedIndex[token.ID]
	if !ok { // メモリ上に対応するポスティングリストがない
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

	if p != nil { // 既に対象ドキュメントのポスティングが存在する
		p.Positions = append(p.Positions, pos)
		i.InvertedIndex[token.ID] = postingList
	} else { // まだ対象ドキュメントのポスティングが存在しない
		if docID < postingList.Postings.DocumentID { // 追加されるポスティングのドキュメントIDが最小の時
			postingList.Postings = NewPostings(docID, []uint64{pos}, postingList.Postings)
		} else { // 追加されるポスティングのドキュメントIDが最小でない時
			// ドキュメントIDが昇順になるように挿入する場所を探索
			var t *Postings = postingList.Postings
			for t.Next != nil && t.Next.DocumentID < docID {
				t = t.Next
			}
			t.Push(NewPostings(docID, []uint64{pos}, nil))
		}
		i.InvertedIndex[token.ID] = postingList
	}
	return nil
}
