//
// jsast.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-10-09 17:13
// Distributed under terms of the MIT license.
//

package tpls

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

func NewTsTpl(fPath string) (*simpleTsTemplate, error) {
	body := []byte{}
	var err error
	if utils.FileExists(fPath) {
		body, err = os.ReadFile(fPath)
		if err != nil {
			return nil, err
		}
	}
	s := &simpleTsTemplate{
		fPath: fPath,
		body:  string(body),
	}
	return s, nil
}

type simpleTsTemplate struct {
	fPath string
	body  string
}

func (t *simpleTsTemplate) Body() string {
	return t.body
}

func (t *simpleTsTemplate) Init(args params, p ...string) error {
	temp := &bytes.Buffer{}
	err := T(p...).Execute(temp, args)
	if err != nil {
		return err
	}
	t.body = temp.String()
	return nil
}

func (t *simpleTsTemplate) AddInterface(name string, fields ...[2]string) {
	tReg := regexp.MustCompile(fmt.Sprintf("(?ms:export interface %s .*?^})", name))
	res := tReg.FindString(t.body)
	if res == "" {
		customs := &bytes.Buffer{}
		logv.AssertError(T("snaps", "ts", "interface").Execute(customs, Params().With("fields", fields).With("name", name)))
		t.body += customs.String()
	} else {
		// t.body = strings.Replace(t.body, res, customs.String(), 1)
		newF := res
		for _, field := range fields {
			freg := fmt.Sprintf(".*%s:.*", strings.ReplaceAll(field[0], "?", `\?`))
			fold := regexp.MustCompile(freg).FindString(res)
			if fold == "" {
				newF = strings.Replace(newF, "}", fmt.Sprintf("  %s: %s\n}", field[0], field[1]), 1)
			} else {
				// not update exist field type
				// newF = strings.Replace(newF, fold, fmt.Sprintf("  %s: %s", field[0], field[1]), 1)
			}
		}
		t.body = strings.Replace(t.body, res, newF, 1)
	}
}

func (t *simpleTsTemplate) AddFunc(name string, ps params, p ...string) {
	tReg := regexp.MustCompile(fmt.Sprintf("(// keep.*\n)*(?ms:export function %s.*?^})", name))
	res := tReg.FindStringSubmatch(t.body)
	if len(res) == 0 {
		customs := &bytes.Buffer{}
		logv.AssertError(T(p...).Execute(customs, ps))
		t.body += customs.String()
	} else if res[1] != "" {
		// find // keep
	} else {
		customs := &bytes.Buffer{}
		logv.AssertError(T(p...).Execute(customs, ps))
		t.body = strings.Replace(t.body, res[0], customs.String(), 1)
	}
}

func (t *simpleTsTemplate) Dump(nPath ...string) error {
	p := t.fPath
	if len(nPath) > 0 && nPath[0] != "" {
		p = nPath[0]
	}
	t.body = strings.ReplaceAll(t.body, "\n\n\n", "\n\n")
	f, err := utils.MkFile(p)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(t.body))
	return err
}
