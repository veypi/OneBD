//
// model.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 17:12
// Distributed under terms of the MIT license.
//

package gen

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

func init() {
	cmds.Model.Command = gen_model
}

var (
	nameObj = cmds.Model.String("n", "user", "target model name")
)

var (
	nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
)

func gen_model() error {
	dirPath := *cmds.DirPath
	name := []byte(*nameObj)
	name[0] = bytes.ToLower(name[:1])[0]
	fname := string(name)
	name[0] = bytes.ToUpper(name[:1])[0]
	objName := string(name)

	logx.Assert(utils.PathIsDir(dirPath), fmt.Sprintf("%s not exist or is not directory.", dirPath))
	logx.Assert(nameRegex.MatchString(*nameObj), fmt.Sprintf("invalid target name %s", name))
	fname = logx.AssertFuncErr(filepath.Abs(dirPath + "/models/" + fname + ".go"))
	if utils.FileExists(fname) && !*cmds.ForceWrite {
		return errors.New(fmt.Sprintf("file %s exists, use -force to overwrite", fname))
	}
	logx.Info().Msgf("generate <%s> Object in %s", objName, fname)
	fObj := logx.AssertFuncErr(utils.MkFile(fname))
	defer fObj.Close()
	defer tpls.GoFmt(fname)
	return tpls.T("models/obj").Execute(fObj, tpls.Params().With("obj", objName))
}
