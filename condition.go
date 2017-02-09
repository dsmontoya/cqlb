package cqlb

import "fmt"

const (
	equal = iota
	in
	greaterThan
	greaterThanOrEqualTo
	lessThan
	lessThanOrEqualTo
)

type Condition struct {
	op     int
	key    string
	values []interface{}
}

func Eq(key string, value interface{}) Condition {
	return setCondition(equal, key, value)
}

func GT(key string, value interface{}) Condition {
	return setCondition(greaterThan, key, value)
}

func GTE(key string, value interface{}) Condition {
	return setCondition(greaterThanOrEqualTo, key, value)
}

func In(key string, values ...interface{}) Condition {
	return setCondition(in, key, values)
}

func LT(key string, value interface{}) Condition {
	return setCondition(lessThan, key, value)
}

func LTE(key string, value interface{}) Condition {
	return setCondition(lessThanOrEqualTo, key, value)
}

func (c *Condition) String() string {
	var s string
	var opStr string
	key := c.key
	switch c.op {
	case equal:
		opStr = "= ?"
	case in:
		opStr = fmt.Sprintf("IN (%s)", inOpSlots(c.values))
	case greaterThan:
		opStr = "> ?"
	case greaterThanOrEqualTo:
		opStr = ">= ?"
	case lessThan:
		opStr = "< ?"
	case lessThanOrEqualTo:
		opStr = "<= ?"
	default:
		return ""
	}
	s = fmt.Sprintf("%s %s", key, opStr)
	return s
}

func inOpSlots(values []interface{}) string {
	var s string
	l := len(values)
	for i := 0; i < l; i++ {
		s += "?"
		if i+1 != l {
			s += ","
		}
	}
	return s
}

func setCondition(op int, key string, values ...interface{}) Condition {
	return Condition{
		op:     op,
		key:    key,
		values: values,
	}
}
