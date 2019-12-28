package OneBD

import (
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
)

const (
	Version = "0.0.1"
)

func New(cfg *config.Config) core.Application {
	return libs.NewApplication(cfg)
}
