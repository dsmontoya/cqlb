package cqlb

import (
	"reflect"
	"strings"

	"github.com/gocql/gocql"
)

func compile(v interface{}, cols []gocql.ColumnInfo) error {
	return nil
}

func fields(v interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	value := reflect.ValueOf(v)
	indirect := reflect.Indirect(value)
	t := indirect.Type()
	for i := 0; i < t.NumField(); i++ {
		var inf interface{}
		fv := indirect.Field(i)
		kind := fv.Kind()
		if kind.String() == "slice" {
			inf = contentOfSlice(fv)
		}
		f := t.Field(i)
		tag := f.Tag.Get("cql")
		if inf == nil {
			inf = indirect.Field(i).Interface()
		}
		if tag != "" {
			result[tag] = inf
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
		f := v.Index(i)
		if f.Kind().String() != "ptr" && f.Kind().String() != "struct" {
			slice[i] = f.Interface()
			continue
		}
		elem := f.Elem()
		if elem.Kind() != reflect.Struct {
			slice[i] = elem.Interface()
			continue
		}
		nestedFields := fields(f.Interface())
		slice[i] = nestedFields
	}
	return slice
}

func getType(v interface{}) {

}
