package OneBD

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
)

const (
	Version = "0.1.5"
)

func New(cfg *core.Config) *libs.Application {
	return libs.NewApplication(cfg)
}
