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

	"github.com/veypi/OneBD/obd/cmds"
	"github.com/veypi/OneBD/obd/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

var (
	embeded = newCmd.String("e", "BaseModel", "embeded struct field, should be defined in models/init.go")
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
		fAst = logv.AssertFuncErr(tpls.NewAst(fAbsPath))
	} else {
		fAst = tpls.NewEmptyAst(fragment[len(fragment)-2])
		if !isRoot {
			fAst.AddImport(fmt.Sprintf("%s/%s", *cmds.RepoName, *cmds.DirModel))
		}
	}
	args := make([]*ast.Field, 0, 5)
	if isRoot {
		if *embeded != "" {
			args = append(args, tpls.NewAstField("", *embeded, ""))
		}
	} else {
		if *embeded != "" {
			args = append(args, tpls.NewAstField("", fmt.Sprintf("%s.%s", *cmds.DirModel, *embeded), ""))
		}
	}
	if isSubResource != "" {
		ownerObj := utils.SnakeToCamel(isSubResource)
		args = append(args, tpls.NewAstField(fmt.Sprintf("%sID", ownerObj), "string",
			fmt.Sprintf("`json:\"%s_id\" methods:\"get,list,post,put,patch,delete\" parse:\"path\"`", isSubResource)))
		args = append(args, tpls.NewAstField(ownerObj, "*"+ownerObj,
			fmt.Sprintf("`json:\"%s\"`", isSubResource)))
	}
	args = append(args, tpls.NewAstField("Name", "string", "`json:\"name\" methods:\"post,put,*patch,*list\" parse:\"json\"`"))
	if *nameObj == "demo" {
		args = append(args, tpls.NewAstField("QueryA", "string", "`json:\"query_a\" methods:\"post,*list\" parse:\"query\"`"))
		args = append(args, tpls.NewAstField("HeaderB", "string", "`json:\"header_b\" methods:\"post,*list\" parse:\"header\"`"))
	}
	logv.AssertError(fAst.AddStructWithFields(utils.SnakeToCamel(fname),
		args...,
	))
	return fAst.Dump(fAbsPath)
}
