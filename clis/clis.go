//
// clis.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 14:24
// Distributed under terms of the MIT license.
//

package main

import (
	_ "github.com/veypi/OneBD/clis/app"
	"github.com/veypi/OneBD/clis/cmds"
	_ "github.com/veypi/OneBD/clis/model"
	"github.com/veypi/utils/logx"
)

func main() {
	// logx.DisableCaller()
	cmds.Main.Before = func() error {
		l, err := logx.ParseLevel(*cmds.LogLevel)
		if err != nil {
			return err
		}
		logx.SetLevel(l)
		return nil
	}
	cmds.Parse()
	err := cmds.Main.Run()
	if err != nil {
		logx.Warn().Msg(err.Error())
	}
}
