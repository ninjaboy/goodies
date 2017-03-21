package goodies

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpTransport struct {
	address    string
	serializer RequestResponseSerialiser
	//credentials can be added
}

type HttpServer struct {
	cp         CommandProcesser
	serializer RequestResponseSerialiser
}

func (tr HttpTransport) Process(req GoodiesRequest, res *GoodiesResponse) error {
	data, err := tr.serializer.serialiseRequest(req)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequest("POST", tr.address, bytes.NewReader(data))
	httpRequest.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpRequest)
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

func (tr HttpServer) Serve(reqData []byte) []byte {
	var req GoodiesRequest
	var res GoodiesResponse
	err := tr.serializer.deserialiseRequest(reqData, &req)
	if err != nil {
		res = GoodiesResponse{false, "", err}
	} else {
		res = tr.cp.HandleCommand(req)
	}

	data, err := tr.serializer.serialiseResponse(res)
	if err != nil {
		panic("Cannot serialise response, don't know what to do :(")
	}
	return data
}
