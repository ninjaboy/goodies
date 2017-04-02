package goodies

import (
	"fmt"
	"strings"
)

// ErrUnknownCommand Indicates unsupported command requested
type ErrUnknownCommand struct {
	name string
}

func (e ErrUnknownCommand) Error() string {
	return fmt.Sprintf("ErrUnknownCommand: Unknown command requested: %v", e.name)
}

// ErrInternalError Indicates software error
type ErrInternalError struct {
	str string
}

func (e ErrInternalError) Error() string {
	return fmt.Sprintf("ErrInternalError: %v", e.str)
}

// ErrCommandArgumentsMismatch Indicates arguments count mismatch
type ErrCommandArgumentsMismatch struct {
	str string
}

func (e ErrCommandArgumentsMismatch) Error() string {
	return fmt.Sprintf("ErrCommandArgumentsMismatch: %v", e.str)
}


type ErrTypeMismatch struct {
	err string
}

func (e ErrTypeMismatch) Error() string {
	return fmt.Sprintf("ErrTypeMismatch: %v", e.err)
}

type ErrNotFound struct {
	key string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("ErrNotFound: Item for key was not found: %v", e.key)
}

type ErrDictKeyNotFound struct {
	key string
}

func (e ErrDictKeyNotFound) Error() string {
	return fmt.Sprintf("ErrDictKeyNotFound: Item for key was not found in a dictionary: %v", e.key)
}

type ErrTransformation struct {
	str string
}

func (e ErrTransformation) Error() string {
	return fmt.Sprintf("ErrTransformation: %v", e.str)
}

func ErrorFromString(str string) error {
	switch {
	case strings.HasPrefix(str, "ErrDictKeyNotFound"):
		return ErrDictKeyNotFound{getParameter(str)}
	case strings.HasPrefix(str, "ErrNotFound"):
		return ErrNotFound{getParameter(str)}
	case strings.HasPrefix(str, "ErrTypeMismatch"):
		return ErrTypeMismatch{getParameter(str)}
	case strings.HasPrefix(str, "ErrCommandArgumentsMismatch"):
		return ErrCommandArgumentsMismatch{getParameter(str)}
	case strings.HasPrefix(str, "ErrCommandArgumentsMismatch"):
		return ErrCommandArgumentsMismatch{getParameter(str)}
	case strings.HasPrefix(str, "ErrInternalError"):
		return ErrInternalError{getParameter(str)}
	case strings.HasPrefix(str, "ErrUnknownCommand"):
		return ErrUnknownCommand{getParameter(str)}
	case strings.HasPrefix(str, "ErrTransformation"):
		return ErrTransformation{getParameter(str)}
	default:
		return ErrInternalError{fmt.Sprintf("UNKNOWN ERROR RECEIVED: %v", str)}
	}
}

func getParameter(str string) string {
	return str[strings.LastIndex(str, ": ")+2:]
}
