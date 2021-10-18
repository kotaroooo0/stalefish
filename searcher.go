package stalefish

import (
	"sort"
)

type Logic int

const (
	AND Logic = iota + 1
	OR
)

func (l Logic) String() string {
	switch l {
	case AND:
		return "AND"
	case OR:
		return "OR"
	default:
		return "Unknown"
	}
}

type Searcher interface {
	Search() ([]Document, error)
}

type MatchAllSearcher struct {
	storage Storage
}

func NewMatchAllSearcher(storage Storage) MatchAllSearcher {
	return MatchAllSearcher{storage: storage}
}

func (ms MatchAllSearcher) Search() ([]Document, error) {
	return ms.storage.GetAllDocuments()
}

type MatchSearcher struct {
	tokenStream TokenStream
	logic       Logic
	storage     Storage
	sorter      Sorter
}

func NewMatchSearcher(tokenStream TokenStream, logic Logic, storage Storage, sorter Sorter) MatchSearcher {
	return MatchSearcher{
		tokenStream: tokenStream,
		logic:       logic,
		storage:     storage,
		sorter:      sorter,
	}
}

func (ms MatchSearcher) Search() ([]Document, error) {
	// トークンストリームが空ならマッチするドキュメントなし
	if ms.tokenStream.Size() == 0 {
		return []Document{}, nil
	}

	// IDを取得するため
	tokens, err := ms.storage.GetTokensByTerms(ms.tokenStream.Terms())
	if err != nil {
		return nil, err
	}

	// 対応トークンが一つも存在していない場合、結果は空
	if len(tokens) == 0 {
		return []Document{}, nil
	}

	// AND検索の場合、対応するトークンが全て存在していなかったら結果は空
	if ms.logic == AND && len(tokens) != len(ms.tokenStream.Terms()) {
		return []Document{}, nil
	}

	// ストレージから転置リストを取得する
	inverted, err := ms.storage.GetInvertedIndexByTokenIDs(tokenIDs(tokens))
	if err != nil {
		return nil, err
	}

	// トークンごとのポスティングを取得
	postings := make([]*Postings, len(inverted))
	for i, t := range tokens {
		postings[i] = inverted[t.ID].Postings
	}

	// ポスティングリストを走査
	var matchedIds []DocumentID
	if ms.logic == AND {
		matchedIds = andMatch(postings)
	} else if ms.logic == OR {
		matchedIds = orMatch(postings)
	}
	documents, err := ms.storage.GetDocuments(matchedIds)
	if err != nil {
		return nil, err
	}
	if ms.sorter == nil {
		return documents, nil
	}
	return ms.sorter.Sort(documents, inverted, tokens)
}

func tokenIDs(tokens []Token) []TokenID {
	ids := make([]TokenID, len(tokens))
	for i, t := range tokens {
		ids[i] = t.ID
	}
	return ids
}

// AND検索
func andMatch(postings []*Postings) []DocumentID {
	var ids []DocumentID = make([]DocumentID, 0)
	for notAllNil(postings) {
		if isSameDocumentId(postings) {
			ids = append(ids, postings[0].DocumentID)
			for i := range postings {
				postings[i] = postings[i].Next
			}
			continue
		}
		idx := minIdx(postings)
		postings[idx] = postings[idx].Next
	}
	return ids
}

// OR検索
func orMatch(postings []*Postings) []DocumentID {
	ids := []DocumentID{}
	for !allNil(postings) {
		for i, p := range postings {
			if p == nil {
				continue
			}
			ids = append(ids, p.DocumentID)
			postings[i] = postings[i].Next
		}
	}
	return uniqueDocumentId(ids)
}

// 最小のドキュメントIDを持つポスティングリストのインデックス
func minIdx(postings []*Postings) int {
	min := 0
	for i := 1; i < len(postings); i++ {
		if postings[min].DocumentID > postings[i].DocumentID {
			min = i
		}
	}
	return min
}

