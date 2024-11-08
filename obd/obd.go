//
// obd.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 14:24
// Distributed under terms of the MIT license.
//

package main

import (
	_ "github.com/veypi/OneBD/obd/api"
	_ "github.com/veypi/OneBD/obd/app"
	"github.com/veypi/OneBD/obd/cmds"
	_ "github.com/veypi/OneBD/obd/model"
	_ "github.com/veypi/OneBD/obd/ts"
	"github.com/veypi/utils/logv"
)

func main() {
	// logv.DisableCaller()
	cmds.Parse()
	cmds.Main.Before = func() error {
		l, err := logv.ParseLevel(*cmds.LogLevel)
		if err != nil {
			return err
		}
		logv.SetLevel(l)
		return nil
	}
	err := cmds.Main.Run()
	if err != nil {
		logv.Warn().Msg(err.Error())
	}
}
