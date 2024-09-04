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
	"github.com/veypi/utils/logx"
)

func main() {
	cfg.CMD.Command = runWeb
	cfg.CMD.Parse()
	err := cfg.CMD.Run()
	if err != nil {
		logx.Warn().Msg(err.Error())
	}
}

func runWeb() error {
	app, err := rest.New(&cfg.Config.RestConf)
	if err != nil {
		return err
	}
    apiRouter := app.Router().SubRouter("api")
	{{.common.api}}.Use(apiRouter)
	apiRouter.Use(func(x *rest.X, data any) error {
		if data != nil {
			return x.JSON(data)
		}
		return nil
	})
	app.Router().Print()
	return app.Run()
}
