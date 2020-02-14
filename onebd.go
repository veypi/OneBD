package OneBD

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
)

const (
	Version = "v0.3.5"
)

type Router = core.Router
type Meta = core.Meta
type MetaFunc = core.MetaFunc
type Handler = core.Handler
type HandlerPool = core.HandlerPool
type RequestLifeCycle = core.RequestLifeCycle
type Config = core.Config

func New(cfg *Config) *libs.Application {
	return libs.NewApplication(cfg)
}
