package core

import (
	"github.com/veypi/OneBD/rfc"
	"io"
	"net/http"
	"time"
)

// http包 原生处理函数
type OriginFunc = func(w http.ResponseWriter, r *http.Request)

// 元处理函数,非周期处理下对request 做处理
type MetaFunc = func(m Meta)

// Meta 请求辅助处理单元
type Meta interface {
	http.ResponseWriter
	// 生命周期 Init-> Write/Flush/StreamWrite ->TryReset
	Init(w http.ResponseWriter, r *http.Request, params map[string]uint)
	TryReset()
	RemoteAddr() string
	RequestPath() string
	Method() rfc.Method
	ReadJson(ptr interface{}) error
	UnmarshalBody(ptr interface{}, fc func([]byte, interface{}) error) error
	Query(string) string
	GetHeader(string) string
	Params(string) string
	ParamsInt(string) int
	Status() rfc.Status
	SetHeader(key, value string)
	AliveTime() time.Duration
	// 根据设计哲学，不提供cookie 相关处理函数
	// 必要附带信息请手动添加到header里，如各种权限token
	//SetCookie(key, value string)
	StreamRead(io.Writer) (int64, error)
	StreamWrite(io.Reader) (int64, error)
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
	// 禁用 meta 对 response 做处理, 用于用户自己管理 responseWriter
	DisableSelfWriter()
}
