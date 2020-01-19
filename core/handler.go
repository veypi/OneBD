package core

import (
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
)

type Handler interface {
	// 请求生命周期 init -> get/post/etc  -> finished -> onResponse(返回处理) -> tryReset
	// 一旦某一周期返回会触发OnError 接口 并结束
	// init 方法最好调用TryRest 确保重置过
	Init(w http.ResponseWriter, r *http.Request) error
	Get() (interface{}, error)
	Post() (interface{}, error)
	Put() (interface{}, error)
	Patch() (interface{}, error)
	Head() (interface{}, error)
	Options() (interface{}, error)
	Delete() (interface{}, error)
	Trace() (interface{}, error)
	Finished() error
	OnResponse(data interface{})
	OnError(err error)
	// TryReset 方法最好把继承的所有handler 的TryReset调用一遍
	TryReset() // 尝试reset 历史信息 若没有重置，则直接重置， 若有则不做任何事情

	Meta() Meta
}

// 元处理函数,非周期处理下对request 做处理
type MetaFunc func(m Meta)

// Meta 请求辅助处理单元
type Meta interface {
	// 生命周期 Init->TryReset
	Init(w http.ResponseWriter, r *http.Request)
	TryReset()
	Method() rfc.Method
	SetStatus(status rfc.Status)
	Status() rfc.Status
	SetHeader(key, value string)
	// 根据设计哲学，不提供cookie 相关处理函数
	// 必要附带信息请手动添加到header里，如各种权限token
	//SetCookie(key, value string)
}
