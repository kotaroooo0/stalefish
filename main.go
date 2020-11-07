package main

import (
	"fmt"
	"reflect"

	"github.com/k0kubun/pp"
)

type StructBuilder struct {
	field []reflect.StructField
}

func NewStructBuilder() *StructBuilder {
	return &StructBuilder{}
}

func (b *StructBuilder) AddField(fname string, ftype reflect.Type) {
	b.field = append(
		b.field,
		reflect.StructField{
			Name: fname,
			Type: ftype,
		})
}

func (b *StructBuilder) Build() Struct {
	strct := reflect.StructOf(b.field)
	index := make(map[string]int)
	for i := 0; i < strct.NumField(); i++ {
		index[strct.Field(i).Name] = i
	}
	return Struct{strct, index}
}

type Struct struct {
	strct reflect.Type
	index map[string]int
}

func (s *Struct) NewInstance() *Instance {
	instance := reflect.New(s.strct).Elem()
	return &Instance{instance, s.index}
}

type Instance struct {
	internal reflect.Value
	index    map[string]int
}

func (i *Instance) SetString(name, value string) {
	i.internal.FieldByName(name).SetString(value)
}

func (i *Instance) SetBool(name string, value bool) {
	i.internal.FieldByName(name).SetBool(value)
}

func (i *Instance) SetInt(name string, value int) {
	i.internal.FieldByName(name).SetInt(int64(value))
}

func (i *Instance) SetFloat(name string, value float64) {
	i.internal.FieldByName(name).SetFloat(value)
}

func main() {
	b := NewStructBuilder()
	b.AddField("Name", reflect.TypeOf(""))
	b.AddField("Age", reflect.TypeOf(1))
	person := b.Build()

	// pp.Print(person.index)

	i := person.NewInstance()
	i.SetString("Name", "gopher")
	i.SetInt("Age", 8)
	fmt.Println(i.internal.FieldByName("Name"))

	r := i.internal.FieldByName("Age")
	pp.Print(r.Kind())
	fmt.Println(r.Kind())
	// fmt.Println(r.Interface().(r.Kind()))
	// var hoge int64 = i.internal.FieldByName("Age").
	// pp.Print(i.Value())
	// pp.Print(i.Pointer())

	db := NewDocumentBuilder()
	db.AddField("Address", "string")
	db.AddField("Name", "string")
	// err := db.AddField("Number", "hoge")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	dt := db.Build()
	dv := dt.New()
	dv.Set("Address", "Fukui")
	dv.Set("Number", 1241)
	dv.Set("HOGE", 1241)

	// dv.Fields.FieldByName("Address").SetString("x string")
	// dv.Fields.FieldByName("Number").SetInt(41241)
	// dv.Fields.SetString("goge")
	fmt.Println(dv.Fields.FieldByName("Age"))
	pp.Print(dv.Get("Address"))
	pp.Print(dv.Get("Age"))
	pp.Print(dv.Get("Number"))

	hoge, _ := dv.Get("Address")
	fmt.Println(hoge)
}
