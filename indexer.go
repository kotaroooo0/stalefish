package stalefish

import (
	"fmt"
)

type Indexer struct {
	Storage          Storage
	Analyzer         Analyzer
	InvertedIndexMap InvertedIndexMap
}

func NewIndexer(storage Storage, analyzer Analyzer, invertedIndexMap InvertedIndexMap) *Indexer {
	return &Indexer{
		Storage:          storage,
		Analyzer:         analyzer,
		InvertedIndexMap: invertedIndexMap,
	}
}

const INDEX_SIZE_THRESHOLD = 10

// TODO: Goと言えばgoroutine, 並列化をしていきたい
// TODO: AddDocumentsも作りたい
// 1.文書からトークンを取り出す
// 2.トークンごとにポスティングリストを作って、それをメモリ上の転置インデックスに追加する
// 3.メモリ上の転置インデックスがある程度のサイズになったら、ストレージ上の転置インデックスにマージする
func (i *Indexer) AddDocument(doc Document) error {
	// ストレージに文書を格納
	docID, err := i.Storage.AddDocument(doc)
	if err != nil {
		return err
	}
	doc.ID = docID

	// 文書から転置リストを構築
	i.UpdateMemoryInvertedIndexByDocument(doc)

	// ストレージ上の転置インデックスにマージする
	if len(i.InvertedIndexMap) >= INDEX_SIZE_THRESHOLD {
		for tokenID, invertedIndexValue := range i.InvertedIndexMap {
			// マージ元の転置リストをストレージから読み出す
			storageInvertIndexValue, err := i.Storage.GetInvertedIndexByTokenID(tokenID)
			if err != nil {
				return err
			}

			if storageInvertIndexValue.PostingList == nil { // ストレージのポスティングリストが空の時
				// TODO: DB接続回数が減るので、ループ後にまとめて追加する方が良い
				i.Storage.UpsertInvertedIndex(invertedIndexValue)
			} else {
				// ストレージ上の転置リストとメモリの転置リストをマージする
				merged, err := merge(invertedIndexValue, storageInvertIndexValue)
				if err != nil {
					return err
				}
				// TODO: DB接続回数が減るので、ループ後にまとめて追加する方が良い
				// マージした転置リストをストレージに永続化する
				i.Storage.UpsertInvertedIndex(merged)
			}
		}

		// メモリの転置インデックスをリセット
		i.InvertedIndexMap = InvertedIndexMap{}
	}
	return nil
}

// 文書からメモリ上の転置インデックスを更新する
func (i *Indexer) UpdateMemoryInvertedIndexByDocument(doc Document) error {
	tokens := i.Analyzer.Analyze(doc.Body)
	for pos, token := range tokens.Tokens {
		if err := i.UpdateMemoryInvertedIndexByToken(doc.ID, token, pos); err != nil {
			return err
		}
	}
	return nil
}

// トークンからメモリ上の転置インデックスを更新する
func (i *Indexer) UpdateMemoryInvertedIndexByToken(docID DocumentID, term Token, pos int) error {
	// ストレージにIDの管理を任せる
	i.Storage.AddToken(NewToken(term.Term))

	token, err := i.Storage.GetTokenByTerm(term.Term)
	if err != nil {
		return err
	}

	invertedIndexValue, ok := i.InvertedIndexMap[token.ID]
	if !ok { // 対応するinvertedIndexValueがない場合
		i.InvertedIndexMap[token.ID] = InvertedIndexValue{
			Token:          token,
			PostingList:    newPostings(docID, []int{pos}, 1, nil),
			DocsCount:      1,
			PositionsCount: 1,
		}
		return nil
	}

	// ドキュメントに対応するポスティングが存在するかどうか
	// p == nilになる前にループ終了: 存在する
	// p == nilまでループが回る: 存在しない
	p := invertedIndexValue.PostingList
	for p.documentId != docID && p != nil {
		p = p.next
	}

	if p != nil { // 既に対象ドキュメントのポスティングが存在する
		p.positions = append(p.positions, pos)
		p.positionsCount++

		invertedIndexValue.PositionsCount++
		i.InvertedIndexMap[token.ID] = invertedIndexValue
	} else { // まだ対象ドキュメントのポスティングが存在しない
		// ドキュメントIDが昇順になるように挿入する場所を探索
		t := invertedIndexValue.PostingList
		for t.next.documentId < docID {
			t = t.next
		}
		t.push(newPostings(docID, []int{pos}, 1, nil))

		invertedIndexValue.DocsCount++
		invertedIndexValue.PositionsCount++
		i.InvertedIndexMap[token.ID] = invertedIndexValue
	}
	return nil
}

// TODO: メソッドにした方がいいかも？
func merge(memory, storage InvertedIndexValue) (InvertedIndexValue, error) {
	// 同じトークンに対する転置リストでなければエラーを返す
	if memory.Token.ID != storage.Token.ID || memory.Token.Term != storage.Token.Term {
		return InvertedIndexValue{}, fmt.Errorf("error: not match inverted index")
	}

	merged := InvertedIndexValue{
		Token:          memory.Token,
		PostingList:    nil,
		PositionsCount: 0,
		DocsCount:      0,
	}

	var smaller, larger *postings
	if memory.PostingList.documentId <= storage.PostingList.documentId {
		merged.PostingList = memory.PostingList
		smaller = memory.PostingList
		larger = storage.PostingList
	} else {
		merged.PostingList = storage.PostingList
		smaller = storage.PostingList
		larger = memory.PostingList
	}

	for {
		merged.DocsCount++

		if smaller.next.documentId < larger.documentId {
			smaller = smaller.next
		}

		if smaller.next.documentId > larger.documentId {
			larNext := larger.next
			smaNext := smaller.next
			smaller.next = larger
			larger.next = smaNext
			smaller = larger
			larger = larNext
		}

		if smaller.next.documentId == larger.documentId {
			smaller = smaller.next
			larger = larger.next
		}

		if larger == nil {
			for smaller != nil {
				smaller = smaller.next
				merged.DocsCount++
			}
			break
		}
		if smaller.next == nil {
			smaller.next = larger
			for larger != nil {
				larger = larger.next
				merged.DocsCount++
			}
			break
		}
	}
	return merged, nil
}
