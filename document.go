package main

import "reflect"

type DocumentBuilder struct {
	Fields []reflect.StructField
}

func NewDocumentBuilder() *DocumentBuilder {
	return &DocumentBuilder{}
}

// string, int, float, bool フィールドを持つ
func (b *DocumentBuilder) AddField(n, t string) {
	switch
	case
	b.Fields = append(
		b.Fields,
		reflect.StructField{
			Name: n,
			Type: t,
		})
}

func (b *DocumentBuilder) Build() DocumentType {
	s := reflect.StructOf(b.Fields)
	index := make(map[string]int)
	for i := 0; i < s.NumField(); i++ {
		index[s.Field(i).Name] = i
	}
	return DocumentType{Struct: s}
}

type DocumentType struct {
	Struct reflect.Type
}

func (d *DocumentType) New() *Document {
	return &Document{
		Fields: reflect.New(d.Struct).Elem(),
	}
}

type Document struct {
	Fields reflect.Value
}
