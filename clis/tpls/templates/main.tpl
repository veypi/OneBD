//
// main.go
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package main

import (
	"{{.common.repo}}/cfg"
	"{{.common.repo}}/{{.common.api}}"

	"github.com/veypi/OneBD/rest"
	"github.com/veypi/OneBD/rest/middlewares"
	"github.com/veypi/utils/logv"
)

func main() {
	cfg.CMD.Command = runWeb
	cfg.CMD.Parse()
	err := cfg.CMD.Run()
	if err != nil {
		logv.Warn().Msg(err.Error())
	}
}

func runWeb() error {
	app, err := rest.New(&cfg.Config.RestConf)
	if err != nil {
		return err
	}
    apiRouter := app.Router().SubRouter("api")
	{{.common.api}}.Use(apiRouter)

	apiRouter.Use(middlewares.JsonResponse)
	apiRouter.SetErrFunc(middlewares.JsonErrorResponse)
	app.Router().Print()
	return app.Run()
}
