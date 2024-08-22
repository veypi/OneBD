//
// app.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 19:51
// Distributed under terms of the MIT license.
//

package app

import (
	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils/logx"
)

var (
	newCmd = cmds.App.SubCommand("new", "create a new application")
)

func init() {
	newCmd.Command = build_app
}

func build_app() error {
	mainF := tpls.OpenFile("main.go")
	defer mainF.Close()
	logx.AssertError(tpls.T("main").Execute(mainF, tpls.Params()))

	cfgF := tpls.OpenFile("cfg", "cfg.go")
	defer cfgF.Close()
	logx.AssertError(tpls.T("cfg/cfg").Execute(cfgF, tpls.Params()))

	logx.AssertError(tpls.GoFmt("."))
	logx.AssertError(tpls.GoInit(*cmds.RepoName))
	logx.AssertError(tpls.GoModtidy())
	return nil
}
