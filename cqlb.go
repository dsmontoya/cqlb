package cqlb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/entropyx/gocql"
	"github.com/jinzhu/inflection"
)

const (
	insertQueryTemplate = "insert into %s (%s) values(%s);"
	whereQueryTemplate  = "select %s from %s %s %s %s;"
	updateQueryTemplate = "update %s set %s where %s;"
	batchQueryTemplate  = "begin batch %s apply batch;"
)

type fieldTag struct {
	Name       string
	OmitEmpty  bool
	PrimaryKey bool
}

type Session struct {
	s              *gocql.Session
	query          interface{}
	args           []interface{}
	sel            []string
	limit          int
	allowFiltering bool
	consistency    gocql.Consistency
	value          reflect.Value
	indirect       reflect.Value
	tableName      string
	pageSize       int
	prefetch       float64
}

func SetSession(s *gocql.Session) *Session {
	return &Session{s: s}
}

func (s *Session) AllowFiltering(b bool) *Session {
	c := s.clone()
	c.allowFiltering = b
	return c
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
	//fmt.Println("fields", f)
	return s.s.Query(stmt, f["values"].([]interface{})...).Exec()
}

func (s *Session) Update(v interface{}) error {
	f := fieldsUpdate(v)
	stmt := updateQuery(f)
	// fmt.Println("fields: ", f)
	// fmt.Println("stmt: ", stmt)
	return s.s.Query(stmt).Exec()
}

func (s *Session) Batch(stmts string) error {
	stmt := batchQuery(stmts)
	// fmt.Println("fields: ", f)
	// fmt.Println("stmt: ", stmt)
	return s.s.Query(stmt).Exec()
}

func (s *Session) GetUpdateStmt(v interface{}) string {
	f := fieldsUpdate(v)
	stmt := updateQuery(f)
	return stmt
}

func (s *Session) GetInsertStmt(v interface{}) string {
	f := fieldsInsert(v)
	stmt := insertQuery(f)
	return stmt
}

func (s *Session) Iter(value interface{}) *gocql.Iter {
	c := s.clone()
	var fields map[string]interface{}
	v := reflect.ValueOf(value)
	c.setModel(v)
	query := c.query
	args := c.args
	vq := reflect.ValueOf(query)
	kindQuery := vq.Kind()
	switch kindQuery {
	case reflect.Map:
		fields = whereFieldsFromMap(query)
	case reflect.Struct:
		fields = whereFieldsFromCondList(query, args)
	}
	values := fields["values"].([]interface{})
	wQuery := c.whereQuery(fields)
	fmt.Printf("[%s] %s\n", time.Now(), wQuery)
	q := c.s.Query(wQuery, values...)
	if consistency := c.consistency; consistency > 0 {
		q = q.Consistency(consistency)
	}
	q = q.PageSize(c.pageSize)
	if prefetch := c.prefetch; prefetch > 0 {
		q.Prefetch(prefetch)
	}
	//fmt.Println("num: ", q.Iter().NumRows())
	return q.Iter()
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

func (s *Session) PageSize(n int) *Session {
	c := s.clone()
	c.pageSize = n
	return c
}

func (s *Session) Prefetch(p float64) *Session {
	c := s.clone()
	c.prefetch = p
	return c
}

// func (s *Session) Scan(value interface{}) bool {
// 	var fields map[string]interface{}
// 	v := reflect.ValueOf(value)
// 	s.setModel(v)
// 	query := s.query
// 	vq := reflect.ValueOf(query)
// 	kindQuery := vq.Kind()
// 	switch kindQuery {
// 	case reflect.Map:
// 		fields = whereFieldsFromMap(query)
// 	}
// 	values := fields["values"].([]interface{})
// 	q := s.s.Query(s.whereQuery(fields), values...)
// 	if consistency := s.consistency; consistency > 0 {
// 		q = q.Consistency(consistency)
// 	}
// 	b := cqlr.BindQuery(q)
// 	return b.Scan(value)
// }

func (s *Session) Select(sel ...string) *Session {
	c := s.clone()
	c.sel = append(c.sel, sel...)
	return c
}

func (s *Session) Table(name string) *Session {
	c := s.clone()
	c.tableName = name
	return c
}

func (s *Session) Token(sel ...string) *Session {
	c := s.clone()
	var tokenFields []string
	for i := 0; i < len(sel); i++ {
		s := sel[i]
		tokenFields = append(sel, fmt.Sprintf("token(%s)", s))
	}
	c.sel = append(c.sel, tokenFields...)
	return c
}

func (s *Session) Where(args ...interface{}) *Session {
	ns := s.clone()
	if len(args) <= 0 {
		return ns
	}
	ns.query = args[0]
	ns.args = args[1:len(args)]
	return ns
}

func (s *Session) allowFilteringString() string {
	if s.allowFiltering == true {
		return "ALLOW FILTERING"
	}
	return ""
}

func (s *Session) clone() *Session {
	ns := *s
	return &ns
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

func (s *Session) whereQuery(f map[string]interface{}) string {
	var conditionsString string
	sel := s.selectString()
	limit := s.limitString()
	allowFiltering := s.allowFilteringString()
	if conditions := f["conditions"].(string); conditions != "" {
		conditionsString = fmt.Sprintf("WHERE %v", conditions)
	}
	query := fmt.Sprintf(whereQueryTemplate, sel, s.tableName, conditionsString, limit, allowFiltering)
	return query
}

func compile(v interface{}, cols []gocql.ColumnInfo) error {
	return nil
}

func contentOfSlice(v reflect.Value) []interface{} {
	slice := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		f := reflect.Indirect(v.Index(i))
		slice[i] = f.Interface()
	}
	return slice
}

func fields(v interface{}) map[string]interface{} {
	var names string
	var slots string
	var values []interface{}
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
	return result
}

func fieldsInsert(v interface{}) map[string]interface{} {
	var names string
	var slots string
	var values []interface{}
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
		if fvIndirect.Kind() == reflect.Map {
			slots += formatMap(fvIndirect)
		} else {
			slots += getValue(fvIndirect)
		}
		strategies[tagName] = inf
		values = append(values, inf)
	}
	result["names"] = names
	result["values"] = values
	result["slots"] = slots
	return result
}

