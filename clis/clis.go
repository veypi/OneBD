//
// clis.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 14:24
// Distributed under terms of the MIT license.
//

package main

import (
	_ "github.com/veypi/OneBD/clis/api"
	_ "github.com/veypi/OneBD/clis/app"
	"github.com/veypi/OneBD/clis/cmds"
	_ "github.com/veypi/OneBD/clis/model"
	"github.com/veypi/utils/logv"
)

func main() {
	// logv.DisableCaller()
	cmds.Main.Before = func() error {
		l, err := logv.ParseLevel(*cmds.LogLevel)
		if err != nil {
			return err
		}
		logv.SetLevel(l)
		return nil
	}
	cmds.Parse()
	err := cmds.Main.Run()
	if err != nil {
		logv.Warn().Msg(err.Error())
	}
}
