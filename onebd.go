package OneBD

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
)

const (
	Version = "0.0.1"
)

func New() core.Application {
	return libs.NewApplication()
}
