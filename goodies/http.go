package goodies

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpCommandClient struct {
	address    string
	serializer RequestResponseSerialiser
	client     http.Client
}

func NewGoodiesHttpCommandClient(address string) HttpCommandClient {
	client := http.Client{}
	ser := JsonRequestResponseSerialiser{}
	return HttpCommandClient{address, ser, client}
}

func NewGoodiesHttpServer(port string, defTtl time.Duration, storage string, persistInterval time.Duration) *http.Server {
	g := NewGoodiesPersistedStorage(defTtl, storage, persistInterval)
	commandProcessor := NewGoodiesCommandsProcessor(g)
	serialiser := JsonRequestResponseSerialiser{}
	handler := goodiesHTTPServer{commandProcessor, serialiser}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: &handler}
	return server
}

type goodiesHTTPServer struct {
	commandProcessor CommandProcessor
	serializer       RequestResponseSerialiser
}

func (s *goodiesHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("Cannot read incoming request") //TODO: there are potentially better ways to handle this. Probably just return error
	}
	w.Write(s.serveCommandBytes(data))
}

func (tr HttpCommandClient) Process(req CommandRequest) CommandResponse {
	data, err := tr.serializer.SerialiseRequest(req)
	if err != nil {
		return NewCommandResponseFromError(err)
	}

	httpRequest, err := http.NewRequest("POST", tr.address, bytes.NewReader(data))
	httpRequest.Header.Set("Content-Type", "application/json")

	resp, err := tr.client.Do(httpRequest)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NewCommandResponseFromError(ErrInternalError{fmt.Sprintf("Conectivity issue: %v", resp.Status)})
	}
	body, _ := ioutil.ReadAll(resp.Body)

	res := &CommandResponse{}
	err = tr.serializer.DeserialiseResponse(body, res)
	if err != nil {
		return NewCommandResponseFromError(err)
	}
	return *res
}

func (s goodiesHTTPServer) serveCommandBytes(reqData []byte) []byte {
	var req CommandRequest
	var res CommandResponse
	err := s.serializer.DeserialiseRequest(reqData, &req)
	if err != nil {
		res = NewCommandResponseFromError(err)
	} else {
		res = s.commandProcessor.Process(req)
	}

	data, err := s.serializer.SerialiseResponse(res)
	if err != nil {
		panic("Cannot serialise response")
	}
	return data
}
