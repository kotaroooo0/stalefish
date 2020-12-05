package stalefish

import (
	"fmt"
)

type Searcher interface {
	Search() ([]Document, error)
}

type MatchAllSearcher struct {
	Storage Storage
}

func NewMatchAllSearcher() (MatchAllSearcher, error) {
	return MatchAllSearcher{}, nil
}

type PhraseSearcher struct {
	Tokens  []string
	Storage Storage
}

func NewPhraseSearcher(tokens []string) (PhraseSearcher, error) {
	return PhraseSearcher{
		Tokens: tokens,
	}, nil
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
	var invertIndexHash InvertIndexHash
	invertedIndexValues := make(InvertedIndexValues, len(ps.Tokens))
	for i, t := range ps.Tokens {
		// トークンに対するポスティングリストが空ならそこで空スライスを返す
		tokenID := ps.Storage.GetTokenID(t)
		invertedIndexValue := ps.Storage.GetInvertIndexByTokenID(tokenID)
		invertIndexHash[tokenID] = invertedIndexValue
		invertedIndexValues[i] = invertedIndexValue
	}

	// ドキュメント数が少ない順にソート
	// sort.Sort(invertedIndexValues)

	var matchedDocumentIDs []DocumentID

	cursors := make([]int, len(ps.Tokens))
	sizes := make([]int, len(ps.Tokens))
	docIDs := make([]DocumentID, len(ps.Tokens))
	for i := 0; i < len(ps.Tokens); i++ {
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
			if isPhraseMatch(ps.Tokens, invertedIndexValues, cursors) {
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

	// AND
	// var res PostingList
	// for i, v := range invertedIndexValues {
	// 	if i == 0 {
	// 		res = invertedIndexValues[i].PostingList
	// 		continue
	// 	}
	// 	// resに含まれいるドキュメントがinvertedIndexValues[i].PostingListに含まれているか
	// 	for j, r := range res {
	// 		if !contains(v.PostingList, r.DocumentID) {
	// 			res = remove(res, j)
	// 		}
	// 	}
	// }
}

// [
//	[5,9,20],
//	[2,6,30],
//	[7],
// ]
// が与えられて、相対ポジションに変換してintスライス間で共通する要素があるか判定する
func isPhraseMatch(tokens []string, invertedIndexValues InvertedIndexValues, cursors []int) bool {
	var relativePositionsList [][]int
	for i := range tokens {
		relativePositionsList[i] = decrementIntSlice(invertedIndexValues[i].PostingList[cursors[i]].Positions, i)
	}

	for _, sourceRelativePosition := range relativePositionsList[0] {
		for _, targetRelativePositions := range relativePositionsList[1:] {
			if !intContains(targetRelativePositions, sourceRelativePosition) {
				return false
			}
		}
	}
	return true
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

// // pからi番目の要素を削除する
// func remove(p PostingList, i int) PostingList {
// 	return append(p[:i], p[i+1:]...)
// }

// // pにDocumentIDが含まれているか
// func contains(p PostingList, docID DocumentID) bool {
// 	for _, v := range p {
// 		if docID == v.DocumentID {
// 			return true
// 		}
// 	}
// 	return false
// }

// sにnが含まれているか
func intContains(s []int, n int) bool {
	for _, v := range s {
		if n == v {
			return true
		}
	}
	return false
}
