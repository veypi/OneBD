package OneBD

import (
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/libs"
)

const (
	Version = "0.1.0"
)

func New(cfg *config.Config) *libs.Application {
	return libs.NewApplication(cfg)
}
