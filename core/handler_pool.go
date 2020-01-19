package core

type HandlerPool interface {
	SetNew(newFunc func() Handler)
	Acquire() Handler
	Release(h Handler)
}
