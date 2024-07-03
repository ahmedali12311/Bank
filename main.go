package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store storage, fname, lname, pw string) *Account {
	acc, err := NewAccount(fname, lname, pw)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.createAccount(acc); err != nil {
		log.Fatal(err)
	}
	return acc
}

func seedAccounts(s storage) {
	seedAccount(s, "Ahmed", "GG", "Hunter69")
}
func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := NewPostGresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}
	fmt.Println("starting server at port 8080")

	server := NewApiServer(":8080", store)
	server.Run()
}
