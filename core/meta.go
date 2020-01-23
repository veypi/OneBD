package core

import (
	"github.com/lightjiang/OneBD/rfc"
	"io"
	"net/http"
)

// 元处理函数,非周期处理下对request 做处理
type MetaFunc func(m Meta)

// Meta 请求辅助处理单元
type Meta interface {
	// 生命周期 Init-> Flush/StreamWrite ->TryReset
	Init(w http.ResponseWriter, r *http.Request, app AppInfo)
	TryReset()
	RemoteAddr() string
	RequestPath() string
	Query(string) string
	Params(string) string
	Method() rfc.Method
	SetStatus(status rfc.Status)
	Status() rfc.Status
	SetHeader(key, value string)
	// 根据设计哲学，不提供cookie 相关处理函数
	// 必要附带信息请手动添加到header里，如各种权限token
	//SetCookie(key, value string)
	StreamRead(io.Writer)
	StreamWrite(io.Reader)
	Flush()
	Write([]byte)
	ResetBuf()
}
