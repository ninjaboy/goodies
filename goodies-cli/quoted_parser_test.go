package main

import (
	"testing"
)

func TestQuotedParser(t *testing.T) {
	var input string
	input = "1 2 3"
	res, _ := GetFieldsConsideringQuotes(input)
	if len(res) != 3 || res[0] != "1" || res[1] != "2" || res[2] != "3" {
		t.Error("Unquoted parsing doesn't work")
	}

	input = "1 \"2\\\" \" 3"
	res, _ = GetFieldsConsideringQuotes(input)
	if len(res) != 3 || res[0] != "1" || res[1] != "2\" " || res[2] != "3" {
		t.Error("Quoted parsing doesn't work")
	}

	input = "1 \"2\\\" \" \"3 "
	res, err := GetFieldsConsideringQuotes(input)
	if err == nil {
		t.Error("Err mismatch quotes is expected but not thrown")
	}

}
