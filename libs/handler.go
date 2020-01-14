package libs

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"net/http"
	"sync/atomic"
)

type payLoad struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func (p *payLoad) Method() rfc.Method {
	panic("implement me")
}

func (p *payLoad) SetStatus(status rfc.Status) {
	panic("implement me")
}

func (p *payLoad) Status() rfc.Status {
	panic("implement me")
}

func (p *payLoad) SetCookie(key, value string) {
	panic("implement me")
}

//BaseHandler  请求 基本处理流程
type BaseHandler struct {
	utils.FastLocker
	empty   uint32
	payLoad payLoad
	app     *Application
}

func (h *BaseHandler) Init(w http.ResponseWriter, r *http.Request) error {
	h.TryReset()
	h.payLoad.Request = r
	h.payLoad.Writer = w
	atomic.StoreUint32(&h.empty, 0)
	return nil
}

func (h *BaseHandler) Get() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Post() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Put() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Patch() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Head() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Delete() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Finished() error {
	panic("implement me")
}

func (h *BaseHandler) OnResponse(data interface{}) {
	panic("implement me")
}

func (h *BaseHandler) OnError(err error) {
	panic("implement me")
}

func (h *BaseHandler) TryReset() {
	if atomic.CompareAndSwapUint32(&h.empty, 1, 0) {
		h.payLoad.Writer = nil
		h.payLoad.Request = nil
	}
}

func (h *BaseHandler) Meta() core.Meta {
	return &h.payLoad
}
