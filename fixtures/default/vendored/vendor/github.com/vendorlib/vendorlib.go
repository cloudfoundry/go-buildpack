package vendorlib

import "fmt"

var A = 1

func init() {
	fmt.Println("Init: a.A ==", A)
}
