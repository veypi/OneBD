package handler

import (
	"github.com/veypi/OneBD/core"
	"github.com/veypi/OneBD/rfc"
	"github.com/veypi/utils"
	"github.com/veypi/utils/log"
)

//Base  请求 基本处理流程
type Base struct {
	utils.FastLocker
	empty utils.SafeBool
	meta  core.Meta
}

func (h *Base) Init(m core.Meta) error {
	h.TryReset()
	h.meta = m
	h.empty.ForceSetFalse()
	return nil
}

func (h *Base) Get() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Post() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Put() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Patch() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Head() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Delete() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Options() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Trace() (interface{}, error) {
	h.Meta().WriteHeader(rfc.StatusNotFound)
	return nil, nil
}

func (h *Base) Finished() error {
	return nil
}

func (h *Base) OnResponse(interface{}) {
}

func (h *Base) OnError(err error) {
	log.Warn().Err(err).Msg("err")
}

func (h *Base) TryReset() {
	if h.empty.SetTrue() {
		// 确保payload被释放
		h.meta = nil
		// 确保锁释放
		h.Unlock()
	}
}

func (h *Base) Meta() core.Meta {
	return h.meta
}
