//
// tpls.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 18:39
// Distributed under terms of the MIT license.
//

package tpls

import (
	"embed"
	"html/template"
	"os"
	"time"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

var (
	tplDir = cmds.Main.String("tpl", "./.tpls", "template dir")
)

type params map[string]any

var Params = func() params {
	m := make(params)
	m["common"] = map[string]any{
		"repo":   *cmds.RepoName,
		"date":   time.Now().Format("2006-01-02 15:04:05"),
		"noedit": "// Auto generated by OneBD. DO NOT EDIT",
	}
	return m
}

func (p params) WithData(t map[string]any) params {
	for k, v := range t {
		p[k] = v
	}
	return p
}

func (p params) With(k string, v any) params {
	p[k] = v
	return p
}

func T(p string) *template.Template {
	return loadTpl(p)
}

//go:embed templates
var tplFs embed.FS

func loadTpl(f string) *template.Template {
	f = f + ".tpl"
	body, err := tplFs.ReadFile("templates/" + f)
	if err != nil {
		logx.Warn().Msgf("load origin template %s error: %v\n%s\n\n------", f, err, body)
	}
	temp := utils.PathJoin(*tplDir, f)
	if utils.FileExists(temp) {
		logx.Info().Msgf("read template from %s", temp)
		body = logx.AssertFuncErr(os.ReadFile(temp))
	} else if err != nil {
		logx.Warn().Msgf("not found template %s", f)
	}
	tmpl := logx.AssertFuncErr(template.New("").Parse(string(body)))
	return tmpl
}
