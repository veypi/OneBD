//
// err.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-28 14:16
// Distributed under terms of the GPL license.
//

package crud

import "fmt"

type Err struct {
	msg string
}

var (
	ErrMissArg = NewErr("miss arg: %s")
	ErrParse   = NewErr("parse request data failed")
	Err500     = NewErr("internal server error")
)

func NewErr(msg string) *Err {
	return &Err{msg: msg}
}

func (e *Err) Error() string {
	return e.msg
}

func (e *Err) Fmt(msg ...any) error {
	return fmt.Errorf(e.msg, msg...)
}

func (e *Err) Wrap(err error) error {
	return fmt.Errorf("%w: %s", err, e)
}
