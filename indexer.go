package stalefish

type Indexer struct {
	Storage       Storage       // 永続化層
	Analyzer      Analyzer      // 文章分割のためのアナライザ
	InvertedIndex InvertedIndex // 転置インデックス(メモリ上)
}

func NewIndexer(storage Storage, analyzer *Analyzer) *Indexer {
	return &Indexer{
		Storage:       storage,
		Analyzer:      *analyzer,
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

	// ストレージ上の転置インデックスにマージする
	if len(i.InvertedIndex) >= INDEX_SIZE_THRESHOLD {
		for tokenID, invertedIndexValue := range i.InvertedIndex {
			// マージ元の転置リストをストレージから読み出す
			storageInvertIndexValue, err := i.Storage.GetInvertedIndexByTokenID(tokenID)
			if err != nil {
				return err
			}

			if storageInvertIndexValue.PostingList == nil { // ストレージのポスティングリストが空の時
				// TODO: DB接続回数が減るので、ループ後にまとめて追加する方が良い
				i.Storage.UpsertInvertedIndex(tokenID, invertedIndexValue)
			} else {
				// ストレージ上の転置リストとメモリの転置リストをマージする
				merged, err := invertedIndexValue.Merge(storageInvertIndexValue)
				if err != nil {
					return err
				}
				// TODO: DB接続回数が減るので、ループ後にまとめて追加する方が良い
				// マージした転置リストをストレージに永続化する
				i.Storage.UpsertInvertedIndex(tokenID, merged)
			}
		}

		// メモリの転置インデックスをリセット
		i.InvertedIndex = InvertedIndex{}
	}
	return nil
}

// 文書からメモリ上の転置インデックスを更新する
func (i *Indexer) updateMemoryInvertedIndexByDocument(doc Document) error {
	tokens := i.Analyzer.Analyze(doc.Body)
	for pos, token := range tokens.Tokens {
		if err := i.updateMemoryInvertedIndexByToken(doc.ID, token, uint64(pos)); err != nil {
			return err
		}
	}
	return nil
}

// トークンからメモリ上の転置インデックスを更新する
func (i *Indexer) updateMemoryInvertedIndexByToken(docID DocumentID, term Token, pos uint64) error {
	// ストレージにIDの管理を任せる
	i.Storage.AddToken(NewToken(term.Term))
	token, err := i.Storage.GetTokenByTerm(term.Term)
	if err != nil {
		return err
	}

	invertedIndexValue, ok := i.InvertedIndex[token.ID]
	if !ok { // メモリ上に対応するinvertedIndexValueがない
		i.InvertedIndex[token.ID] = InvertedIndexValue{
			PostingList:    NewPostings(docID, []uint64{pos}, 1, nil),
			DocsCount:      1,
			PositionsCount: 1,
		}
		return nil
	}

	// ドキュメントに対応するポスティングが存在するかどうか
	// p == nilになる前にループ終了: 存在する
	// p == nilまでループが回る: 存在しない
	var p *Postings = invertedIndexValue.PostingList
	for p != nil && p.DocumentID != docID {
		p = p.Next
	}

	if p != nil { // 既に対象ドキュメントのポスティングが存在する
		p.Positions = append(p.Positions, pos)
		p.PositionsCount++

		invertedIndexValue.PositionsCount++
		i.InvertedIndex[token.ID] = invertedIndexValue
	} else { // まだ対象ドキュメントのポスティングが存在しない
		if docID < invertedIndexValue.PostingList.DocumentID { // 追加されるポスティングのドキュメントIDが最小の時
			invertedIndexValue.PostingList = NewPostings(docID, []uint64{pos}, 1, invertedIndexValue.PostingList)
		} else { // 追加されるポスティングのドキュメントIDが最小でない時
			// ドキュメントIDが昇順になるように挿入する場所を探索
			var t *Postings = invertedIndexValue.PostingList
			for t.Next != nil && t.Next.DocumentID < docID {
				t = t.Next
			}
			t.push(NewPostings(docID, []uint64{pos}, 1, nil))
		}
		invertedIndexValue.DocsCount++
		invertedIndexValue.PositionsCount++
		i.InvertedIndex[token.ID] = invertedIndexValue
	}
	return nil
}
