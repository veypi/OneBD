package hpool

import (
	"github.com/lightjiang/OneBD/core"
	"sync"
)

type handlerPool struct {
	app     core.AppInfo
	newFunc func() core.Handler
	pool    *sync.Pool
}

func NewHandlerPool(newFunc func() core.Handler, app core.AppInfo) core.HandlerPool {
	p := &handlerPool{
		pool: &sync.Pool{},
		app:  app,
	}
	p.SetNew(newFunc)
	return p
}

func (p *handlerPool) SetNew(newFunc func() core.Handler) {
	p.newFunc = newFunc
	p.pool.New = func() interface{} {
		return newFunc()
	}
}

func (p *handlerPool) Acquire() core.Handler {
	ctx := p.pool.Get().(core.Handler)
	return ctx
}

func (p *handlerPool) Release(h core.Handler) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	h.TryReset()
	p.pool.Put(h)
}
