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
)

var (
	apiGen   = Main.SubCommand("api", "generate api code")
	jsGen    = Main.SubCommand("js", "generate js code")
	modelGen = Main.SubCommand("model", "generate a new object model")
)

var (
	RepoName   = Main.String("repo", "app", "repository name")
	DirPath    = Main.String("dir", "./", "dir to generate")
	ForceWrite = Main.Bool("y", false, "force to overwrite file")
)

var LogLevel = Main.String("l", "info", "log level: trace|debug|info|warn|error|fatal")

func Parse() {
	Main.Parse()
}
