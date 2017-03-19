package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"goodies/goodies-server"
)

func main() {
	fmt.Println("Connecting to 127.0.0.1:9006")
	
	json.Marshal()
	res, err := http.Post("http://127.0.0.1:9006/goodies",
		"application/json",
		bytes.NewReader())
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", robots)
}
