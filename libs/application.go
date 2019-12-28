package libs

import (
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/core"
)

type application struct {
	config  *config.Config
	ctxPool core.CtxPool
	// todo:: log 分级处理log
	// todo:: router
	//
}

func NewApplication(cfg *config.Config) core.Application {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	app := &application{
		config: config.DefaultConfig(),
	}
	app.ctxPool = NewCtxPool(func() core.Context {
		return NewContext(app)
	})
	return app
}

func (app *application) CtxPool() core.CtxPool {
	return app.ctxPool
}

func (app *application) Config() *config.Config {
	return app.config
}