func fieldsUpdate(v interface{}) map[string]interface{} {
	var columnsUpdate string
	var keysUpdate string
	var values []interface{}
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
		if tag.Name != "" {
			tagName = tag.Name
		} else {
			tagName = strings.ToLower(f.Name)
		}
		if tag.PrimaryKey {
			if keysUpdate != "" {
				keysUpdate += " and "
			}
			//keysUpdate += tagName + " = ?"
			keysUpdate += tagName + " = " + getValue(fvIndirect)
		} else {
			if columnsUpdate != "" {
				columnsUpdate += ", "
			}
			if fvIndirect.Kind() == reflect.Map {
				columnsUpdate += listMap(tagName, fvIndirect)
			} else {
				//columnsUpdate += tagName + " = ?"
				columnsUpdate += tagName + " = " + getValue(fvIndirect)
			}
		}
		values = append(values, inf)
	}
	result["columns"] = columnsUpdate
	result["keys"] = keysUpdate
	result["values"] = values

	return result
}

func insertQuery(f map[string]interface{}) string {
	query := fmt.Sprintf(insertQueryTemplate, f["table_name"], f["names"], f["slots"])
	return query
}

func updateQuery(f map[string]interface{}) string {
	query := fmt.Sprintf(updateQueryTemplate, f["table_name"], f["columns"], f["keys"])
	return query
}

func batchQuery(s string) string {
	query := fmt.Sprintf(batchQueryTemplate, s)
	return query
}

func isZero(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func listMap(nameMap string, fv reflect.Value) string {
	keys := fv.MapKeys()
	var columnsMap string
	for i := 0; i < len(keys); i++ {
		if columnsMap != "" {
			columnsMap += ", "
		}
		columnsMap += nameMap + "[" + getValue(keys[i]) + "] = " + getValue(fv.MapIndex(keys[i]))
	}

	return columnsMap
}

func formatMap(fv reflect.Value) string {
	keys := fv.MapKeys()
	var columnsMap string
	for i := 0; i < len(keys); i++ {
		if columnsMap != "" {
			columnsMap += ", "
		}
		columnsMap += getValue(keys[i]) + ": " + getValue(fv.MapIndex(keys[i]))
	}

	columnsMap = "{" + columnsMap + "}"

	return columnsMap
}

func getValue(fv reflect.Value) string {
	var value string
	//fmt.Println(fv.Type().String())
	switch fv.Type().String() {
	case "string":
		value = "'" + fv.String() + "'"
	case "gocql.UUID":
		value = fv.Interface().(gocql.UUID).String()
	case "int", "int8", "int16", "int32", "int64":
		value = strconv.FormatInt(fv.Int(), 10)
	case "float32", "float64":
		value = strconv.FormatFloat(fv.Float(), 'f', -1, 64)
	case "bool":
		value = strconv.FormatBool(fv.Bool())
	default:
		fmt.Println("Invalid type")
		value = "error"
	}
	return value
}

func tag(f reflect.StructField) *fieldTag {
	ft := &fieldTag{}
	tag := f.Tag.Get("cql")
	opts := strings.Split(tag, ",")
	ft.Name = opts[0]
	if len(opts) > 1 && opts[1] == "omitempty" {
		ft.OmitEmpty = true
	}
	if len(opts) > 1 && opts[1] == "primary_key" {
		ft.PrimaryKey = true
	}
	return ft
}

func operatorForValue(value reflect.Value) string {
	kind := value.Kind().String()
	fmt.Println(kind)
	return ""
}

func whereFieldsFromCondList(query interface{}, args []interface{}) map[string]interface{} {
	var conditions string
	var values []interface{}
	var names []string
	result := make(map[string]interface{})
	condList := append(args, query)
	for i := 0; i < len(condList); i++ {
		condition := condList[i].(Condition)
		if i != 0 {
			conditions += " AND "
		}
		conditions += condition.String()

		names = append(names, condition.key)
		values = append(values, condition.values...)
	}
	result["conditions"] = conditions
	result["values"] = values
	result["names"] = names

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
		value := v.MapIndex(key)
		inf := value.Interface()
		if i != 0 {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("%s = ?", keyString)
		names = append(names, keyString)
		values = append(values, inf)
	}
	result["conditions"] = conditions
	result["values"] = values
	result["names"] = names
	fmt.Println("duration whereFieldsFromMap", time.Since(t1))
	return result
}
