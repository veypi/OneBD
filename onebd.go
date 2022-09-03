package OneBD

import (
	"github.com/veypi/OneBD/core"
	"github.com/veypi/OneBD/libs"
	"github.com/veypi/OneBD/libs/handler"
	"github.com/veypi/OneBD/libs/hpool"
)

const (
	Version = "v0.4.4"
)

type Router = core.Router
type Meta = core.Meta
type MetaFunc = core.MetaFunc
type Handler = core.Handler
type HandlerPool = core.HandlerPool
type RequestLifeCycle = core.RequestLifeCycle
type Config = core.Config
type BaseHandler = handler.Base
type WebsocketConn = core.WebSocketConn

var NewHandlerPool = hpool.New

func New(cfg *Config) *libs.Application {
	return libs.NewApplication(cfg)
}
