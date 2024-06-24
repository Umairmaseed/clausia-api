package middleware

import (
	"encoding/json"
	"errors"
)

// ErrMessage is used when only a message should be returned
type ErrMessage struct {
	Code int   `json:"code"`
	Err  error `json:"message"`
}

// NewErrMsg ...
func NewErrMsg(code int, err error) ErrMessage {
	return ErrMessage{
		Code: code,
		Err:  err,
	}
}

func NewErrStr(code int, err string) ErrMessage {
	return ErrMessage{
		Code: code,
		Err:  errors.New(err),
	}
}

func (e ErrMessage) Error() string {
	return e.Err.Error()
}

// ErrJSON should be used when a object should be converted to a message
type ErrJSON struct {
	Code   int                    `json:"code"`
	Object map[string]interface{} `json:"errorValue"`
}

func (e ErrJSON) Error() string {
	data, err := json.Marshal(e.Object)
	if err != nil {
		panic("Could not encode error object")
	}
	return string(data)
}
