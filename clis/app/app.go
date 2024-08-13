//
// app.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 19:51
// Distributed under terms of the MIT license.
//

package app

import (
	"errors"
	"fmt"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

var (
	newCmd = cmds.App.SubCommand("new", "create a new application")
)

func init() {
	newCmd.Command = build_app
}

func build_app() error {
	fname := utils.PathJoin(*cmds.DirPath, "main.go")
	if utils.FileExists(fname) && !*cmds.ForceWrite {
		return errors.New(fmt.Sprintf("file %s exists, use -force to overwrite", fname))
	}
	logx.Info().Msgf("auto generate %s", fname)
	fObj := logx.AssertFuncErr(utils.MkFile(fname))
	defer fObj.Close()
	logx.AssertError(tpls.T("main").Execute(fObj, tpls.Params()))
	logx.AssertError(tpls.GoFmt(fname))
	logx.AssertError(tpls.GoInit(*cmds.RepoName))
	logx.AssertError(tpls.GoModtidy())
	return nil
}
