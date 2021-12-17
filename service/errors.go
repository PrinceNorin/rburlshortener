package service

import "strings"

func newError(msg string, code int) error {
	return &simpleError{
		messages: []string{msg},
		code:     code,
	}
}

type simpleError struct {
	messages []string
	code     int
}

func (err *simpleError) Error() string {
	return strings.Join(err.messages, ", ")
}

func (err *simpleError) StatusCode() int {
	return err.code
}

func (err *simpleError) Response() interface{} {
	return err.messages
}
