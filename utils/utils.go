package utils

import (
	"runtime"
	"strconv"
)

func CallPath(s int) string {
	_, f, l, _ := runtime.Caller(s + 1)
	return f + ":" + strconv.Itoa(l)
}

func ListContainStr(l []string, s string) (bool, int) {
	for i, v := range l {
		if v == s {
			return true, i
		}
	}
	return false, -1
}
