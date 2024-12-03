package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	pythonModuleLog, err := exec.Command("python", "-c", "import ujson; print(ujson.dumps([{'hello': 'world'}]))").Output()
	if err != nil {
		log.Print("ERROR:", err)
		fmt.Fprintf(res, "ERROR: %v\n", err)
	} else {
		fmt.Fprintf(res, "The python script evaluated to: %s\n", pythonModuleLog)
	}
}
