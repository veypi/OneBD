//
// err.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-09-24 21:22
// Distributed under terms of the MIT license.
//

package rest

import (
	"errors"
	"fmt"
)

var (
	ErrCrash      = errors.New("crash")
	ErrNotFound   = errors.New("not found")
	ErrParse      = errors.New("parse request data failed")
	ErrArgMissing = fmt.Errorf("%w: missing arg", ErrParse)
	ErrArgInvalid = fmt.Errorf("%w: invalid argument", ErrParse)
)
