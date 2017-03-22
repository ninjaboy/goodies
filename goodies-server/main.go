package main

import (
	"bufio"
	"fmt"
	"goodies/goodies"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request, server goodies.CommandServer) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("Cannot read incoming request")
	}
	fmt.Println(string(data))
	w.Write(server.Serve(data))
}

func formatError(err string) string {
	return fmt.Sprintf("error %v", err)
}

func main() {
	g := goodies.NewGoodies(1*time.Minute, "./goodies.dat", 30*time.Second)
	cp := goodies.NewGoodiesCommandsProcessor(g)
	ser := goodies.JsonRequestResponseSerialiser{}
	server := goodies.HttpServer{cp, ser}

	http.HandleFunc("/goodies", func(w http.ResponseWriter, r *http.Request) { handler(w, r, server) })
	http.ListenAndServe(":9006", nil) //9006 as for good

	fmt.Println("Enter any text to exit")
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Exiting...")
	g.Stop()
	<-time.After(5 * time.Second)
	fmt.Println("Bye")
}
