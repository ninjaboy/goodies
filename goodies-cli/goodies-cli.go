package main

import (
	"bufio"
	"fmt"
	"goodies/goodies"
	"os"
	"strings"
)

type TextOverHttpClient struct {
	isConnected bool
	connString  string
	transport   goodies.HttpTransport
}

func (t *TextOverHttpClient) Connect(address string) error {
	t.connString = address //TODO: add url validation
	t.isConnected = true

	t.transport = goodies.NewGoodiesHttpTransport(address)
	return nil
}

func main() {
	client := TextOverHttpClient{}
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, _ := in.ReadString('\n')
		quit := processInput(line, &client)
		if quit {
			break
		}
	}
}

func processInput(command string, client *TextOverHttpClient) bool {
	if strings.HasPrefix(command, "Quit") {
		fmt.Println("Bye!")
		return true
	}

	if strings.HasPrefix(command, "Connect ") {
		addr := command[len("Connect "):] //TODO: add validation
		fmt.Println("Connecting to ", addr)
		err := client.Connect(addr)

		if err != nil {
			fmt.Printf("Connection failed: %v", err)
		}
		return false
	}

	if !client.isConnected {
		fmt.Println("Please connect to server first (example: Connect http://servername:port/)")
		return false
	}

	com := &goodies.GoodiesRequest{}
	err := tryBuildCommand(command, com)
	fmt.Println(com.Name, com.Parameters)
	if err != nil {
		fmt.Printf("Command formatted incorrectly: %v\n", err)
	}

	res := &goodies.GoodiesResponse{}
	err = client.transport.Process(*com, res)
	if err != nil {
		fmt.Println("Transport error occurred:", err)
		return false
	}
	if res.Err != nil {
		fmt.Println(res.Err)
		return false
	}
	fmt.Println("Ok", res.Result)
	return false
}

func tryBuildCommand(command string, req *goodies.GoodiesRequest) error {
	fields, err := GetFieldsConsideringQuotes(command)
	if err != nil {
		return err
	}
	req.Name = fields[0]
	req.Parameters = fields[1:]
	return nil
}
