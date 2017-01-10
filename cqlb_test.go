package cqlb

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type User struct {
	Name           string            `cql:"name,omitempty"`
	Password       string            `cql:"password"`
	EmailAddresses []string          `cql:"email_addresses"`
	Phones         map[string]string `cql:"phones"`
	Addresses      []*Address        `cql:"addresses"`
}

type Address struct {
	Street  string `cql:"street"`
	Number  string `cql:"number"`
	City    string `cql:"city"`
	Country string `cql:"country"`
}

func TestCQLM(t *testing.T) {
	Convey("Given a user", t, func() {
		user := &User{
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

			Convey("The table name should be users", func() {
				So(f["table_name"], ShouldEqual, "users")
			})

			Convey("The slots should be equal to '?,?,?,?,?'", func() {
				So(f["slots"], ShouldEqual, "?,?,?,?,?")
			})

			Convey("The names should be equal 'name,password,email_addresses,phones,addresses'", func() {
				So(f["names"], ShouldEqual, "name,password,email_addresses,phones,addresses")
			})

			Convey("The insert query", func() {
				query := insertQuery(f)
				fmt.Println(query)

				Convey("should contain the fields", func() {
					So(query, ShouldEqual, fmt.Sprintf(insertQueryTemplate, f["table_name"], f["names"], f["slots"]))
				})

				Convey("should end with ;", func() {
					So(query, ShouldEndWith, ";")
				})
			})
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

		Convey("content should be an array of Address", func() {
			So(content[0], ShouldHaveSameTypeAs, Address{})
		})
	})

}
