package main

import (
	"fmt"
	"github.com/ZiCog/shiny-thing/foo"
	"net/http"
	"os"
)

func main() {
	foo.Do()
	http.HandleFunc("/", hello)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, world")
}
