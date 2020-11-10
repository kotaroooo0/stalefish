package stalefish

import "strings"

type CharFilter interface {
	Filter(string) string
}

type MappingCharFilter struct {
	Mapper map[string]string // key->valueにマッピングする
}

func (c MappingCharFilter) Filter(s string) string {
	for k, v := range c.Mapper {
		s = strings.Replace(s, k, v, -1)
	}
	return s
}
