package ctx

import (
	"net/http"
	"sync"
)

type Pool struct {
	pool *sync.Pool
}

func New(newFunc func() Context) *Pool {
	return &Pool{
		pool: &sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

func (p *Pool) Get(w http.ResponseWriter, r *http.Request) Context {
	ctx := p.pool.Get().(Context)
	ctx.Start(w, r)
	return ctx
}

func (p *Pool) Release(ctx Context) {
	ctx.Stop()
	ctx.ResetCtx()
	p.pool.Put(ctx)
}
