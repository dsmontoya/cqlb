// CREATE KEYSPACE test WITH REPLICATION = { 'class' : 'NetworkTopologyStrategy', 'dc1' : 1, 'datacenter1': 1 };
//
// CREATE TYPE address (
// 	street TEXT,
// 	number TEXT,
// 	city TEXT,
// 	country TEXT
// );
//
// CREATE TABLE users (
// 	name TEXT,
// 	password TEXT,
// 	email_addresses list<TEXT>,
// 	phones map<TEXT, TEXT>,
// 	addresses list<frozen <address>>,
// 	PRIMARY KEY (name)
// );

package cqlb

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/gocql/gocql"
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
	Convey("Given a session", t, func() {
		cluster := gocql.NewCluster(os.Getenv("CASSANDRA_HOST"))
		cluster.Keyspace = "test"
		cluster.Consistency = gocql.Any
		session, _ := cluster.CreateSession()
		s := SetSession(session)
		Convey("When the session is cloned", func() {
			ns := s.clone()
			Convey("And a model is set to the new session", func() {
				ns.Model(&User{})

				Convey("The table name should be empty", func() {
					So(ns.tableName, ShouldBeBlank)
				})
			})
		})

		Convey("When a User is set", func() {
			ns := s.Model(&User{})
			Convey("The table name should be 'users'", func() {
				So(ns.tableName, ShouldEqual, "users")
			})
		})

		Convey("Given a User", func() {
			user := &User{}
			Convey("When it is inserted", func() {
				err := s.Insert(user)

				Convey("err should be nil", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

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
