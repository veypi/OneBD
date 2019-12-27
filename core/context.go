package core

import "net/http"

// Context 请求相关数据和过程的周期管理
type Context interface {
	// Start 开始请求 初始化ctx
	Start(w http.ResponseWriter, r *http.Request)
	// 结束请求
	Stop()
	// 重置ctx, 用于pool回收
	ResetCtx()
}
