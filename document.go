package main

import (
	"fmt"
	"reflect"
)

type DocumentBuilder struct {
	Fields []reflect.StructField
}

func NewDocumentBuilder() *DocumentBuilder {
	return &DocumentBuilder{}
}

// TODO: reflectよくわからん
func (b *DocumentBuilder) AddField(name, typ string) error {
	var t reflect.Type
	switch typ {
	case "string":
		t = reflect.TypeOf("")
	// case "int":
	// 	t = reflect.TypeOf(1)
	// case "float":
	// 	t = reflect.TypeOf(0.1)
	// case "bool":
	// 	t = reflect.TypeOf(true)
	default:
		return fmt.Errorf("error: invalid type")
	}
	b.Fields = append(
		b.Fields,
		reflect.StructField{
			Name: name,
			Type: t,
		},
	)
	return nil
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
	ID     int
	Fields reflect.Value
}

func (d *Document) Get(name string) (interface{}, error) {
	rv := d.Fields.FieldByName(name)
	if (rv == reflect.Value{}) {
		return nil, fmt.Errorf("error: not found given field")
	}
	return rv.Interface(), nil
}

func (d *Document) Set(name string, value interface{}) error {
	rv := d.Fields.FieldByName(name)
	if (rv == reflect.Value{}) {
		return fmt.Errorf("error: not found given field")
	}
	switch rv.Type() {
	case reflect.TypeOf(""):
		rv.SetString(value.(string))
	// case reflect.TypeOf(1):
	// 	rv.SetInt(int64(value.(int)))
	// case reflect.TypeOf(0.1):
	// 	rv.SetFloat(value.(float64))
	// case reflect.TypeOf(true):
	// 	rv.SetBool(value.(bool))
	default:
		return fmt.Errorf("error: don't match any type")
	}
	return nil
}
