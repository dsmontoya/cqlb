package cqlm

import (
	"fmt"
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
			fmt.Println(f)
		})
	})
}
