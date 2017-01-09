package cqlb

import (
	"reflect"
	"strings"

	"github.com/gocql/gocql"
)

type fieldTag struct {
	Name      string
	OmitEmpty bool
}

type Session struct{}

func SetSession(*gocql.Session) *Session {
	return &Session{}
}

func (s *Session) Insert(v interface{}) {
	f := fields(v)
	insertQuery(f)
}

func insertQuery(f map[string][]interface{}) string {
	return ""
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

func fields(v interface{}) map[string][]interface{} {
	var names []interface{}
	var values []interface{}
	result := make(map[string][]interface{}, 2)
	value := reflect.ValueOf(v)
	indirect := reflect.Indirect(value)
	t := indirect.Type()
	for i := 0; i < t.NumField(); i++ {
		var inf interface{}
		f := t.Field(i)
		fv := indirect.Field(i)
		tag := tag(f)
		if fv.IsValid() == false && tag.OmitEmpty == true {
			continue
		}
		fvIndirect := reflect.Indirect(fv)
		inf = fvIndirect.Interface()
		if tag.Name != "" {
			names = append(names, tag.Name)
		} else {
			names = append(names, strings.ToLower(f.Name))
		}
		values = append(values, inf)
	}
	result["names"] = names
	result["values"] = values
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
