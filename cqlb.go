package cqlb

import (
	"fmt"
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
	fmt.Println(result)
	for i := 0; i < t.NumField(); i++ {
		var inf interface{}
		fv := indirect.Field(i)
		fmt.Println("type", fv.Type())
		kind := fv.Kind()
		fmt.Println("kind", kind)
		if kind.String() == "slice" {
			slice := make([]interface{}, fv.Len())
			for j := 0; j < fv.Len(); j++ {
				//ifv := reflect.Indirect(fv)
				f := fv.Index(j)
				if f.Kind().String() != "ptr" || f.Kind().String() != "struct" {
					slice = append(slice, f)
					continue
				}
				fmt.Println("kind 2", fv.Index(j).Kind().String())
				fmt.Println("elem", f.Elem())
				if f.Elem().Kind() != reflect.Struct {
					slice = append(slice, f)
					continue
				}

				nestedFields := fields(f.Interface())
				slice = append(slice, nestedFields)
				fmt.Println("nestedFields", nestedFields)
			}
			inf = slice
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

func getType(v interface{}) {

}
