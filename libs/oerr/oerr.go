package oerr

import (
	"errors"
	"strconv"
	"strings"
)

// 错误描述

type Code uint

/*

5位10进制码表示错误, 00000 etc.

0 代表未知，或不必定义的有通用意义的错误

## 第1位 错误类型
	- 1 : 系统级错误 比如 内存申请失败, 系统调用失败，文件打开失败等等
	- 2 : 数据库错误
	- 3 : 资源错误
	- 4 : 权限错误
	- 5 : 配置错误
	- 6 : 参数错误

## 第2位 2级错误类型

## 第3,4位 具体错误编号

## 第5位 错误严重程度
	- 0 : 无任何影响错误，简单重试可以解决
	- 1 : 无影响错误，重试不可解决
	- 2 : 有影响用户体验或系统性能错误, 重试可解决
    - 3 : 有影响用户体验或系统性能错误, 重试不可解决
	- 4 : 有影响组件功能的错误, 重试可解决
	- 5 : 有影响组件功能的错误， 重试不可解决
	- 6 : 有影响服务运行的错误， 重启可解决
	- 7 : 有影响服务运行的错误，重启不可解决
	- 8 : 有影响系统运行的错误
	- 9 : 本不可能发生的错误，例如被人攻击导致数据异常产生的逻辑错误

*/
const (
	Unknown Code = 0
)

// 参数类型错误
/*
	-1: 协议参数
	-2: 接口参数
	-3: 函数参数
*/
const (
	MethodNotSupport Code = 61111
	MethodNotAllowed Code = 61121

	UrlPatternNotSupport Code = 63117
)

var codeMap = map[Code]string{
	Unknown:              "unknown error",
	MethodNotSupport:     "this http method is not supported",
	MethodNotAllowed:     "this http method is not allowed",
	UrlPatternNotSupport: "this router's url pattern is not supported.",
}

func (c Code) String() string {
	s, ok := codeMap[c]
	if ok && len(s) > 0 {
		return s
	}
	return codeMap[Unknown]
}

// 附加错误详细原因
func (c Code) Attach(err ...error) error {
	return errors.New(strings.Join([]string{strconv.Itoa(int(c)), ":", c.String(), ":"}, ""))
}

func OfType(errMsg string) Code {
	res := strings.SplitN(errMsg, ":", 1)
	if len(res) > 0 {
		c, _ := strconv.Atoi(res[0])
		return Code(c)
	}
	return Unknown
}
