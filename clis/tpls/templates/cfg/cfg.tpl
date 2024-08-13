//
// cfg.go
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package cfg

import (
	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils/flags"
	"github.com/veypi/utils/logx"
)

type config struct {
	rest.RestConf
}

var Config = &config{}

var CMD = flags.New("{{.common.repo}}", "the backend server of {{.common.repo}}")
var CfgDump = CMD.SubCommand("cfg", "generate cfg file")

var configFile = CMD.String("f", "./dev.yaml", "the config file")


func init() {
	CMD.StringVar(&Config.Host, "h", "0.0.0.0", "host")
	CMD.IntVar(&Config.Port, "p", 6000, "port")
	CMD.StringVar(&Config.LoggerLevel, "l", "info", "log level")
	CMD.Before = func() error {
		flags.LoadCfg(*configFile, Config)
		CMD.Parse()
		logx.SetLevel(logx.AssertFuncErr(logx.ParseLevel(Config.LoggerLevel)))
		return nil
	}
	CfgDump.Command = func() error {
		flags.DumpCfg(*configFile, Config)
		return nil
	}
}
