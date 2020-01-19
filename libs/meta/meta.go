package meta

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
	"sync/atomic"
)

// payLoad 请求基本处理包
type payLoad struct {
	empty       uint32
	writer      http.ResponseWriter
	request     *http.Request
	status      rfc.Status
	ifSetStatus bool
}

func (p *payLoad) Init(w http.ResponseWriter, r *http.Request) {
	p.TryReset()
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

func New(w http.ResponseWriter, r *http.Request) core.Meta {
	p := &payLoad{}
	p.Init(w, r)
	return p
}
