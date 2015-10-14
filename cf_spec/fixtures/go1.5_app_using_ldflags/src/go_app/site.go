package main

import (
	"fmt"
	"net/http"
	"os"
)

var linker_flag string

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, linker_flag)
	fmt.Fprintln(res, "done")
}
