package stalefish

import (
	"math"
	"sort"
)

type Sorter interface {
	Sort([]Document, InvertedIndex, []Token) ([]Document, error)
}

type TfIdfSorter struct {
	storage Storage
}

func NewTfIdfSorter(storage Storage) *TfIdfSorter {
	return &TfIdfSorter{
		storage: storage,
	}
}

func (s *TfIdfSorter) Sort(docs []Document, invertedIndex InvertedIndex, tokens []Token) ([]Document, error) {
	allDocsCount, err := s.storage.CountDocuments()
	if err != nil {
		return nil, err
	}

	var documentScores documentScores = make([]documentScore, len(docs))
	for i, doc := range docs {
		var sum float64
		for _, token := range tokens {
			postingList := invertedIndex[token.ID]
			tf := float64(postingList.AppearanceCountInDocument(doc.ID)) / float64(doc.TokenCount)
			idf := math.Log2(float64(allDocsCount)/float64(postingList.Size()+1)) + 1
			sum += tf * idf
		}
		documentScores[i] = NewDocumentScore(doc, sum)
	}
	sort.Sort(sort.Reverse(documentScores))
	return documentScores.toDocuments(), nil
}

type documentScore struct {
	document Document
	score    float64
}

func NewDocumentScore(doc Document, score float64) documentScore {
	return documentScore{
		document: doc,
		score:    score,
	}
}

type documentScores []documentScore

func (ds documentScores) Len() int { return len(ds) }

func (ds documentScores) Less(i, j int) bool { return ds[i].score < ds[j].score }

func (ds documentScores) Swap(i, j int) { ds[i], ds[j] = ds[j], ds[i] }

func (ds documentScores) toDocuments() []Document {
	docs := make([]Document, len(ds))
	for i, d := range ds {
		docs[i] = d.document
	}
	return docs
}
