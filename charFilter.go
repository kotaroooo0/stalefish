package stalefish

import "strings"

type CharFilter interface {
	filter(string) string
}

type MappingCharFilter struct {
	mapper map[string]string // key->valueにマッピングする
}

func NewMappingCharFilter(mapper map[string]string) *MappingCharFilter {
	return &MappingCharFilter{mapper: mapper}
}

func (c MappingCharFilter) filter(s string) string {
	for k, v := range c.mapper {
		s = strings.Replace(s, k, v, -1)
	}
	return s
}
