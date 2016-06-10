package main

import (
	"fmt"
	"goodies/goodies"
)

func main() {
	v := goodies.NewGoodies()
	fmt.Println("Hello")
	v.Set("test", "Hello world")
	fmt.Println(v.Get("test"))
	fmt.Println("Bye")
}
