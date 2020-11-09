package stalefish

type CharFilter interface {
	Filter(string) string
}
