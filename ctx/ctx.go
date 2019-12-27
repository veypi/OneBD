package ctx

import (
	"OneBD"
	"net/http"
)

// Context 请求相关数据和过程的周期管理
type Context interface {
	// Start 开始请求 初始化ctx
	Start(w http.ResponseWriter, r *http.Request)
	// 结束请求
	Stop()
	// 重置ctx, 用于pool回收
	ResetCtx()
}

type context struct {
	writer  http.ResponseWriter
	request *http.Request
	app     OneBD.Application
}

func NewContext(app OneBD.Application) Context {
	return &context{app: app}
}

func (ctx *context) Start(w http.ResponseWriter, r *http.Request) {
	ctx.writer = w
	ctx.request = r
}
func (ctx *context) Stop() {
}

func (ctx *context) ResetCtx() {
	ctx.writer = nil
	ctx.request = nil
}
