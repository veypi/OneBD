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
	"github.com/veypi/utils/logv"
)

func GoFmt(fp string) {
	_, err := run("gofmt", "-w", fp)
	if err != nil {
		logv.Warn().Msgf("gofmt -w %s failed", fp)
	}
}

func GoInit(name string) {
	_, err := run("go", "mod", "init", name)
	if err != nil {
		logv.Warn().Msgf("go mod init %s failed", name)
	}
}

func GoModtidy() {
	// TODO: test code
	run("bash", "-c", "echo '\nreplace github.com/veypi/OneBD => ../../workspace/OneBD/\n' >> go.mod")
	run("bash", "-c", "echo '\nreplace github.com/veypi/utils => ../../workspace/utils/\n' >> go.mod")
	_, err := run("go", "mod", "tidy")
	if err != nil {
		logv.Warn().Msgf("go mod tidy failed: %v", err)
	}
}

func run(acts ...string) (string, error) {
	if len(acts) == 0 {
		return "", nil
	}
	logv.Trace().Msgf("run %s", strings.Join(acts, " "))
	cmd := exec.Command(acts[0], acts[1:]...)
	cmd.Dir = *cmds.DirRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}
