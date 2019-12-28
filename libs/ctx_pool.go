package libs

import (
	"github.com/lightjiang/OneBD/core"
	"net/http"
	"sync"
)

type ctxPool struct {
	pool *sync.Pool
}

func NewCtxPool(newFunc func() core.Context) core.CtxPool {
	return &ctxPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

func (p *ctxPool) SetCtx(newFunc func() core.Context) {
	p.pool.New = func() interface{} {
		return newFunc()
	}
}

func (p *ctxPool) Get(w http.ResponseWriter, r *http.Request) core.Context {
	ctx := p.pool.Get().(core.Context)
	ctx.Start(w, r)
	return ctx
}

func (p *ctxPool) Release(ctx core.Context) {
	ctx.Stop()
	ctx.ResetCtx()
	p.pool.Put(ctx)
}
