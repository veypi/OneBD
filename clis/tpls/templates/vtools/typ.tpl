//
// typ.go
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package vtools

import "time"

var TimeFormat = "2006-01-02 15:04:05"

func Str2Int(s string) int {
	return 0
}

func Str2Time(s string) time.Time {
	t, err := time.Parse(TimeFormat, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