// スライスに含まれる全てのポスティングリストが指すキュメントIDが同じかどうか
func isSameDocumentId(postings []*Postings) bool {
	for i := 0; i < len(postings)-1; i++ {
		if postings[i].DocumentID != postings[i+1].DocumentID {
			return false
		}
	}
	return true
}

// 全てのポスティングリストがnilではない
func notAllNil(postings []*Postings) bool {
	for _, p := range postings {
		if p == nil {
			return false
		}
	}
	return true
}

// 全てのポスティングリストがnil
func allNil(postings []*Postings) bool {
	for _, p := range postings {
		if p != nil {
			return false
		}
	}
	return true
}

func uniqueDocumentId(ids []DocumentID) []DocumentID {
	m := make(map[DocumentID]struct{})
	for _, id := range ids {
		m[id] = struct{}{}
	}
	uniq := []DocumentID{}
	for k := range m {
		uniq = append(uniq, k)
	}
	sort.Slice(uniq, func(i, j int) bool { return uniq[i] < uniq[j] })
	return uniq
}

type PhraseSearcher struct {
	tokenStream TokenStream
	storage     Storage
	sorter      Sorter
}

func NewPhraseSearcher(tokenStream TokenStream, storage Storage, sorter Sorter) PhraseSearcher {
	return PhraseSearcher{
		tokenStream: tokenStream,
		storage:     storage,
		sorter:      sorter,
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
	// トークンストリームが空ならマッチするドキュメントなし
	if ps.tokenStream.Size() == 0 {
		return []Document{}, nil
	}

	// IDを取得するため
	tokens, err := ps.storage.GetTokensByTerms(ps.tokenStream.Terms())
	if err != nil {
		return nil, err
	}

	// 対応するトークンが全て存在していなかったら結果は空
	if len(tokens) != len(ps.tokenStream.Terms()) {
		return []Document{}, nil
	}

	// ストレージから転置リストを取得する
	inverted, err := ps.storage.GetInvertedIndexByTokenIDs(tokenIDs(tokens))
	if err != nil {
		return nil, err
	}

	// トークンごとのポスティングを取得
	postings := make([]*Postings, len(inverted))
	for i, t := range tokens {
		postings[i] = inverted[t.ID].Postings
	}

	var matchedDocumentIDs []DocumentID
	for {
		if isSameDocumentId(postings) { // カーソルが指す全てのDocIDが等しい時
			// フレーズが等しければ結果に追加
			if isPhraseMatch(ps.tokenStream, postings) {
				matchedDocumentIDs = append(matchedDocumentIDs, postings[0].DocumentID)
			}

			// カーソルを全て動かす
			for i := range postings {
				postings[i] = postings[i].Next
			}
		} else {
			// 一番小さいカーソルを動かす
			idx := minIdx(postings)
			postings[idx] = postings[idx].Next
		}

		if !notAllNil(postings) {
			break
		}
	}
	documents, err := ps.storage.GetDocuments(matchedDocumentIDs)
	if err != nil {
		return nil, err
	}
	if ps.sorter == nil {
		return documents, nil
	}
	return ps.sorter.Sort(documents, inverted, tokens)
}

// [
//	[5,9,20],
//	[2,6,30],
//	[7],
// ]
// が与えられて、相対ポジションに変換してintスライス間で共通する要素があるか判定する
func isPhraseMatch(tokenStream TokenStream, postings []*Postings) bool {
	// 相対ポジションリストを作る
	relativePositionsList := make([][]uint64, tokenStream.Size())
	for i := 0; i < tokenStream.Size(); i++ {
		relativePositionsList[i] = decrementUintSlice(postings[i].Positions, uint64(i))
	}

	// 共通の要素が存在すればフレーズが存在するということになる
	postitions := relativePositionsList[0]
	for _, relativePositions := range relativePositionsList[1:] {
		if !hasCommonElement(postitions, relativePositions) {
			return false
		}
	}
	return true
}

func decrementUintSlice(s []uint64, n uint64) []uint64 {
	result := make([]uint64, len(s))
	for i, e := range s {
		result[i] = e - n
	}
	return result
}

func hasCommonElement(s1 []uint64, s2 []uint64) bool {
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}
