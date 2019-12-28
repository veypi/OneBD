package libs

import (
	"crypto/tls"
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/core"
	"golang.org/x/net/netutil"
	"net"
	"net/http"
)

type application struct {
	config   *config.Config
	ctxPool  core.CtxPool
	server   *http.Server
	listener net.Listener
	router   core.Router
	// todo:: log 分级处理log
	// todo:: router
	//
}

func NewApplication(cfg *config.Config) core.Application {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	app := &application{
		config: cfg,
	}
	app.server = &http.Server{
		Addr:              cfg.Host,
		Handler:           app.router,
		TLSConfig:         cfg.TlsCfg,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		// TODO
		BaseContext: nil,
		// TODO
		ConnContext: nil,
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

func (app *application) Serve() error {
	l, e := app.NetListener()
	if e != nil {
		return e
	}
	return app.server.Serve(l)
}

func (app *application) NetListener() (net.Listener, error) {
	if app.listener != nil {
		return app.listener, nil
	}
	l, err := net.Listen("tcp", app.config.Host)
	if err != nil {
		return nil, err
	}
	if app.config.TlsCfg != nil && len(app.config.TlsCfg.Certificates) > 0 && app.config.TlsCfg.GetCertificate != nil {
		l = tls.NewListener(l, app.config.TlsCfg)
	}
	if app.config.MaxConnections > 0 {
		l = netutil.LimitListener(l, app.config.MaxConnections)
	}
	app.listener = l
	return app.listener, nil
}
