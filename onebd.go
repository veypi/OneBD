package OneBD

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
	"github.com/lightjiang/OneBD/libs/handler"
	"github.com/lightjiang/OneBD/libs/hpool"
)

const (
	Version = "v0.4.1"
)

type Router = core.Router
type Meta = core.Meta
type MetaFunc = core.MetaFunc
type Handler = core.Handler
type HandlerPool = core.HandlerPool
type RequestLifeCycle = core.RequestLifeCycle
type Config = core.Config
type BaseHandler = handler.Base

var NewHandlerPool = hpool.New

func New(cfg *Config) *libs.Application {
	return libs.NewApplication(cfg)
}
