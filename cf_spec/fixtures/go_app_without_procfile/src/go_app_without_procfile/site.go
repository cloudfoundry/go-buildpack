package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
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
