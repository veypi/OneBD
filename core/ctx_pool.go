package core

import "net/http"

type CtxPool interface {
	SetCtx(newFunc func() Context)
	Get(w http.ResponseWriter, r *http.Request) Context
	Release(ctx Context)
}
