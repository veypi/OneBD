package meta

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
	"sync"
	"sync/atomic"
)

var pool = sync.Pool{
	New: func() interface{} {
		return &payLoad{}
	},
}

// payLoad 请求基本处理包
type payLoad struct {
	empty       uint32
	app         core.AppInfo
	writer      http.ResponseWriter
	request     *http.Request
	status      rfc.Status
	ifSetStatus bool
}

func (p *payLoad) Init(w http.ResponseWriter, r *http.Request, app core.AppInfo) {
	p.TryReset()
	p.app = app
	p.writer = w
	p.request = r
	p.status = rfc.StatusOK
	p.ifSetStatus = false
	atomic.StoreUint32(&p.empty, 0)
}

func (p *payLoad) Method() rfc.Method {
	return rfc.Method(p.request.Method)
}

func (p *payLoad) SetStatus(status rfc.Status) {
	p.status = status
}

func (p *payLoad) Status() rfc.Status {
	return p.status
}

func (p *payLoad) SetHeader(key, value string) {
	if p.ifSetStatus {
	}
}

func (p *payLoad) TryReset() {
	if atomic.CompareAndSwapUint32(&p.empty, 1, 0) {
		p.writer = nil
		p.request = nil
	}
}

func Acquire(w http.ResponseWriter, r *http.Request, app core.AppInfo) core.Meta {
	m := pool.Get().(core.Meta)
	m.Init(w, r, app)
	return m
}

func Release(m core.Meta) {
	m.TryReset()
	pool.Put(m)
}
