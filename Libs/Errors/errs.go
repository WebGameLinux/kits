package Errors

import "fmt"

const (
		nilPointCode   = 20001
		nilClientCode  = 20002
		typeError      = 20003
		unmarshalError = 20004
)

type Error struct {
		Message string
		Code    int
}

func _error(innerCode int, innerMessage string, message string) *Error {
		return &Error{Message: innerMessage + ",cause : " + message, Code: innerCode}
}

func NewError(message string, code int) *Error {
		return &Error{Message: message, Code: code}
}

func NilPointError(message string) *Error {
		return _error(nilPointCode, "nil point error", message)
}

func NilClientError(message string) *Error {
		return _error(nilClientCode, "nil client error", message)
}

func TypeError(message string) *Error {
		return _error(typeError, "param type error", message)
}

func UnmarshalError(message string) *Error {
		return _error(unmarshalError, "unmarshal failed error", message)
}

func (e *Error) Error() string {
		return fmt.Sprintf("%d - %s", e.Code, e.Message)
}
