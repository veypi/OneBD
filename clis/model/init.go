//
// init.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-22 16:16
// Distributed under terms of the MIT license.
//

package model

import (
	"regexp"

	"github.com/veypi/OneBD/clis/cmds"
)

var (
	newCmd = cmds.Model.SubCommand("new", "create a new model struct")
	genCmd = cmds.Model.SubCommand("gen", "generate model code from model struct")
)

var (
	nameObj   = newCmd.String("n", "user", "target model name")
	nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\._]*$`)
)

var methodReg = regexp.MustCompile(`methods:"([^"]+)"`)
var allowedMethods = []string{
	"Get", "List", "Post", "Put",
	"Patch", "Delete"}

func init() {
	newCmd.Command = new_model
	genCmd.Command = gen_model
}
