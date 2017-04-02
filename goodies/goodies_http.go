package goodies

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type GoodiesHttpCommandClient struct {
	address    string
	serializer RequestResponseSerialiser
	client     http.Client
}

func NewGoodiesClient(address string) Provider {
	return goodiesClient{NewGoodiesHttpCommandClient(address)}
}

func NewGoodiesHttpCommandClient(address string) GoodiesHttpCommandClient {
	client := http.Client{}
	ser := jsonRequestResponseSerialiser{}
	return GoodiesHttpCommandClient{address, ser, client}
}

func NewGoodiesHttpServer(port string, defTtl time.Duration, storage string, persistInterval time.Duration) *http.Server {
	g := NewGoodiesPersistedStorage(defTtl, storage, persistInterval)
	commandProcessor := NewGoodiesCommandsProcessor(g)
	serialiser := jsonRequestResponseSerialiser{}
	handler := goodiesHTTPServer{commandProcessor, serialiser}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: &handler}
	return server
}

type goodiesHTTPServer struct {
	commandProcessor CommandProcesser
	serializer       RequestResponseSerialiser
}

func (s *goodiesHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("Cannot read incoming request")
	}
	w.Write(s.serveCommandBytes(data))
}

func (tr GoodiesHttpCommandClient) Process(req CommandRequest, res *CommandResponse) error {
	data, err := tr.serializer.SerialiseRequest(req)
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
	return tr.serializer.DeserialiseResponse(body, res)
}

func (s goodiesHTTPServer) serveCommandBytes(reqData []byte) []byte {
	var req CommandRequest
	var res CommandResponse
	err := s.serializer.DeserialiseRequest(reqData, &req)
	if err != nil {
		res = CommandResponse{false, "", err}
	} else {
		res = s.commandProcessor.HandleCommand(req)
	}

	data, err := s.serializer.SerialiseResponse(res)
	if err != nil {
		panic("Cannot serialise response")
	}
	return data
}
