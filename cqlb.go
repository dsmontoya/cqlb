package cqlb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gocql/gocql"
	"github.com/jinzhu/inflection"
)

const (
	insertQueryTemplate = "insert into %s (%s) values(%s);"
)

type fieldTag struct {
	Name      string
	OmitEmpty bool
}

type Session struct {
	s         *gocql.Session
	query     interface{}
	args      []interface{}
	value     reflect.Value
	tableName string
}

func SetSession(s *gocql.Session) *Session {
	return &Session{s: s}
}

func (s *Session) Insert(v interface{}) error {
	f := fields(v)
	stmt := insertQuery(f)
	return s.s.Query(stmt, f["values"]).Exec()
}

func (s *Session) Where(query interface{}, args ...interface{}) *Session {
	ns := s.clone()
	ns.query = query
	ns.args = args
	return ns
}

func (s *Session) Model(v interface{}) *Session {
	ns := s.clone()
	value := reflect.ValueOf(v)
	indirect := reflect.Indirect(value)
	t := indirect.Type()
	ns.value = value
	ns.tableName = inflection.Plural(strings.ToLower(t.Name()))
	return ns
}

func (s *Session) clone() *Session {
	ns := *s
	return &ns
}

func insertQuery(f map[string]interface{}) string {
	query := fmt.Sprintf(insertQueryTemplate, f["table_name"], f["names"], f["slots"])
	return query
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
	var names string
	var slots string
	var values []interface{}
	result := make(map[string]interface{})
	value := reflect.ValueOf(v)
	indirect := reflect.Indirect(value)
	t := indirect.Type()
	result["table_name"] = inflection.Plural(strings.ToLower(t.Name()))
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
		if i != 0 {
			names += ","
			slots += ","
		}
		if tag.Name != "" {
			names += tag.Name
		} else {
			names += strings.ToLower(f.Name)
		}
		slots += "?"
		values = append(values, inf)
	}
	result["names"] = names
	result["values"] = values
	result["slots"] = slots
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
