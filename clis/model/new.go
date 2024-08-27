//
// model.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 17:12
// Distributed under terms of the MIT license.
//

package model

import (
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

func new_model() error {
	if !nameRegex.MatchString(*nameObj) {
		panic("invalid name")
	}
	if strings.Count(*nameObj, ".") > 1 {
		panic("invalid name")
	}
	fragment := strings.Split(*nameObj, "/")
	fragment = append([]string{*cmds.DirModel}, fragment...)
	name := fragment[len(fragment)-1]
	fname := utils.CamelToSnake(name)
	fragment[len(fragment)-1] = fname + ".go"
	if strings.Contains(fname, ".") {
		tmps := strings.Split(fname, ".")
		// 修改生成对象名和文件名
		fragment[len(fragment)-1] = tmps[0] + ".go"
		fname = tmps[1]
	}
	fAbsPath := utils.PathJoin(append([]string{*cmds.DirRoot}, fragment...)...)
	var err error
	if !utils.FileExists(fAbsPath) {
		fObj := tpls.OpenAbsFile(fAbsPath)
		defer fObj.Close()
		err = tpls.T("models", "obj").Execute(fObj, tpls.Params().With("package", fragment[len(fragment)-2]).With("Obj", utils.SnakeToCamel(fname)))
	} else {
		fAst := logx.AssertFuncErr(tpls.NewAst(fAbsPath))
		logx.AssertError(fAst.AddStructIfNotExist(utils.SnakeToCamel(fname),
			tpls.NewAstField("CreatedAt", "time.Time", "`json:\"created_at\" methods:\"get,post,put,patch,list,delete\"`"),
			tpls.NewAstField("UpdatedAt", "time.Time", "`json:\"updated_at\" methods:\"get,list\"`"),
			tpls.NewAstField("DeletedAt", "gorm.DeletedAt", "`json:\"deleted_at\" gorm:\"index\"`"),
		))
		err = fAst.Dump(fAbsPath)
	}
	return err
}
