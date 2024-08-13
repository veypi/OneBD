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

var CMD = flags.New("{{.repo}}", "the backend server of main")
var CfgDump = CMD.SubCommand("cfg", "generate cfg file")

var configFile = CMD.String("f", "./dev.yml", "the config file")

var (
	host = CMD.String("host", "", "host")
	port = CMD.Int("port", 0, "port")
)

func init() {
	CMD.Before = func() error {
		err := LoadCfg()
		if err != nil {
			return err
		}
		if *host != "" {
			Config.Host = *host
		}
		if *port != 0 {
			Config.Port = *port
		}
		return nil
	}
	CfgDump.Command = DumpCfg
}

func LoadCfg() error {
	flags.LoadCfg(*configFile, Config)
	return nil
}

func DumpCfg() error {
	logx.Debug().Msgf("%v", *Config)
	flags.DumpCfg(*configFile, Config)
	return nil
}

