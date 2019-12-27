package libs

import "github.com/lightjiang/OneBD/core"

type application struct {
	config  *Config
	ctxPool core.CtxPool
	// todo:: log 分级处理log
	// todo:: router
	//
}

func NewApplication() core.Application {
	app := &application{
		config: DefaultConfig(),
	}
	app.ctxPool = NewCtxPool(func() core.Context {
		return NewContext(app)
	})
	return app
}

func (app *application) CtxPool() core.CtxPool {
	return app.ctxPool
}
