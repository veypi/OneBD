//
// server.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-06 20:00
// Distributed under terms of the MIT license.
//

package rest

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/veypi/utils/logx"
	"golang.org/x/net/netutil"
)

func New(c *RestConf) (*Application, error) {
	if err := c.IsValid(); err != nil {
		return nil, err
	}
	app := &Application{
		config: c,
		router: NewRouter(),
	}
	app.server = &http.Server{
		Addr:              c.Url(),
		TLSConfig:         c.TlsCfg,
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
	app.server.Handler = app.Router()
	return app, nil
}

type Application struct {
	router   Router
	config   *RestConf
	server   *http.Server
	listener net.Listener
}

func (app *Application) Router() Router {
	return app.router
}

func (app *Application) Run() error {
	logx.WithNoCaller.Info().Msg("listening " + app.config.Url())
	l, e := app.netListener()
	if e != nil {
		return e
	}
	return app.server.Serve(l)
}

func (app *Application) netListener() (net.Listener, error) {
	if app.listener != nil {
		return app.listener, nil
	}
	l, err := net.Listen("tcp", app.config.Url())
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
