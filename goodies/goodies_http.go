package goodies

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpTransport struct {
	address    string
	serializer RequestResponseSerialiser
	client     http.Client
}

type GoodiesHttpServer struct {
	commandProcessor CommandProcesser
	serializer       RequestResponseSerialiser
}

func NewGoodiesClient(address string) Provider {
	return Client{NewGoodiesHttpTransport(address)}
}

func NewGoodiesHttpTransport(address string) HttpTransport {
	client := http.Client{}
	ser := JsonRequestResponseSerialiser{}
	return HttpTransport{address, ser, client}
}

func NewGoodiesHttpServer(port string, defTtl time.Duration, storage string, storageDump time.Duration) *http.Server {
	g := NewGoodies(defTtl, storage, storageDump)
	cp := NewGoodiesCommandsProcessor(g)
	serialiser := JsonRequestResponseSerialiser{}
	result := GoodiesHttpServer{cp, serialiser}

	server := &http.Server{Addr: ":" + port, Handler: &result}

	return server
}

func (s *GoodiesHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("Cannot read incoming request")
	}
	//fmt.Println(string(data))
	w.Write(s.Serve(data))
}

func (tr HttpTransport) Process(req GoodiesRequest, res *GoodiesResponse) error {
	data, err := tr.serializer.serialiseRequest(req)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequest("POST", tr.address, bytes.NewReader(data))
	httpRequest.Header.Set("Content-Type", "application/json")

	resp, err := tr.client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ErrInternalError{fmt.Sprintf("Conectivity issue: %v", resp.Status)}
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return tr.serializer.deserialiseResponse(body, res)
}

func (tr GoodiesHttpServer) Serve(reqData []byte) []byte {
	var req GoodiesRequest
	var res GoodiesResponse
	err := tr.serializer.deserialiseRequest(reqData, &req)
	if err != nil {
		res = GoodiesResponse{false, "", err}
	} else {
		res = tr.commandProcessor.HandleCommand(req)
	}

	data, err := tr.serializer.serialiseResponse(res)
	if err != nil {
		panic("Cannot serialise response")
	}
	return data
}
