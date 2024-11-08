//
// cmds.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 15:35
// Distributed under terms of the MIT license.
//

package cmds

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/veypi/utils"
	"github.com/veypi/utils/flags"
	"github.com/veypi/utils/logv"
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
	RepoName   = Main.String("repo", "", "repository name, for example: app,github.com/veypi/OneBD, will auto detect from go.mod")
	DirRoot    = Main.String("dir", "./", "repo root dir")
	DirApi     = Main.String("dirapi", "api", "api dir")
	DirModel   = Main.String("dirmodel", "models", "model dir")
	ForceWrite = Main.Bool("y", false, "force to overwrite file")
)

var LogLevel = Main.String("l", "debug", "log level: trace|debug|info|warn|error|fatal")

func Parse() {
	Main.Parse()
	Main.Before = func() error {
		l, err := logv.ParseLevel(*LogLevel)
		if err != nil {
			return err
		}
		logv.SetLevel(l)
		if *RepoName == "" {
			absPath, err := filepath.Abs(*DirRoot)
			if err != nil {
				return err
			}
			*RepoName = filepath.Base(absPath)
			gomod := filepath.Join(absPath, "go.mod")
			if utils.FileExists(gomod) {
				name, err := getmodName()
				if err == nil && name != "" {
					*RepoName = name
				}
			}
		}

		return nil
	}
}

func getmodName() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	moduleInfo := out.String()
	moduleName := strings.TrimSpace(strings.Split(moduleInfo, " ")[0])
	return moduleName, nil
}
