package main

import (
	"errors"
)

func GetFieldsConsideringQuotes(command string) ([]string, error) {
	var items []string
	var buf string
	var quoteMode bool
	var escapeMode bool
	quote := rune('"')
	for _, c := range command {
		switch {
		case c == quote && escapeMode:
			escapeMode = false
			buf = buf[:len(buf)-1] + string(quote)
		case c == ' ' && !quoteMode:
			fallthrough
		case c == quote && quoteMode:
			if buf != "" {
				items = append(items, buf)
			}
			buf = ""
			if quoteMode {
				quoteMode = false
			}
		case c == '"' && !quoteMode:
			quoteMode = true
		case c == '\n':
			continue
		case c == '\\':
			escapeMode = true
			fallthrough
		default:
			buf += string(c)
		}
	}
	if quoteMode {
		return nil, errors.New("Malformed input string. Quotes mismatched")
	}
	items = append(items, buf)
	return items, nil
}
