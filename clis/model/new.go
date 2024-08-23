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
)

func new_model() error {
	if !nameRegex.MatchString(*nameObj) {
		panic("invalid name")
	}
	fragment := strings.Split(*nameObj, ".")
	name := fragment[len(fragment)-1]
	fragment = append([]string{*cmds.DirModel}, fragment...)
	fname := utils.CamelToSnake(name)
	fragment[len(fragment)-1] = fname + ".go"

	fObj := tpls.OpenFile(fragment...)
	defer fObj.Close()
	err := tpls.T("models", "obj").Execute(fObj, tpls.Params().With("package", fragment[len(fragment)-2]).With("Obj", utils.SnakeToCamel(fname)))
	return err
}
