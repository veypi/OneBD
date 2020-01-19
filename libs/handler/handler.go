package handler

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"sync/atomic"
)

//BaseHandler  请求 基本处理流程
type BaseHandler struct {
	utils.FastLocker
	empty   uint32
	payLoad core.Meta
}

func (h *BaseHandler) Init(m core.Meta) error {
	h.TryReset()
	h.payLoad = m
	atomic.StoreUint32(&h.empty, 0)
	return nil
}

func (h *BaseHandler) Get() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Post() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Put() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Patch() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Head() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Delete() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Options() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Trace() (interface{}, error) {
	h.Meta().SetStatus(rfc.StatusNotFound)
	return nil, nil
}

func (h *BaseHandler) Finished() error {
	return nil
}

func (h *BaseHandler) OnResponse(data interface{}) {
}

func (h *BaseHandler) OnError(err error) {
}

func (h *BaseHandler) TryReset() {
	if atomic.CompareAndSwapUint32(&h.empty, 1, 0) {
		// 确保payload被释放
		h.payLoad = nil
		// 确保锁释放
		h.Unlock()
	}
}

func (h *BaseHandler) Meta() core.Meta {
	return h.payLoad
}
