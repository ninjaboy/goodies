package main

import (
	"bufio"
	"flag"
	"fmt"
	"goodies/goodies"
	"net"
	"time"
)

var debug bool

func main() {

	var port string
	var ttl int
	var storagePath string
	var storageInterval int

	flag.StringVar(&port, "p", "19006", "port to be listening on")
	flag.StringVar(&storagePath, "f", "./bin/goodies.dat", "path to a storage file")
	flag.IntVar(&ttl, "ttl", 30, "item default ttl (seconds)")
	flag.IntVar(&storageInterval, "stor", 30, "storage default timeout (seconds)")
	flag.BoolVar(&debug, "debug", false, "prints debug information")

	flag.Parse()

	g := goodies.NewGoodiesPersistedStorage(time.Second*time.Duration(ttl),
		storagePath,
		time.Duration(storageInterval)*time.Second)
	defer g.Stop()
	commandProcessor := goodies.NewGoodiesCommandsProcessor(g)
	serialiser := goodies.JsonRequestResponseSerialiser{}

	log("Starting server on port", ":"+port)
	if debug {
		fmt.Println()
	}
	ln, _ := net.Listen("tcp", ":"+port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error occured while listening", err)
			return
		}
		go handleConnection(conn, serialiser, commandProcessor)
	}
}

func log(lines ...interface{}) {
	if debug {
		fmt.Println(lines...)
	}
}

func handleConnection(conn net.Conn, serialiser goodies.RequestResponseSerialiser, cp goodies.CommandProcessor) {
	defer conn.Close()
	for {
		message, _ := bufio.NewReader(conn).ReadBytes('\n')
		if len(message) == 0 {
			log("Closing connection")
			return
		}
		log("Message Received:", string(message))
		req := &goodies.CommandRequest{}
		serialiser.DeserialiseRequest([]byte(message), req)

		resp := cp.Process(*req)
		data, err := serialiser.SerialiseResponse(resp)
		data = append(data, '\n')
		if err != nil {
			log("Error occured during request processing", err.Error())
		} else {
			log("Response:", string(data))
			conn.Write(data)
		}
	}
}
