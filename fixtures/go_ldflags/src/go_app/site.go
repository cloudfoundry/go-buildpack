package main

import (
	"fmt"
	"net/http"
	"os"
)

var (
	linker_flag       string
	other_linker_flag string
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
	fmt.Fprintf(res, "linker_flag=%s\n", linker_flag)
	fmt.Fprintf(res, "other_linker_flag=%s\n", other_linker_flag)
	fmt.Fprintln(res, "done")
}
