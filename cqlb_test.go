// CREATE KEYSPACE test WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
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
// 	age INT,
// 	sex TINYINT,
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
	Password       string            `cql:"password,omitempty"`
	Age            int               `cql:"age,omitempty"`
	Sex            int8              `cql:"sex,omitempty"`
	EmailAddresses []string          `cql:"email_addresses,omitempty"`
	Phones         map[string]string `cql:"phones,omitempty"`
	Addresses      []Address         `cql:"addresses,omitempty"`
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
		session, err := cluster.CreateSession()
		fmt.Println(err, os.Getenv("CASSANDRA_HOST"))
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

		Convey("Given a User", func() {
			user := &User{Name: "jhon", Password: "lol"}
			Convey("When it is inserted", func() {
				err := s.Insert(user)

				Convey("err should be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("The new user should exist", func() {
					u := &User{}
					s.Where(map[string]interface{}{"name": "jhon"}).Scan(u)
					So(user.Name, ShouldNotBeEmpty)
				})
			})

			Convey("When it is set as model", func() {
				ns := s.Model(user)
				Convey("The table name should be 'users'", func() {
					So(ns.tableName, ShouldEqual, "users")
				})
			})
		})
	})

	Convey("Given a user", t, func() {
		user := &User{
			Name:     "Jhon",
			Password: "super-secret-password",
			Addresses: []Address{
				Address{
					Street: "London",
				},
			},
		}

		Convey("When the model is compiled", func() {
			f := fields(user)
			fmt.Println("fields", f)

			Convey("The table name should be users", func() {
				So(f["table_name"], ShouldEqual, "users")
			})

			Convey("The slots should be equal to '?,?,?'", func() {
				So(f["slots"], ShouldEqual, "?,?,?")
			})

			Convey("The names should be equal 'name,password,addresses'", func() {
				So(f["names"], ShouldEqual, "name,password,addresses")
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

func TestIsZero(t *testing.T) {
	Convey("Integer:", t, func() {
		Convey("0 should be zero", func() {
			So(isZero(0), ShouldBeTrue)
		})
		Convey("1 should not be zero", func() {
			So(isZero(1), ShouldBeFalse)
		})
	})
	Convey("String:", t, func() {
		Convey("'' should be zero", func() {
			So(isZero(""), ShouldBeTrue)
		})
		Convey("'string' should not be zero", func() {
			So(isZero("string"), ShouldBeFalse)
		})
	})
	Convey("Map:", t, func() {
		user := &User{}
		Convey("an empty map should be zero", func() {
			So(isZero(user.Phones), ShouldBeTrue)
		})
		Convey("a non empty map should not be zero", func() {
			So(isZero(map[string]string{"key": "value"}), ShouldBeFalse)
		})
	})
}
