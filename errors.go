package wh

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrorCode int
type WhError struct {
	Code ErrorCode
	Msg  string
}

const (
	CodeIgnoreError ErrorCode = iota
	CodeFormatError
)

var IgnoreError = &WhError{CodeIgnoreError, "ignore"} //数据处理方法返回此错误将不忽略此条件

func NewFormatError(msg string) *WhError {
	return &WhError{CodeFormatError, "(data fromat error)" + msg} //数据格式错误
}

func (e *WhError) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code, e.Msg)
}

func IsError(err error, code ErrorCode) bool {
	cerr := errors.Cause(err)
	werr, ok := cerr.(*WhError)
	if !ok {
		return false
	}

	return werr.Code == code
}
