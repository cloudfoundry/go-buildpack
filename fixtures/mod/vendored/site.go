package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/satori/go.uuid"
)

type Config struct {
	Age int
}

func main() {
	var conf Config
	if _, err := toml.Decode("whatever", &conf); err != nil {
		// handle error
		fmt.Println("Unhandled error")
	}

	u2 := uuid.NewV4()
	fmt.Printf("UUIDv4: %s\n", u2)

	http.HandleFunc("/", hello)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "go, world")

	url := os.Getenv("DATABASE_URL")

	fmt.Fprintln(res, ".")
	fmt.Fprintln(res, url)
	fmt.Fprintln(res, ".")
}
