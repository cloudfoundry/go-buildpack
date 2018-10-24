package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/", printDotNet)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func printDotNet(res http.ResponseWriter, req *http.Request) {
	out, err := exec.Command("dotnet", "--version").Output()

	if err != nil {
		fmt.Fprintln(res, "error:", err)
		return
	}

	fmt.Fprintln(res, "dotnet:", string(out))
}
