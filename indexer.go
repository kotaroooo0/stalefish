package stalefish

import (
	"fmt"
	"sort"
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

	// 転置リストの全てのポスティングリストをドキュメントIDの昇順でソート
	i.SortPostingList()

	// ストレージ上の転置インデックスにマージする
	if len(i.InvertedIndexMap) >= INDEX_SIZE_THRESHOLD {
		for tokenID, invertedIndexValue := range i.InvertedIndexMap {
			// マージ元の転置リストをストレージから読み出す
			storageInvertIndexValue, err := i.Storage.GetInvertedIndexByTokenID(tokenID)
			if err != nil {
				return err
			}

			if len(storageInvertIndexValue.PostingList) == 0 { // ストレージのポスティングリストが空の時
				// TODO: DB接続回数が減るので、ループ後にまとめて追加する方が良い
				i.Storage.UpsertInvertedIndex(invertedIndexValue)
			} else {
				// ストレージ上の転置リストとメモリの転置リストをマージする
				merged, err := MergeInvertedIndex(invertedIndexValue, storageInvertIndexValue)
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
	return i.UpdateMemoryInvertedIndexByText(doc.ID, doc.Body)
}

// Textからメモリ上の転置インデックスを更新する
func (i *Indexer) UpdateMemoryInvertedIndexByText(docID DocumentID, text string) error {
	tokens := i.Analyzer.Analyze(text)
	for pos, token := range tokens {
		if err := i.UpdateMemoryInvertedIndexByToken(docID, token, pos); err != nil {
			return err
		}
	}
	return nil
}

// トークンからメモリ上の転置インデックスを更新する
func (i *Indexer) UpdateMemoryInvertedIndexByToken(docID DocumentID, term string, pos int) error {
	// ストレージにIDの管理を任せる
	i.Storage.AddToken(NewToken(term))

	token, err := i.Storage.GetTokenByTerm(term)
	if err != nil {
		return err
	}
	invertedIndexValue, ok := i.InvertedIndexMap[token.ID]
	if !ok {
		i.InvertedIndexMap[token.ID] = InvertedIndexValue{
			Token: token,
			PostingList: []Posting{
				Posting{
					DocumentID:     docID,
					Positions:      []int{pos},
					PositionsCount: 1,
				},
			},
			DocsCount:      1,
			PositionsCount: 1,
		}
	} else {
		// TODO: メソッド化してもいいかも
		// 対象ドキュメントのポスティングが存在するかどうか
		// 存在するならば、そのポスティングのポスティングリスト上のインデックスを返す
		var targetPostingIdx int
		isExistPosting := false
		for i := 0; i < len(invertedIndexValue.PostingList); i++ {
			p := invertedIndexValue.PostingList[i]
			if p.DocumentID == docID {
				targetPostingIdx = i
				isExistPosting = true
				break
			}
		}

		if isExistPosting { // 既に対象ドキュメントのポスティングが存在する
			invertedIndexValue.PostingList[targetPostingIdx].Positions = append(invertedIndexValue.PostingList[targetPostingIdx].Positions, pos)
			invertedIndexValue.PostingList[targetPostingIdx].PositionsCount++
			invertedIndexValue.PositionsCount++
			i.InvertedIndexMap[token.ID] = invertedIndexValue
		} else { // まだ対象ドキュメントのポスティングが存在しない
			invertedIndexValue.PostingList = append(invertedIndexValue.PostingList,
				Posting{
					DocumentID:     docID,
					Positions:      []int{pos},
					PositionsCount: 1,
				})
			invertedIndexValue.DocsCount++
			invertedIndexValue.PositionsCount++
			i.InvertedIndexMap[token.ID] = invertedIndexValue
		}
	}
	return nil
}

func (i *Indexer) SortPostingList() {
	for _, v := range i.InvertedIndexMap {
		sort.Slice(v.PostingList, func(i, j int) bool {
			return v.PostingList[i].DocumentID < v.PostingList[j].DocumentID
		})
	}
}

// TODO: メソッドにした方がいいかも？
func MergeInvertedIndex(memoryInvertedIndex, storageInvertIndex InvertedIndexValue) (InvertedIndexValue, error) {
	// 同じトークンに対する転置リストでなければエラーを返す
	if memoryInvertedIndex.Token.ID != storageInvertIndex.Token.ID || memoryInvertedIndex.Token.Term != storageInvertIndex.Token.Term {
		return InvertedIndexValue{}, fmt.Errorf("error: not match inverted index")
	}

	// 生成物
	var merged InvertedIndexValue = InvertedIndexValue{
		Token:          memoryInvertedIndex.Token,
		PostingList:    PostingList{},
		PositionsCount: 0,
		DocsCount:      0,
	}

	// ポスティングリストをマージする
	memoryPostingList := memoryInvertedIndex.PostingList
	storagePostingList := storageInvertIndex.PostingList
	i, j := 0, 0 // i: memoryPostingListのカーソル、 j: storagePostingListのカーソル
	for {
		if i == len(memoryPostingList) {
			merged.DocsCount++
			merged.PositionsCount += storagePostingList[j].PositionsCount
			merged.PostingList = append(merged.PostingList, storagePostingList[j])
			j++
		} else if j == len(storagePostingList) {
			merged.DocsCount++
			merged.PositionsCount += memoryPostingList[i].PositionsCount
			merged.PostingList = append(merged.PostingList, memoryPostingList[i])
			i++
		} else if memoryPostingList[i].DocumentID < storagePostingList[j].DocumentID {
			merged.DocsCount++
			merged.PositionsCount += memoryPostingList[i].PositionsCount
			merged.PostingList = append(merged.PostingList, memoryPostingList[i])
			i++
		} else if memoryPostingList[i].DocumentID > storagePostingList[j].DocumentID {
			merged.DocsCount++
			merged.PositionsCount += storagePostingList[j].PositionsCount
			merged.PostingList = append(merged.PostingList, storagePostingList[j])
			j++
		} else if memoryPostingList[i].DocumentID == storagePostingList[j].DocumentID {
			merged.DocsCount++
			merged.PositionsCount += storagePostingList[j].PositionsCount
			merged.PostingList = append(merged.PostingList, storagePostingList[j])
			i++
			j++
		} else {
			return InvertedIndexValue{}, fmt.Errorf("error: not reachable")
		}
		if i == len(memoryPostingList) && j == len(storagePostingList) {
			break
		}
	}
	return merged, nil
}
