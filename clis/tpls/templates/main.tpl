//
// main.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-07 15:25
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
	{{.common.api}}.Use(app.Router())
	app.Router().Print()
	return app.Run()
}
