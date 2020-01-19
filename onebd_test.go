package OneBD

import (
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/handler"
	"testing"
)

type testHandler struct {
	handler.BaseHandler
}

func (h *testHandler) Get() (interface{}, error) {
	return nil, nil
}

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
	app.Router().Set("/", func() core.Handler {
		app.Logger().Warn("creating a handler")
		return &testHandler{}
	})
	app.Router().SetNotFoundFunc(func(meta core.Meta) {
		app.Logger().Info("checking 404 status")
	})
	app.Router().SetInternalErrorFunc(func(meta core.Meta) {
		app.Logger().Info("checking 500 status")
	})
	err := app.Run()
	t.Error(err)
}
