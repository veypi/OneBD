package OneBD

import "OneBD/ctx"

const (
	Version = "0.0.1"
)

type Application interface {
	CtxPool() *ctx.Pool
}

type application struct {
	config  *Config
	ctxPool *ctx.Pool
	// todo:: log 分级处理log
	// todo:: router
	//
}

func New() Application {
	app := &application{
		config: DefaultConfig(),
	}
	app.ctxPool = ctx.New(func() ctx.Context {
		return ctx.NewContext(app)
	})
	return app
}

func (app *application) CtxPool() *ctx.Pool {
	return app.ctxPool
}
