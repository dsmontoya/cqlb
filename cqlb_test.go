package cqlb

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type User struct {
	Test           []*string
	Name           string `cql:"name"`
	Password       string `cql:"password"`
	EmailAddresses []string
	Phones         map[string]string `cql:"phones"`
	Addresses      []*Address        `cql:"addresses"`
}

type Address struct {
	Street  string
	Number  string
	City    string
	Country string
}

func TestCQLM(t *testing.T) {
	Convey("Given a user", t, func() {
		str := "lol"
		user := &User{
			Test:     []*string{&str},
			Name:     "Jhon",
			Password: "super-secret-password",
			EmailAddresses: []string{
				"me@jhon.me",
				"jhon@gmail.com",
			},
			Phones: map[string]string{
				"Home": "12345",
				"Work": "54321",
			},
			Addresses: []*Address{
				&Address{
					Street:  "London Street",
					Number:  "501",
					City:    "LA",
					Country: "USA",
				},
				&Address{
					Street:  "Reforma",
					Number:  "222",
					City:    "Mexico City",
					Country: "Mexico",
				},
			},
		}

		Convey("When the model is compiled", func() {
			f := fields(user)
			fmt.Println("fields", f)
		})
	})

	Convey("When you get the content of an array of strings", t, func() {
		slice := []string{
			"one",
			"two",
			"three",
		}
		v := reflect.ValueOf(slice)
		content := contentOfSlice(v)
		Convey("content should have length 3 ", func() {
			So(content, ShouldHaveLength, 3)
		})
	})

	Convey("When you get the content of an array of integers", t, func() {
		slice := []int{
			1,
			2,
			3,
		}
		v := reflect.ValueOf(slice)
		content := contentOfSlice(v)
		Convey("content should have length 3 ", func() {
			So(content, ShouldHaveLength, 3)
		})
	})

	Convey("When you get the content of an array of non-struct pointers", t, func() {
		var str = " a string"
		slice := []*string{
			&str,
			&str,
			&str,
		}
		v := reflect.ValueOf(slice)
		content := contentOfSlice(v)

		Convey("content should have length 3 ", func() {
			So(content, ShouldHaveLength, 3)
		})

		Convey("content should be an array of strings (non-pointer)", func() {
			So(content[0], ShouldHaveSameTypeAs, str)
		})
	})

	Convey("When you get the content of an array of pointers to structs", t, func() {
		slice := []*Address{
			&Address{},
			&Address{},
			&Address{},
		}
		v := reflect.ValueOf(slice)
		content := contentOfSlice(v)

		Convey("content should have length 3 ", func() {
			So(content, ShouldHaveLength, 3)
		})

		Convey("content should be an array of map[string]interface{}", func() {
			So(content[0], ShouldHaveSameTypeAs, map[string]interface{}{})
		})
	})

}
