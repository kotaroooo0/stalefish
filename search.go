package main

import "fmt"

type Searcher struct {
	Index    Index
	Analyzer Analyzer
}

type Query struct {
	Keyword string
	Fields  []string
}

// search queries the index for the given text.
func (s Searcher) Search(q Query) ([]int, error) {
	var r []int
	for _, field := range q.Fields {
		invertedIndex, ok := s.Index[field]
		if !ok {
			return nil, fmt.Errorf("error: invalid field name")
		}
		for _, token := range s.Analyzer.Analyze(q.Keyword) {
			if ids, ok := invertedIndex[token]; ok {
				if r == nil {
					r = ids
				} else {
					r = intersection(r, ids)
				}
			}
		}
	}
	return r, nil
}

// intersection returns the set intersection between a and b.
// a and b have to be sorted in ascending order and contain no duplicates.
func intersection(a []int, b []int) []int {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	r := make([]int, 0, maxLen)
	var i, j int
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			r = append(r, a[i])
			i++
			j++
		}
	}
	return r
}
