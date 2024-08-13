//
// main.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-07 15:25
// Distributed under terms of the MIT license.
//
{{.common.noedit}}

package main

import (
	"{{.common.repo}}/cfg"

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
	app.Router().Print()
	return app.Run()
}
