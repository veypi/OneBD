package OneBD

import (
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/handler"
	"github.com/lightjiang/OneBD/libs/hpool"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Host:           "0.0.0.0:4000",
		Charset:        "",
		TimeFormat:     "",
		PostMaxMemory:  0,
		TlsCfg:         nil,
		MaxConnections: 0,
	}
	cfg.BuildLogger()
	app := New(cfg)
	app.Router().Set("/", hpool.New(func() core.Handler {
		app.Logger().Info("creating a handler")
		return &handler.BaseHandler{}
	}))
	err := app.Run()
	t.Error(err)
}
