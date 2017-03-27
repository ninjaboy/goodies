package main

import (
	"bufio"
	"fmt"
	"goodies/goodies"
	"os"
	"time"
)

func formatError(err string) string {
	return fmt.Sprintf("error %v", err)
}

func main() {
	server := goodies.NewGoodiesHttpServer("9006", 1*time.Minute, "./goodies.dat", 30*time.Second)
	fmt.Println("Listening on: 0.0.0.0:9006")
	go server.ListenAndServe()

	fmt.Println("Enter any text to exit")
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Exiting...")
	server.Shutdown(nil)
	<-time.After(5 * time.Second)
	fmt.Println("Bye")
}
