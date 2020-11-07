package main

// index is an inverted index. It maps tokens to document ID.
type FieldInvertedIndex map[string][]int

type Index map[string]FieldInvertedIndex

// add adds documents to the index.
func (idx Index) Add(docs []Document, analyzer Analyzer) {
	for _, doc := range docs {
		rt := doc.Fields.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			// kind := field.Type.Kind()                   // 型
			value := doc.Fields.FieldByName(field.Name) // value は interface です。

			fieldInvertedIndex := FieldInvertedIndex{}
			for _, token := range analyzer.Analyze(value.String()) {
				ids := fieldInvertedIndex[token]
				if ids != nil && ids[len(ids)-1] == doc.ID {
					// Don't add same ID twice.
					continue
				}
				fieldInvertedIndex[token] = append(ids, doc.ID)
			}
			idx.FieldToInvertedIndex[field.Name] = fieldInvertedIndex
		}
	}
}
