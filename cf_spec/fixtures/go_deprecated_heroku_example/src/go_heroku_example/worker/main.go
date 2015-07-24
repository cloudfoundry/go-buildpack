package main

import (
	"fmt"
	"github.com/kr/go-heroku-example/message"
	"time"
)

func main() {
	for {
		fmt.Println(message.Hello)
		time.Sleep(10 * time.Second)
	}
}
