package core

import (
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
)

type Handler interface {
	// 请求生命周期 init -> get/post/etc  -> finished -> onResponse/onError(返回处理或错误同一处理) -> tryReset
	// init 方法最好调用TryRest 确保重置过
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
	// TryReset 方法最好把继承的所有handler 的TryReset调用一遍
	TryReset() // 尝试reset 历史信息 若没有重置，则直接重置， 若有则不做任何事情

	Meta() Meta
}

// Meta handler 辅助处理单元
type Meta interface {
	Method() rfc.Method
	SetStatus(status rfc.Status)
	Status() rfc.Status
	SetCookie(key, value string)
}
