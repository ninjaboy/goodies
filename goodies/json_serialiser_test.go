package goodies

import "testing"

func TestRequestSerialisation(t *testing.T) {
	s := JsonRequestResponseSerialiser{}
	req := CommandRequest{Name: "SET", Parameters: []string{"Key", "Test \" Data", "30"}}

	data, err := s.SerialiseRequest(req)
	if err != nil {
		t.Error("Serialisation of request failed")
	}

	var req2 CommandRequest
	err = s.DeserialiseRequest(data, &req2)
	if err != nil {
		t.Error("Deserialisation of request failed")
	}

	if req.Name != req2.Name {
		t.Error("Deserialised request doesn't match expected")
	}

	for i, p := range req.Parameters {
		if req2.Parameters[i] != p {
			t.Errorf("Deserialised parameter %v doesn't match expected (%v != %v)",
				i, p, req2.Parameters[i])
		}
	}
}
