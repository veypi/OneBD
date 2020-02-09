package libs

import (
	"crypto/tls"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/router"
	"github.com/lightjiang/utils/log"
	"golang.org/x/net/netutil"
	"net"
	"net/http"
)

type Application struct {
	config   *core.Config
	server   *http.Server
	listener net.Listener
	router   core.Router
	// todo:: log 分级处理log
	// todo:: router
	//
}

func NewApplication(cfg *core.Config) *Application {
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	cfg.BuildLogger()
	app := &Application{
		config: cfg,
	}
	app.server = &http.Server{
		Addr:              cfg.Host,
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
	// 判断是否使用内置路由
	if cfg.Router != nil {
		app.router = cfg.Router
	} else {
		app.router = router.NewMainRouter()
	}
	app.server.Handler = app.router
	return app
}

func (app *Application) Router() core.Router {
	return app.router
}

func (app *Application) Config() *core.Config {
	return app.config
}

func (app *Application) Run() error {
	log.Info().Msg("\nRouting Table" + app.router.String())
	l, e := app.netListener()
	if e != nil {
		return e
	}
	log.Info().Msg("listening http://" + app.config.Host)
	return app.server.Serve(l)
}

func (app *Application) netListener() (net.Listener, error) {
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
