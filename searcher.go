package stalefish

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
	tokenStream *TokenStream
	logic       Logic
	storage     Storage
}

func NewMatchSearcher(tokenStream *TokenStream, logic Logic, storage Storage) MatchSearcher {
	return MatchSearcher{
		tokenStream: tokenStream,
		logic:       logic,
		storage:     storage,
	}
}

func (ms MatchSearcher) Search() ([]Document, error) {
	// トークンストリームが空ならマッチするドキュメントなし
	if ms.tokenStream.size() == 0 {
		return []Document{}, nil
	}

	terms := make([]string, ms.tokenStream.size())
	for i, t := range ms.tokenStream.Tokens {
		terms[i] = t.Term
	}
	// IDを取得するため
	tokens, err := ms.storage.GetTokensByTerms(terms)
	if err != nil {
		return nil, err
	}

	ids := make([]TokenID, len(tokens))
	for i, t := range tokens {
		ids[i] = t.ID
	}
	// ストレージから転置リストを取得する
	invertedIndexValues, err := ms.storage.GetInvertedIndexesByTokenIDs(ids)
	if err != nil {
		return nil, err
	}
	// トークンごとの転置リストを取得
	list := make([]*Postings, len(invertedIndexValues))
	for i, v := range invertedIndexValues {
		list[i] = v.PostingList
	}

	var matchedIds []DocumentID
	if ms.logic == AND {
		matchedIds = andMatch(list)
	} else if ms.logic == OR {
		matchedIds = orMatch(list)
	}
	return ms.storage.GetDocuments(matchedIds)
}

// AND検索
func andMatch(list []*Postings) []DocumentID {
	var ids []DocumentID = make([]DocumentID, 0)
	for notAllNil(list) {
		if isSameDocumentId(list) {
			ids = append(ids, list[0].DocumentID)
			for i := range list {
				list[i] = list[i].Next
			}
			continue
		}
		idx := minIdx(list)
		list[idx] = list[idx].Next
	}
	return ids
}

// OR検索
func orMatch(list []*Postings) []DocumentID {
	var ids []DocumentID = make([]DocumentID, 0)
	for !allNil(list) {
		for i, l := range list {
			if l == nil {
				continue
			}
			ids = append(ids, l.DocumentID)
			list[i] = list[i].Next
		}
	}
	return uniqueDocumentId(ids)
}

// 最小のドキュメントIDを持つポスティングリストのインデックス
func minIdx(list []*Postings) int {
	min := 0
	for i := 1; i < len(list); i++ {
		if list[min].DocumentID > list[i].DocumentID {
			min = i
		}
	}
	return min
}

// スライスに含まれる全てのポスティングリストが指すキュメントIDが同じかどうか
func isSameDocumentId(list []*Postings) bool {
	for i := 0; i < len(list)-1; i++ {
		if list[i].DocumentID != list[i+1].DocumentID {
			return false
		}
	}
	return true
}

// 全てのポスティングリストがnilではない
func notAllNil(list []*Postings) bool {
	for _, p := range list {
		if p == nil {
			return false
		}
	}
	return true
}

// 全てのポスティングリストがnil
func allNil(list []*Postings) bool {
	for _, p := range list {
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
	uniq := make([]DocumentID, len(m))
	for i := range m {
		uniq = append(uniq, i)
	}
	return uniq
}

type PhraseSearcher struct {
	tokenStream *TokenStream
	storage     Storage
}

func NewPhraseSearcher(tokenStream *TokenStream, storage Storage) PhraseSearcher {
	return PhraseSearcher{
		tokenStream: tokenStream,
		storage:     storage,
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
	if ps.tokenStream.size() == 0 {
		return []Document{}, nil
	}

	// トークンごとの転置リストを取得
	invertedIndexValues := make(InvertedIndexValues, ps.tokenStream.size())
	for i, t := range ps.tokenStream.Tokens {
		// IDを取得するため
		token, err := ps.storage.GetTokenByTerm(t.Term)
		if err != nil {
			return nil, err
		}
		// ストレージから転置リストを取得する
		invertedIndexValue, err := ps.storage.GetInvertedIndexByTokenID(token.ID)
		if err != nil {
			return nil, err
		}
		invertedIndexValues[i] = invertedIndexValue
	}
	list := make([]*Postings, ps.tokenStream.size())
	for i, v := range invertedIndexValues {
		list[i] = v.PostingList
	}

	var matchedDocumentIDs []DocumentID
	for {
		if isSameDocumentId(list) { // カーソルが指す全てのDocIDが等しい時
			// フレーズが等しければ結果に追加
			if isPhraseMatch(ps.tokenStream, list) {
				matchedDocumentIDs = append(matchedDocumentIDs, list[0].DocumentID)
			}

			// カーソルを全て動かす
			for i := range list {
				list[i] = list[i].Next
			}
		} else {
			// 一番小さいカーソルを動かす
			idx := minIdx(list)
			list[idx] = list[idx].Next
		}

		if !notAllNil(list) {
			break
		}
	}
	return ps.storage.GetDocuments(matchedDocumentIDs)
}

// [
//	[5,9,20],
//	[2,6,30],
//	[7],
// ]
// が与えられて、相対ポジションに変換してintスライス間で共通する要素があるか判定する
func isPhraseMatch(tokenStream *TokenStream, list []*Postings) bool {
	// 相対ポジションリストを作る
	relativePositionsList := make([][]uint64, tokenStream.size())
	for i := 0; i < tokenStream.size(); i++ {
		relativePositionsList[i] = decrementUintSlice(list[i].Positions, uint64(i))
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
	for i, e := range s {
		s[i] = e - n
	}
	return s
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
