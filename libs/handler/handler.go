package handler

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
	"github.com/lightjiang/OneBD/utils"
	"net/http"
	"sync/atomic"
)

//BaseHandler  请求 基本处理流程
type BaseHandler struct {
	utils.FastLocker
	empty   uint32
	payLoad core.Meta
	app     *libs.Application
}

func (h *BaseHandler) Init(w http.ResponseWriter, r *http.Request) error {
	h.TryReset()
	h.payLoad.Init(w, r)
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

func (h *BaseHandler) Options() (interface{}, error) {
	panic("implement me")
}

func (h *BaseHandler) Trace() (interface{}, error) {
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
		// 确保payload被释放
		h.payLoad.TryReset()
		// 确保锁释放
		h.Unlock()
	}
}

func (h *BaseHandler) Meta() core.Meta {
	return h.payLoad
}
