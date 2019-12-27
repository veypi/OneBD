package core

import "net/http"

type CtxPool interface {
	Get(w http.ResponseWriter, r *http.Request) Context
	Release(ctx Context)
}
