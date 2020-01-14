package core

import "net/http"

type Handler interface {
	// 请求声明周期 init -> get/post/etc  -> finished -> onResponse/onError(返回处理或错误同一处理)
	// init 方法最好调用ifRest 查看是否被重置过上次缓存
	Init(w http.ResponseWriter, r *http.Request) error

	Get() (interface{}, error)
	Post() (interface{}, error)
	Put() (interface{}, error)
	Patch() (interface{}, error)
	Head() (interface{}, error)
	Delete() (interface{}, error)

	Finished() error

	OnResponse(data interface{})
	OnError(err error)

	// Reset 方法最好把继承的所有handler 的reset调用一遍
	IfReset() bool
	Reset()
}

type BaseHandler func(ctx Context)
