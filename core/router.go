package core

import (
	"embed"
	"github.com/veypi/OneBD/rfc"
	"net/http"
)

type Router interface {
	// 设置 请求 生命周期管理函数
	SetRequestLifeCycle(cycle RequestLifeCycle)
	// 注册 get/post/patch/delete/... 等方法
	Set(prefix string, fc interface{}, allowedMethods ...rfc.Method) Router
	WS(prefix string, upgrader WebSocketFunc) Router
	// 设置静态资源访问
	Static(prefix string, directory string)
	EmbedDir(prefix string, fs embed.FS, fsPrefix string)
	EmbedFile(prefix string, path []byte)

	// 自路由
	SubRouter(name string) Router
	ServeHTTP(http.ResponseWriter, *http.Request)

	SetStatusFunc(status rfc.Status, rf MetaFunc)
	SetNotFoundFunc(fc MetaFunc)
	// 对于内部panic错误， 返回500 错误 并执行相应回调
	SetInternalErrorFunc(fc MetaFunc)
	String() string
	// 返回绝对路径
	AbsPath() string
}

type RequestLifeCycle = func(interface{}, Meta)
