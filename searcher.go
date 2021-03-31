package stalefish

import (
	"fmt"
)

type Logic int

const (
	AND Logic = iota
	OR
)

type Searcher interface {
	Search() ([]Document, error)
}

type MatchAllSearcher struct {
	Storage Storage
}

func NewMatchAllSearcher(storage Storage) MatchAllSearcher {
	return MatchAllSearcher{Storage: storage}
}

func (ms MatchAllSearcher) Search() ([]Document, error) {
	return ms.Storage.GetAllDocuments()
}

type MatchSearcher struct {
	TokenStream *TokenStream
	Logic       Logic
	Storage     Storage
}

func NewMatchSearcher(tokenStream *TokenStream, logic Logic, storage Storage) MatchSearcher {
	return MatchSearcher{
		TokenStream: tokenStream,
		Logic:       logic,
		Storage:     storage,
	}
}

func (ms MatchSearcher) Search() ([]Document, error) {
	if ms.TokenStream.size() == 0 {
		return []Document{}, nil
	}
	invertedIndexValues := make(InvertedIndexValues, ms.TokenStream.size())
	for i, t := range ms.TokenStream.Tokens {
		token, err := ms.Storage.GetTokenByTerm(t.Term)
		if err != nil {
			return nil, err
		}
		// ストレージから転置リストを取得する
		invertedIndexValue, err := ms.Storage.GetInvertedIndexByTokenID(token.ID)
		if err != nil {
			return nil, err
		}
		invertedIndexValues[i] = invertedIndexValue
	}
	var matchedDocumentIDs []DocumentID
	checked := make(map[DocumentID]struct{})
	if ms.Logic == AND {
		for _, p := range invertedIndexValues[0].PostingList {
			flag := true
			for i := 1; i < ms.TokenStream.size(); i++ {
				docIDs := make([]DocumentID, len(invertedIndexValues[i].PostingList))
				for _, pos := range invertedIndexValues[i].PostingList {
					docIDs = append(docIDs, pos.DocumentID)
				}
				if !contains(docIDs, p.DocumentID) {
					flag = false
					break
				}
			}
			if flag {
				matchedDocumentIDs = append(matchedDocumentIDs, p.DocumentID)
			}
		}
	} else if ms.Logic == OR {
		for i := range ms.TokenStream.Tokens {
			for _, p := range invertedIndexValues[i].PostingList {
				docID := p.DocumentID
				_, ok := checked[docID]
				if !ok {
					matchedDocumentIDs = append(matchedDocumentIDs, docID)
					checked[docID] = struct{}{}
				}
			}
		}
	}
	docs, err := ms.Storage.GetDocuments(matchedDocumentIDs)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func contains(s []DocumentID, e DocumentID) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type PhraseSearcher struct {
	TokenStream *TokenStream
	Storage     Storage
}

func NewPhraseSearcher(tokenStream *TokenStream, storage Storage) PhraseSearcher {
	return PhraseSearcher{
		TokenStream: tokenStream,
		Storage:     storage,
	}
}

// フレーズ検索 AND
// 1, 検索クエリをトークン分割
// 2, そのトークンが出現する文書数が少ない順にソートする
// 3, それぞれのトークンのポスティングリストを取り出し、文書IDとその出現位置のリストを取り出す
// 4, 全てのトークンで同一の文書IDが含まれ、かつ、各トークンの出現位置が連接していれば検索結果に追加する
// 5, 検索結果に追加した各文書と検索クエリのスコアを計算する
// 6, 検索結果を適合度の降順に並べ替える
// 7, 並び替えられた検索結果のうち、上位のものを検索結果として返す
func (ps PhraseSearcher) Search() ([]Document, error) {
	if ps.TokenStream.size() == 0 {
		return []Document{}, nil
	}

	invertedIndexValues := make(InvertedIndexValues, ps.TokenStream.size())
	for i, t := range ps.TokenStream.Tokens {
		// ストレージからTokenIDを取得する
		token, err := ps.Storage.GetTokenByTerm(t.Term)
		if err != nil {
			return nil, err
		}
		// トークンがストレージに存在しなかった時、空を返す
		if token.ID == 0 {
			return []Document{}, nil
		}

		// ストレージから転置リストを取得する
		invertedIndexValue, err := ps.Storage.GetInvertedIndexByTokenID(token.ID)
		if err != nil {
			return nil, err
		}
		// 転置リストがストレージに存在しなかった時、空を返す
		if len(invertedIndexValue.PostingList) == 0 {
			return []Document{}, nil
		}

		invertedIndexValues[i] = invertedIndexValue
	}

	var matchedDocumentIDs []DocumentID
	cursors := make([]int, ps.TokenStream.size())
	sizes := make([]int, ps.TokenStream.size())
	docIDs := make([]DocumentID, ps.TokenStream.size())
	for i := 0; i < ps.TokenStream.size(); i++ {
		sizes[i] = len(invertedIndexValues[i].PostingList)
	}
	for {
		for i, cursor := range cursors {
			docIDs[i] = invertedIndexValues[i].PostingList[cursor].DocumentID
		}

		isSameDocID := true
		for _, id := range docIDs {
			if id != docIDs[0] {
				isSameDocID = false
			}
		}
		if isSameDocID { // カーソルが指す全てのDocIDが等しい時
			// フレーズが等しければ結果に追加
			if isPhraseMatch(ps.TokenStream, invertedIndexValues, cursors) {
				matchedDocumentIDs = append(matchedDocumentIDs, docIDs[0])
			}

			// カーソルを全て動かす
			cursors = incrementAllCursors(cursors)
		} else {
			// 一番小さいカーソルを動かす
			toBeIncrementedCursor := getMinDocumentIDCursor(cursors, invertedIndexValues)
			cursors[toBeIncrementedCursor]++
		}

		// カーソルがどれか一つでもサイズを越えればBreak
		isBreak, err := isSearchEnd(cursors, sizes)
		if err != nil {
			return nil, err
		}
		if isBreak {
			break
		}
	}

	docs, err := ps.Storage.GetDocuments(matchedDocumentIDs)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// [
//	[5,9,20],
//	[2,6,30],
//	[7],
// ]
// が与えられて、相対ポジションに変換してintスライス間で共通する要素があるか判定する
func isPhraseMatch(tokenStream *TokenStream, invertedIndexValues InvertedIndexValues, cursors []int) bool {
	// 相対ポジションリストを作る
	relativePositionsList := make([][]int, tokenStream.size())
	for i := range tokenStream.Tokens {
		relativePositionsList[i] = decrementIntSlice(invertedIndexValues[i].PostingList[cursors[i]].Positions, i)
	}

	// 共通の要素が存在すればフレーズが存在するということになる
	commonElements := relativePositionsList[0]
	for _, relativePositions := range relativePositionsList[1:] {
		commonElements = intCommonElement(commonElements, relativePositions)
	}
	return len(commonElements) >= 1
}

// 探索終了: true
// 探索継続: false
func isSearchEnd(cursors, postingListSizes []int) (bool, error) {
	if len(cursors) != len(postingListSizes) {
		return false, fmt.Errorf("error: invalid arguments(unmatch slice size)")
	}
	for i, c := range cursors {
		if c >= postingListSizes[i] {
			return true, nil
		}
	}
	return false, nil
}

// 一番ドキュメントIDが小さいカーソルを返す
func getMinDocumentIDCursor(cursors []int, invertedIndexValues InvertedIndexValues) (minCursorIdx int) {
	var minDocumentID DocumentID = 9223372036854775807
	for i, c := range cursors {
		if invertedIndexValues[i].PostingList[c].DocumentID < minDocumentID {
			minCursorIdx = i
		}
	}
	return
}

func incrementAllCursors(cursors []int) []int {
	for i, c := range cursors {
		cursors[i] = c + 1
	}
	return cursors
}

func decrementIntSlice(s []int, n int) []int {
	for i, e := range s {
		s[i] = e - n
	}
	return s
}

func intCommonElement(s1 []int, s2 []int) []int {
	ret := []int{}
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				ret = append(ret, v1)
			}
		}
	}
	return ret
}
