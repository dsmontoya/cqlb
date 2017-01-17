package cqlb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/jinzhu/inflection"
	"github.com/relops/cqlr"
)

const (
	insertQueryTemplate = "insert into %s (%s) values(%s);"
	whereQueryTemplate  = "select %s from %s %s %s %s;"
)

type fieldTag struct {
	Name      string
	OmitEmpty bool
}

type Session struct {
	s           *gocql.Session
	query       interface{}
	args        []interface{}
	sel         []string
	limit       int
	consistency gocql.Consistency
	value       reflect.Value
	indirect    reflect.Value
	tableName   string
}

func SetSession(s *gocql.Session) *Session {
	return &Session{s: s}
}

func (s *Session) Consistency(consistency gocql.Consistency) *Session {
	c := s.clone()
	c.consistency = consistency
	return c
}

func (s *Session) Find(slice interface{}) error {
	// var fields map[string]interface{}
	// var fieldsToScan []interface{}
	// value := reflect.ValueOf(slice)
	// k := value.Kind()
	// if k != reflect.Slice {
	// 	return errors.New("value should be a slice.")
	// }
	// v := value.Index(0)
	// indirect := reflect.Indirect(v)
	// s.setModel(v)
	// query := s.query
	// vq := reflect.ValueOf(query)
	// kindQuery := vq.Kind()
	// switch kindQuery {
	// case reflect.Map:
	// 	fields = whereFieldsFromMap(query)
	// }
	// iter := s.s.Query(s.whereQuery(fields), values).Iter()
	// cols := iter.Columns()
	// values := make([]interface{}, len(cols))
	// names := f["names"].([]string)
	// for i, col := range cols {
	// 	values[i] = f["strategies"].(map[string]interface{})[col.Name]
	// }
	return nil
}

func (s *Session) Insert(v interface{}) error {
	f := fields(v)
	stmt := insertQuery(f)
	fmt.Println("fields", f)
	return s.s.Query(stmt, f["values"].([]interface{})...).Exec()
}

func (s *Session) Iter(value interface{}) *gocql.Iter {
	c := s.clone()
	var fields map[string]interface{}
	v := reflect.ValueOf(value)
	c.setModel(v)
	query := c.query
	vq := reflect.ValueOf(query)
	kindQuery := vq.Kind()
	switch kindQuery {
	case reflect.Map:
		fields = whereFieldsFromMap(query)
	}
	values := fields["values"].([]interface{})
	q := c.s.Query(c.whereQuery(fields), values...)
	if consistency := c.consistency; consistency > 0 {
		q = q.Consistency(consistency)
	}
	return q.Iter()
}

func (s *Session) Where(query interface{}, args ...interface{}) *Session {
	ns := s.clone()
	ns.query = query
	ns.args = args
	return ns
}

func (s *Session) Limit(limit int) *Session {
	c := s.clone()
	c.limit = limit
	return c
}

func (s *Session) Model(value interface{}) *Session {
	v := reflect.ValueOf(value)
	ns := s.clone()
	ns.setModel(v)
	return ns
}

func (s *Session) Scan(value interface{}) bool {
	var fields map[string]interface{}
	v := reflect.ValueOf(value)
	s.setModel(v)
	query := s.query
	vq := reflect.ValueOf(query)
	kindQuery := vq.Kind()
	switch kindQuery {
	case reflect.Map:
		fields = whereFieldsFromMap(query)
	}
	values := fields["values"].([]interface{})
	q := s.s.Query(s.whereQuery(fields), values...)
	if consistency := s.consistency; consistency > 0 {
		q = q.Consistency(consistency)
	}
	b := cqlr.BindQuery(q)
	return b.Scan(value)
}

func (s *Session) Select(sel ...string) *Session {
	c := s.clone()
	c.sel = sel
	return c
}

func (s *Session) Table(name string) *Session {
	c := s.clone()
	c.tableName = name
	return c
}

func (s *Session) limitString() string {
	if limit := s.limit; limit > 0 {
		return fmt.Sprintf("LIMIT %v", limit)
	}
	return ""
}

func (s *Session) selectString() string {
	if sel := s.sel; len(sel) > 0 {
		return strings.Join(sel, ",")
	}
	return "*"
}

func (s *Session) setModel(v reflect.Value) {
	indirect := reflect.Indirect(v)
	t := indirect.Type()
	s.value = v
	s.indirect = indirect
	if s.tableName == "" {
		s.tableName = inflection.Plural(strings.ToLower(t.Name()))
	}
}

func (s *Session) clone() *Session {
	ns := *s
	return &ns
}

func (s *Session) whereQuery(f map[string]interface{}) string {
	var conditionsString string
	sel := s.selectString()
	limit := s.limitString()
	if conditions := f["conditions"].(string); conditions != "" {
		conditionsString = fmt.Sprintf("WHERE %v", conditions)
	}
	query := fmt.Sprintf(whereQueryTemplate, sel, s.tableName, conditionsString, limit, "")
	return query
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
	if len(opts) > 1 && opts[1] == "omitempty" {
		ft.OmitEmpty = true
	}
	return ft
}

func fields(v interface{}) map[string]interface{} {
	var names string
	var slots string
	var values []interface{}
	t1 := time.Now()
	strategies := make(map[string]interface{})
	result := make(map[string]interface{})
	value := reflect.ValueOf(v)
	indirect := reflect.Indirect(value)
	t := indirect.Type()
	result["table_name"] = inflection.Plural(strings.ToLower(t.Name()))
	for i := 0; i < t.NumField(); i++ {
		var inf interface{}
		var tagName string
		f := t.Field(i)
		fv := indirect.Field(i)
		tag := tag(f)
		fvIndirect := reflect.Indirect(fv)
		if fvIndirect.IsValid() == false {
			continue
		}
		inf = fvIndirect.Interface()
		isZero := isZero(inf)
		if isZero == true && tag.OmitEmpty == true {
			continue
		}
		if i != 0 {
			names += ","
			slots += ","
		}
		if tag.Name != "" {
			tagName = tag.Name
		} else {
			tagName = strings.ToLower(f.Name)
		}
		names += tagName
		slots += "?"
		strategies[tagName] = inf
		values = append(values, inf)
	}
	result["names"] = names
	result["values"] = values
	result["slots"] = slots
	fmt.Println("duration cqlb", time.Since(t1))
	return result
}

func whereFieldsFromMap(value interface{}) map[string]interface{} {
	var conditions string
	var values []interface{}
	var names []string
	t1 := time.Now()
	result := make(map[string]interface{})
	v := reflect.ValueOf(value)
	keys := v.MapKeys()
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		keyString := key.String()
		value := v.MapIndex(key).Interface()
		if i != 0 {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("%s = ?", keyString)
		names = append(names, keyString)
		values = append(values, value)
	}
	result["conditions"] = conditions
	result["values"] = values
	result["names"] = names
	fmt.Println("duration whereFieldsFromMap", time.Since(t1))
	return result
}

func isZero(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
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
