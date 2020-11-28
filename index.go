package stalefish

import "fmt"

// 転置インデックス。TokenIDー>転置リスト
type InvertIndexHash map[TokenID]InvertedIndexValue

type TokenID int

// 転置リスト
type InvertedIndexValue struct {
	TokenID        TokenID // トークンID
	Token          string
	PostingList    PostingList // トークンを含むポスティングスリスト
	DocsCount      int         // トークンを含む文書数
	PositionsCount int         // 全文書内でのトークンの出現数
}

// ポスティングリスト。文書IDのリンクリスト
type PostingList []Posting

type Posting struct {
	DocumentID     DocumentID // 文書のID
	Positions      []int      // 文書中の位置情報
	PositionsCount int        // 文書中の位置情報の数
}

type Indexer struct {
	Storage         Storage
	Analyzer        Analyzer
	InvertIndexHash InvertIndexHash
}

const INDEX_SIZE_THRESHOLD = 100

// 1.文書からトークンを取り出す
// 2.トークンごとにポスティングリストを作って、それをメモリ上の転置インデックスに追加する
// 3.メモリ上の転置インデックスがある程度のサイズになったら、ストレージ上の転置インデックスにマージする
func (i *Indexer) AddDocument(doc Document) error {

	// ストレージに文書を格納
	i.Storage.AddDocument(doc)

	// 文書IDを取得
	docID := i.Storage.GetDocumentID(doc.Title)

	// 文書からポスティングリストの集合(転置リスト)を構築
	i.textToPostingLists(DocumentID(docID), doc.Title)
	i.textToPostingLists(DocumentID(docID), doc.Body)

	// ストレージ上の転置インデックスにマージする
	if len(i.InvertIndexHash) >= INDEX_SIZE_THRESHOLD {
		for tokenID, invertedIndex := range i.InvertIndexHash {
			storageInvertIndex := i.fetchPosting(tokenID)
			merged, err := mergeInvertedIndex(invertedIndex, storageInvertIndex)
			if err != nil {
				return err
			}
			i.updateInvertIndex(merged)
		}
		// メモリの転置インデックスをリセット
		i.InvertIndexHash = InvertIndexHash{}
	}
	return nil
}

// textからPostingList
func (i *Indexer) textToPostingLists(docID DocumentID, text string) {
	tokens := i.Analyzer.Analyze(text)
	for pos, token := range tokens {
		i.tokenToPostingList(docID, token, pos)
	}
}

func (i *Indexer) tokenToPostingList(docID DocumentID, token string, pos int) {
	tokenID := i.Storage.GetTokenID(token)
	iiv, ok := i.InvertIndexHash[tokenID]
	if !ok {
		i.InvertIndexHash[tokenID] = InvertedIndexValue{
			TokenID: tokenID,
			Token:   token,
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
		var targetPostingIdx int
		isExistPosting := false
		for i := 0; i < len(iiv.PostingList); i++ {
			p := iiv.PostingList[i]
			if p.DocumentID == docID {
				targetPostingIdx = i
				isExistPosting = true
				break
			}
		}
		if isExistPosting {
			iiv.PostingList[targetPostingIdx].Positions = append(iiv.PostingList[targetPostingIdx].Positions, pos)
			iiv.PostingList[targetPostingIdx].PositionsCount++
			iiv.PositionsCount++
		} else {
			iiv.PostingList = append(iiv.PostingList,
				Posting{
					DocumentID:     docID,
					Positions:      []int{pos},
					PositionsCount: 1,
				})
			iiv.DocsCount++
			iiv.PositionsCount++
		}
	}
}

// ストレージ上のポスティングリストとメモリ上のポスティングリストをマージし保存する
func (i *Indexer) updateInvertIndex(invertedIndex InvertedIndexValue) error {
	return i.Storage.UpsertInvertedIndex(invertedIndex)
}

// マージ元のポスティングリストをストレージから読み出す
func (i *Indexer) fetchPosting(tokenID TokenID) InvertedIndexValue {
	return i.Storage.GetInvertIndexByTokenID(tokenID)
}

func mergeInvertedIndex(memoryInvertedIndex, storageInvertIndex InvertedIndexValue) (InvertedIndexValue, error) {
	if memoryInvertedIndex.TokenID != storageInvertIndex.TokenID || memoryInvertedIndex.Token != storageInvertIndex.Token {
		return InvertedIndexValue{}, fmt.Errorf("error: not match inverted index")
	}
	var merged InvertedIndexValue = InvertedIndexValue{
		TokenID:        memoryInvertedIndex.TokenID,
		Token:          memoryInvertedIndex.Token,
		PostingList:    PostingList{},
		PositionsCount: 0,
		DocsCount:      0,
	}
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
