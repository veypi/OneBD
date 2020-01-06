package core

import "github.com/lightjiang/OneBD/rfc"

type Handler interface {
	ParseBody(ctx Context) error
	Init(ctx Context) error
	Method(ctx Context, method rfc.Method)
	Get(ctx Context) (interface{}, error)
	Post(ctx Context) (interface{}, error)
	Put(ctx Context) (interface{}, error)
	Patch(ctx Context) (interface{}, error)
	Head(ctx Context) (interface{}, error)
	Delete(ctx Context) (interface{}, error)
	OnError(ctx Context, err error) (interface{}, error)
	OnResponse(ctx Context, data interface{})
}

type BaseHandler func(ctx Context)
