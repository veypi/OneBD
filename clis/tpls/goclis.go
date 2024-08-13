//
// fmt.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 19:46
// Distributed under terms of the MIT license.
//

package tpls

import (
	"os/exec"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/utils/logx"
)

func GoFmt(fp string) error {
	_, err := run("gofmt", "-w", fp)
	return err
}

func GoInit(name string) error {
	_, err := run("go", "mod", "init", name)
	return err
}

func GoModtidy() error {
	_, err := run("go", "mod", "tidy")
	return err
}

func run(acts ...string) (string, error) {
	if len(acts) == 0 {
		return "", nil
	}
	logx.Debug().Msgf("run %s", strings.Join(acts, " "))
	cmd := exec.Command(acts[0], acts[1:]...)
	cmd.Dir = *cmds.DirPath
	out, err := cmd.CombinedOutput()
	return string(out), err
}
