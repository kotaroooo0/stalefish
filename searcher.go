package stalefish

import (
	"sort"
)

type Logic int

const (
	AND Logic = iota + 1
	OR
)

type Searcher interface {
	Search() ([]Document, error)
}

type MatchAllSearcher struct {
	storage Storage
}

func NewMatchAllSearcher(storage Storage) MatchAllSearcher {
	return MatchAllSearcher{
		storage: storage,
	}
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
	// tokenStreamが空なら、マッチするドキュメントなしでリターン
	if ms.tokenStream.Size() == 0 {
		return []Document{}, nil
	}

	// トークンIDを取得するためにストレージをREAD
	tokens, err := ms.storage.GetTokensByTerms(ms.tokenStream.Terms())
	if err != nil {
		return nil, err
	}

	// 対応トークンが一つも存在しないなら、マッチするドキュメントなしでリターン
	if len(tokens) == 0 {
		return []Document{}, nil
	}

	// AND検索で対応するトークンが全て存在していなかったら、マッチするドキュメントなしでリターン
	if ms.logic == AND && len(tokens) != len(ms.tokenStream.Terms()) {
		return []Document{}, nil
	}

	// ストレージから転置インデックスをREAD
	inverted, err := ms.storage.GetInvertedIndexByTokenIDs(tokenIDs(tokens))
	if err != nil {
		return nil, err
	}

	// ポスティングリストを抽出
	postings := make([]*Postings, len(inverted))
	for i, t := range tokens {
		postings[i] = inverted[t.ID].Postings
	}

	// ポスティングリストを走査しマッチするドキュメントIDを取得
	var matchedIds []DocumentID
	if ms.logic == AND {
		matchedIds = andMatch(postings)
	} else if ms.logic == OR {
		matchedIds = orMatch(postings)
	}

	// ドキュメントIDからドキュメントを取得
	documents, err := ms.storage.GetDocuments(matchedIds)
	if err != nil {
		return nil, err
	}

	// sorterが指定されていればドキュメントをソートしてリターン
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
			next(postings)
			continue
		}
		idx := minDocumentIDIndex(postings)
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

// ポスティングリストのスライスから最小のドキュメントIDを指しているポスティングリストのインデックス
func minDocumentIDIndex(postings []*Postings) int {
	min := 0
	for i := 1; i < len(postings); i++ {
		if postings[min].DocumentID > postings[i].DocumentID {
			min = i
		}
	}
	return min
}

// スライスに含まれる全てのポスティングリストが指すドキュメントIDが同じかどうか
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

// 全てのポスティングリストを次のポスティングを指すようにする
func next(postings []*Postings) {
	for i := range postings {
		postings[i] = postings[i].Next
	}
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

// ドキュメントIDのスライスで重複を削除
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

func (ps PhraseSearcher) Search() ([]Document, error) {
	// tokenStreamが空なら、マッチするドキュメントなしでリターン
	if ps.tokenStream.Size() == 0 {
		return []Document{}, nil
	}

	// トークンIDを取得するためにストレージをREAD
	tokens, err := ps.storage.GetTokensByTerms(ps.tokenStream.Terms())
	if err != nil {
		return nil, err
	}

	// 対応トークンが一つも存在しないなら、マッチするドキュメントなしでリターン
	if len(tokens) != len(ps.tokenStream.Terms()) {
		return []Document{}, nil
	}

	// ストレージから転置インデックスをREAD
	inverted, err := ps.storage.GetInvertedIndexByTokenIDs(tokenIDs(tokens))
	if err != nil {
		return nil, err
	}

	// ポスティングリストを抽出
	postings := make([]*Postings, len(inverted))
	for i, t := range tokens {
		postings[i] = inverted[t.ID].Postings
	}

	// ポスティングリストを走査しマッチするドキュメントIDを取得
	var ids []DocumentID
	for notAllNil(postings) {
		if isSameDocumentId(postings) { // カーソルが指す全てのDocIDが等しい時
			// フレーズが等しければ結果に追加
			if isPhraseMatch(ps.tokenStream, postings) {
				ids = append(ids, postings[0].DocumentID)
			}
			// カーソルを全て動かす
			next(postings)
			continue
		}
		// 一番小さいカーソルを動かす
		idx := minDocumentIDIndex(postings)
		postings[idx] = postings[idx].Next
	}

	// ドキュメントIDからドキュメントを取得
	documents, err := ps.storage.GetDocuments(ids)
	if err != nil {
		return nil, err
	}

	// sorterが指定されていればドキュメントをソートしてリターン
	if ps.sorter == nil {
		return documents, nil
	}
	return ps.sorter.Sort(documents, inverted, tokens)
}

// フレーズを含むか判定
func isPhraseMatch(tokenStream TokenStream, postings []*Postings) bool {
	// 相対ポジションリストを作る
	relativePositionsList := make([][]uint64, tokenStream.Size())
	for i := 0; i < tokenStream.Size(); i++ {
		relativePositionsList[i] = decremenSlice(postings[i].Positions, uint64(i))
	}

	// 共通の要素が存在すればフレーズが存在する
	return hasCommon(relativePositionsList)
}

// uint64スライスsの各要素をnだけデクリメント
func decremenSlice(s []uint64, n uint64) []uint64 {
	result := make([]uint64, len(s))
	for i, e := range s {
		result[i] = e - n
	}
	return result
}

// 複数のスライスが共通の要素を持っているか判定
func hasCommon(ss [][]uint64) bool {
	s0 := ss[0]
	for _, s1 := range ss[1:] {
		hasCommon := false
		for _, v1 := range s0 {
			for _, v2 := range s1 {
				if v1 == v2 {
					hasCommon = true
				}
			}
		}
		if !hasCommon {
			return false
		}
	}
	return true
}
