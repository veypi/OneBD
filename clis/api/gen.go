//
// gen.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-23 17:48
// Distributed under terms of the MIT license.
//

package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

func gen_model() error {
	if *fromObj == "" {
		return fmt.Errorf("model file or dir path is required")
	}
	absPath := utils.PathJoin(*cmds.DirRoot, *fromObj)
	if !utils.FileExists(absPath) {
		return fmt.Errorf("file or dir not exists: %v", absPath)
	}
	logx.Info().Msgf("\n%s\n%s\n%s", absPath, *cmds.DirRoot, *fromObj)
	var err error
	if utils.PathIsDir(absPath) {
		err = gen_from_dir(absPath)
	} else {
		err = gen_from_file(absPath)
	}
	return err
}

func gen_from_dir(fragments ...string) error {
	absPath := utils.PathJoin(append([]string{*cmds.DirRoot, *cmds.DirModel}, fragments...)...)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(absPath, entry.Name())
		if entry.IsDir() {
			err = gen_from_dir(append(fragments, entry.Name())...)
		} else if !strings.HasSuffix(fullPath, "_gen.go") {
			err = gen_from_file(append(fragments, entry.Name())...)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func gen_from_file(fragments ...string) error {
	return nil
}

func gen_from_struct(fragments ...string) error {
	return nil
}
