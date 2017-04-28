package goodies

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
)

type TCPBasedClient struct {
	address    string
	serialiser RequestResponseSerialiser
	conn       net.Conn
	connected  bool
	lock       sync.RWMutex
}

func NewTCPBasedClient(address string) *TCPBasedClient {
	return &TCPBasedClient{address: address, serialiser: JsonRequestResponseSerialiser{}}
}

func (tr *TCPBasedClient) Process(req CommandRequest) CommandResponse {
	data, err := tr.serialiser.SerialiseRequest(req)
	data = append(data, byte('\n'))
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	if !tr.connected {
		tr.lock.Lock()
		if !tr.connected {
			conn, err := net.Dial("tcp", tr.address)
			if err != nil {
				return NewCommandResponseFromError(err)
			}
			tr.connected = true
			tr.conn = conn
		}
		tr.lock.Unlock()
	}
	n, err := fmt.Fprintf(tr.conn, string(data))
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	if n != len(data) {
		return NewCommandResponseFromError(errors.New("Failed to send a request to the remote endpoint"))
	}
	respData, err := bufio.NewReader(tr.conn).ReadBytes('\n')
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	res := &CommandResponse{}
	err = tr.serialiser.DeserialiseResponse(respData, res)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return *res
}
