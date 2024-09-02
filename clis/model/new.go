//
// model.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 17:12
// Distributed under terms of the MIT license.
//

package model

import (
	"fmt"
	"go/ast"
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
	isRoot := true
	isSubResource := ""
	if len(fragment) > 1 {
		isRoot = false
	}
	fragment = append([]string{*cmds.DirModel}, fragment...)
	name := fragment[len(fragment)-1]
	fname := utils.CamelToSnake(name)
	fragment[len(fragment)-1] = fname + ".go"
	if strings.Contains(fname, ".") {
		tmps := strings.Split(fname, ".")
		// 修改生成对象名和文件名
		fragment[len(fragment)-1] = tmps[0] + ".go"
		fname = tmps[1]
		isSubResource = tmps[0]
	}
	fAbsPath := utils.PathJoin(append([]string{*cmds.DirRoot}, fragment...)...)
	var fAst *tpls.Ast
	if utils.FileExists(fAbsPath) {
		fAst = logx.AssertFuncErr(tpls.NewAst(fAbsPath))
	} else {
		fAst = tpls.NewEmptyAst(fragment[len(fragment)-2])
		if !isRoot {
			fAst.AddImport(fmt.Sprintf("%s/%s", *cmds.RepoName, *cmds.DirModel))
		}
	}
	args := make([]*ast.Field, 0, 5)
	if isRoot {
		args = append(args, tpls.NewAstField("", "BaseModel", ""))
	} else {
		args = append(args, tpls.NewAstField("", fmt.Sprintf("%s.BaseModel", *cmds.DirModel), ""))
	}
	if isSubResource != "" {
		args = append(args, tpls.NewAstField(fmt.Sprintf("%sID", utils.SnakeToCamel(isSubResource)), "string",
			fmt.Sprintf("`json:\"%s_id\" gorm:\"primaryKey;type:varchar(32)\" methods:\"get,post,put,patch,list,delete\" parse:\"path\"`", isSubResource)))
	}
	logx.AssertError(fAst.AddStructWithFields(utils.SnakeToCamel(fname),
		args...,
	))
	return fAst.Dump(fAbsPath)
}
