//
// init.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-23 17:49
// Distributed under terms of the MIT license.
//

package api

import (
	"github.com/veypi/OneBD/clis/cmds"
	"regexp"
)

var (
	newCmd  = cmds.Api.SubCommand("new", "create new api code in api dir")
	nameObj = newCmd.String("n", "user", "target resource name")
	genCmd  = cmds.Api.SubCommand("gen", "generate api code from model files in api dir")
	fromObj = genCmd.String("f", "", "target model file or dir path, relative to root cmd Dirmodel")
)

var (
	nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\._]*$`)
)

func init() {
	newCmd.Command = new_api
	genCmd.Command = gen_api
}
