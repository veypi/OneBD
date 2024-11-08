//
// cmds.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 15:35
// Distributed under terms of the MIT license.
//

package cmds

import (
	"github.com/veypi/utils/flags"
)

var (
	Main  = flags.New("obd", "OneBD command line tools")
	Model = Main.SubCommand("model", "generate code")
	App   = Main.SubCommand("app", "generate application code")
	Tpl   = Main.SubCommand("tpl", "template tools")
	Api   = Main.SubCommand("api", "api tools")
	Ts    = Main.SubCommand("ts", "typescript generate tools")
)

var (
	RepoName   = Main.String("repo", "app", "repository name, for example: app,github.com/veypi/OneBD")
	DirRoot    = Main.String("dir", "./", "repo root dir")
	DirApi     = Main.String("dirapi", "api", "api dir")
	DirModel   = Main.String("dirmodel", "models", "model dir")
	ForceWrite = Main.Bool("y", false, "force to overwrite file")
)

var LogLevel = Main.String("l", "debug", "log level: trace|debug|info|warn|error|fatal")

func Parse() {
	Main.Parse()
}
