package main

type CharFilter interface {
	Filter(string) string
}
