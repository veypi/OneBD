package core

import "github.com/lightjiang/OneBD/config"

type Application interface {
	CtxPool() CtxPool
	Config() *config.Config
}
