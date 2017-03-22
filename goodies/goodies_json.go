package goodies

import (
	"encoding/json"
)

type JsonRequestResponseSerialiser struct{}

type goodiesResponseSer struct {
	Success bool
	Result  string
	ErrStr  string
}

func (ser JsonRequestResponseSerialiser) serialiseRequest(req GoodiesRequest) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, ErrTransformation{err.Error()}
	}
	return data, nil
}

func (ser JsonRequestResponseSerialiser) deserialiseRequest(data []byte, req *GoodiesRequest) error {
	err := json.Unmarshal(data, req)
	if err != nil {
		return ErrTransformation{err.Error()}
	}
	return nil
}

func (ser JsonRequestResponseSerialiser) serialiseResponse(res GoodiesResponse) ([]byte, error) {
	var errDesc string
	if res.Err != nil {
		errDesc = res.Err.Error()
	}
	forSerialisation := goodiesResponseSer{res.Success, res.Result, errDesc}
	data, err := json.Marshal(forSerialisation)
	if err != nil {
		return nil, ErrTransformation{err.Error()}
	}
	return data, nil
}

func (ser JsonRequestResponseSerialiser) deserialiseResponse(data []byte, res *GoodiesResponse) error {
	forDeserialisation := goodiesResponseSer{}
	err := json.Unmarshal(data, forDeserialisation)
	if err != nil {
		return ErrTransformation{err.Error()}
	}
	res.Result = forDeserialisation.Result
	res.Success = forDeserialisation.Success
	res.Err = ErrorFromString(forDeserialisation.ErrStr)
	return nil
}
