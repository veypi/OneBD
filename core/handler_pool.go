package core

import "net/http"

type HandlerPool interface {
	SetNew(newFunc func() Handler)
	Acquire(w http.ResponseWriter, r *http.Request) Handler
	Release(h Handler)
}
