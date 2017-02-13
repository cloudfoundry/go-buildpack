package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", gopath)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func gopath(res http.ResponseWriter, req *http.Request) {

	go_path := os.Getenv("GOPATH")

	if len(go_path) == 0 {
		fmt.Fprintln(res, "GOPATH: not defined")
	} else {
		fmt.Fprintf(res, "GOPATH: %s\n", go_path)
	}
}
