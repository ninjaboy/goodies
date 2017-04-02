package goodies

import (
	"encoding/json"
)

type RequestResponseSerialiser interface {
	SerialiseRequest(CommandRequest) ([]byte, error)
	DeserialiseRequest([]byte, *CommandRequest) error
	SerialiseResponse(CommandResponse) ([]byte, error)
	DeserialiseResponse([]byte, *CommandResponse) error
}

type jsonRequestResponseSerialiser struct{}

type goodiesResponseSer struct {
	Success bool
	Result  string
	ErrStr  string
}

func (ser jsonRequestResponseSerialiser) SerialiseRequest(req CommandRequest) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, ErrTransformation{err.Error()}
	}
	return data, nil
}

func (ser jsonRequestResponseSerialiser) DeserialiseRequest(data []byte, req *CommandRequest) error {
	err := json.Unmarshal(data, req)
	if err != nil {
		return ErrTransformation{err.Error()}
	}
	return nil
}

func (ser jsonRequestResponseSerialiser) SerialiseResponse(res CommandResponse) ([]byte, error) {
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

func (ser jsonRequestResponseSerialiser) DeserialiseResponse(data []byte, res *CommandResponse) error {

	forDeserialisation := goodiesResponseSer{}
	err := json.Unmarshal(data, &forDeserialisation)
	if err != nil {
		return ErrTransformation{err.Error()}
	}
	res.Result = forDeserialisation.Result
	res.Success = forDeserialisation.Success

	if res.Success == true {
		return nil
	}
	res.Err = ErrorFromString(forDeserialisation.ErrStr)
	return nil
}
