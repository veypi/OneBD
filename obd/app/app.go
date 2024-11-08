//
// app.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 19:51
// Distributed under terms of the MIT license.
//

package app

import (
	"github.com/veypi/OneBD/obd/cmds"
	"github.com/veypi/OneBD/obd/tpls"
	"github.com/veypi/utils/logv"
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
	logv.AssertError(tpls.T("main").Execute(mainF, tpls.Params()))

	cfgF := tpls.OpenFile("cfg", "cfg.go")
	defer cfgF.Close()
	logv.AssertError(tpls.T("cfg/cfg").Execute(cfgF, tpls.Params()))

	dbF := tpls.OpenFile("cfg", "db.go")
	defer dbF.Close()
	logv.AssertError(tpls.T("cfg", "db").Execute(dbF, tpls.Params()))

	apiF := tpls.OpenFile(*cmds.DirApi, "init.go")
	defer apiF.Close()
	logv.AssertError(tpls.T("api", "init").Execute(apiF, tpls.Params().With("package", *cmds.DirApi)))

	modelF := tpls.OpenFile(*cmds.DirModel, "init.go")
	defer modelF.Close()
	logv.AssertError(tpls.T("models", "init").Execute(modelF, tpls.Params()))

	devyaml := tpls.OpenFile("cfg", "dev.yaml")
	defer devyaml.Close()
	logv.AssertError(tpls.T("cfg", "dev.yaml").Execute(devyaml, tpls.Params()))

	gitignore := tpls.OpenFile(".gitignore")
	defer gitignore.Close()
	logv.AssertError(tpls.T("gitignore").Execute(gitignore, tpls.Params()))

	makefile := tpls.OpenFile("Makefile")
	defer makefile.Close()
	logv.AssertError(tpls.T("Makefile").Execute(makefile, tpls.Params()))

	tpls.GoFmt(".")
	tpls.GoInit(*cmds.RepoName)
	tpls.GoModtidy()
	return nil
}
