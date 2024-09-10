//
// typ.go
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
// transform any type args to *arg

package vtools

import (
    "time"
    "strconv"
)

var TimeFormat = "2006-01-02 15:04:05"

func Str2Int(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func Str2Uint(s string) uint {
	return uint(Str2Int(s))
}

func Str2Time(s string) time.Time {
	t, _ := time.Parse(TimeFormat, s)
	return t
}

func Star[T any](t T) *T {
	return &t
}

func StarFunc[T any](t T, ok bool) *T {
	if !ok {
		return nil
	}
	return &t
}

func StartError[T any](t T, err error) *T {
	if err != nil {
		return nil
	}
	return &t
}
