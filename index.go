package stalefish

type InvertIndexHash map[int]InvertedIndexValue

type InvertedIndexValue struct {
	TokenID        int
	PostingList    *PostingsList
	DocsCount      int
	PositionsCount int
}

type Indexer struct {
	Repository      Repository
	Analyzer        Analyzer
	InvertIndexHash InvertIndexHash
}

func (i *Indexer) AddDocument(doc Document) {
	i.Repository.AddDocument(doc)
	docID := i.Repository.GetDocumentID(doc.Title)
	textToPostingLists()

	if 2 > 1 {
		// gohge
	}

	for {
		updatePostings()
	}
}

// textからPostingList
func (i *Indexer) textToPostingLists(docID int, token string, pos int) {

}

func (i *Indexer) tokenToPostingList(docID int, token string, pos int) {
	tokenID := i.Repository.GetTokenID(token)

	var pl *PostingsList

	iiv, ok := i.InvertIndexHash[tokenID]
	if ok {
		pl = iiv.PostingList
		pl.PostionsCount++
	} else {
		pl = &PostingsList{
			DocumentID:    docID,
			PostionsCount: 1,
		}
		i.InvertIndexHash[tokenID] = InvertedIndexValue{
			PostingList: pl,
		}
	}
	pl.Positions = append(pl.Positions, pos)
	iiv.PositionsCount++
}

// ストレージ上のポスティングリストとメモリ上のポスティングリストをマージし保存する
func (i *Indexer) updatePostings() {

}

// マージ元のポスティングリストをストレージから読み出す
func (i *Indexer) fetchPostings() {

}

// type FieldInvertedIndex map[string][]int

// type Index map[string]FieldInvertedIndex

// func (idx Index) Indexing(docs []Document, analyzer Analyzer) {
// 	for _, doc := range docs {
// 		for k, v := range doc.Fields {
// 			fieldInvertedIndex, ok := idx[k]
// 			if !ok {
// 				fieldInvertedIndex = FieldInvertedIndex{}
// 			}
// 			for _, token := range analyzer.Analyze(v) {
// 				ids := fieldInvertedIndex[token]
// 				if ids != nil && ids[len(ids)-1] == doc.ID {
// 					// Don't add same ID twice.
// 					continue
// 				}
// 				fieldInvertedIndex[token] = append(ids, doc.ID)
// 			}
// 			idx[k] = fieldInvertedIndex
// 		}
// 	}
// }

// NOTE: Goだけでなく任意の言語から使えるようにする場合はこれでよかった(?)
// add adds documents to the index.
// func (idx Index) Add(docs []Document, analyzer Analyzer) {
// 	for _, doc := range docs {
// 		rt := doc.Fields.Type()
// 		for i := 0; i < rt.NumField(); i++ {
// 			field := rt.Field(i)
// 			// kind := field.Type.Kind()                   // 型
// 			value := doc.Fields.FieldByName(field.Name) // value は interface です。

// 			fieldInvertedIndex := FieldInvertedIndex{}
// 			for _, token := range analyzer.Analyze(value.String()) {
// 				ids := fieldInvertedIndex[token]
// 				if ids != nil && ids[len(ids)-1] == doc.ID {
// 					// Don't add same ID twice.
// 					continue
// 				}
// 				fieldInvertedIndex[token] = append(ids, doc.ID)
// 			}
// 			idx.FieldToInvertedIndex[field.Name] = fieldInvertedIndex
// 		}
// 	}
// }
