package cqlb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gocql/gocql"
)

type fieldTag struct {
	Name      string
	OmitEmpty bool
}

func compile(v interface{}, cols []gocql.ColumnInfo) error {

	return nil
}

func tag(f reflect.StructField) *fieldTag {
	ft := &fieldTag{}
	tag := f.Tag.Get("cql")
	opts := strings.Split(tag, ",")
	ft.Name = opts[0]
	if len(opts) > 1 && opts[0] == "omitempty" {
		ft.OmitEmpty = true
	}
	return ft
}

func fields(v interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	value := reflect.ValueOf(v)
	indirect := reflect.Indirect(value)
	t := indirect.Type()
	for i := 0; i < t.NumField(); i++ {
		var inf interface{}
		f := t.Field(i)
		fv := indirect.Field(i)
		tag := tag(f)
		fmt.Println(tag)
		if fv.IsValid() == false && tag.OmitEmpty == true {
			continue
		}
		fvIndirect := reflect.Indirect(fv)
		inf = fvIndirect.Interface()
		if tag.Name != "" {
			result[tag.Name] = inf
		} else {
			//fmt.Println(f.Name, indirect.Field(f.Index[0]))
			result[strings.ToLower(f.Name)] = inf
			//b.fieldMap[strings.ToLower(f.Name)] = f.Index
		}
	}
	return result
}

func contentOfSlice(v reflect.Value) []interface{} {
	slice := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		f := reflect.Indirect(v.Index(i))
		slice[i] = f.Interface()
	}
	return slice
}

func getType(v interface{}) {

}
