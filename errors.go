package clickhouse

import (
	"fmt"
	"net"
)

type ChErrorCode int

const (
	ErrConnection ChErrorCode = iota
	ErrNoRowsClickhouse
	ErrOtherClickhouse
)

type ChError struct {
	code ChErrorCode
	msg  string
}

func (err ChError) Error() string {
	return err.msg
}

func (err ChError) Code() ChErrorCode {
	return err.code
}

func NewChError(code ChErrorCode, err error) *ChError {
	return &ChError{code: code, msg: err.Error()}
}

var ErrClickhouseNoRows = fmt.Errorf("sql: no rows in result set")

func ConvertError(err error) *ChError {
	if err == nil {
		return nil
	}

	if ne, ok := err.(net.Error); ok {
		return NewChError(ErrConnection, ne)
	}

	if err.Error() == ErrClickhouseNoRows.Error() {
		return NewChError(ErrNoRowsClickhouse, err)
	}

	return NewChError(ErrOtherClickhouse, err)
}
