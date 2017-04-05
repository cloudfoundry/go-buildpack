package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/", go_version)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func go_version(res http.ResponseWriter, req *http.Request) {

	go_version, err := exec.Command("go", "version").Output()

	if err != nil {
		fmt.Fprintf(res, "go toolchain not found, error: %s", err.Error())
	} else {
		fmt.Fprintf(res, "%s", go_version)
	}
}
