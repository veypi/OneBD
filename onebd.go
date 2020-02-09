package OneBD

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
)

const (
	Version = "v0.3.4"
)

func New(cfg *core.Config) *libs.Application {
	return libs.NewApplication(cfg)
}
