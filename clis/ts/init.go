//
// init.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-10-09 15:19
// Distributed under terms of the MIT license.
//
package ts

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

var (
	genCmd  = cmds.Ts.SubCommand("gen", "generate typescript api code from model.gen.go file")
	fromObj = genCmd.String("f", "", "target model file or dir path, relative to root cmd Dirmodel")
	tsDir   = genCmd.String("dir", "./ts", "target ts dir")
)

var (
	nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\._]*$`)
	parseReg  = regexp.MustCompile(`parse:"(path|query|header|json|form)(@\w+)?"`)
	objReg    = regexp.MustCompile(`(\w+)?(Get|List|Post|Put|Patch|Delete)$`)
)

func init() {
	genCmd.Command = gen_api
}

func gen_api() error {
	absPath := utils.PathJoin(*cmds.DirRoot, *cmds.DirModel, *fromObj)
	if !utils.FileExists(absPath) {
		return fmt.Errorf("file or dir not exists: %v", absPath)
	}
	var err error
	logv.Info().Msgf("auto generate typescript api files from model: %s", absPath)
	fragments := make([]string, 0)
	if *fromObj != "" {
		fragments = strings.Split(*fromObj, "/")
	}
	if utils.PathIsDir(absPath) {
		err = gen_from_dir(absPath, fragments...)
	} else {
		fname := fragments[len(fragments)-1]
		if !strings.HasSuffix(fname, ".go") {
			return fmt.Errorf("not support filetype %s", fname)
		}
		if strings.HasSuffix(fname, ".gen.go") {
			fragments[len(fragments)-1] = fname[:len(fname)-7]
			err = gen_from_gen_file(absPath, fragments...)
		} else {
			fragments[len(fragments)-1] = fname[:len(fname)-3]
			err = gen_from_model_file(absPath, fragments...)
		}
	}
	if err != nil {
		return err
	}
	apiF := tpls.OpenAbsFile(*tsDir, "webapi.ts")
	defer apiF.Close()
	return tpls.T("ts", "webapi").Execute(apiF, tpls.Params())
}

func gen_from_dir(dir string, fragments ...string) error {
	absPath := utils.PathJoin(append([]string{*cmds.DirRoot, *cmds.DirModel}, fragments...)...)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}
	indexPath := utils.PathJoin(append(append([]string{*tsDir}, fragments...), "index.ts")...)
	logv.Warn().Msgf("|%s", indexPath)

	for _, entry := range entries {
		ename := entry.Name()
		fullPath := filepath.Join(absPath, ename)
		if entry.IsDir() {
			err = gen_from_dir(fullPath, append(fragments, ename)...)
		} else {
			if strings.HasSuffix(fullPath, ".gen.go") {
				err = gen_from_gen_file(fullPath, append(fragments, ename[:len(ename)-7])...)
			} else if strings.HasSuffix(fullPath, "init.go") {
			} else if strings.HasSuffix(fullPath, ".go") {
				err = gen_from_model_file(fullPath, append(fragments, ename[:len(ename)-3])...)
			}
		}
		if err != nil {
			return err
		}
	}
	return err
}

var typMap = map[string]string{
	"uint":   "Number",
	"int":    "Number",
	"float":  "Number",
	"string": "String",
	"[]byte": "String",
	"Time":   "Date",
}

// generate model file
func gen_from_model_file(fname string, fragments ...string) error {
	fast := logv.AssertFuncErr(tpls.NewAst(fname))
	tstpl, err := tpls.NewTsTpl(utils.PathJoin(*tsDir, "models.ts"))
	if err != nil {
		return err
	}
	for n, obj := range fast.GetAllStructs() {
		fields := make([][2]string, 0, 8)
		for _, f := range obj.Fields.List {
			if len(f.Names) != 0 {
				typ := "any"
				name := utils.CamelToSnake(f.Names[0].String())
				tobj := f.Type
				if t, ok := tobj.(*ast.StarExpr); ok {
					name = name + "?"
					tobj = t.X
				}
				if t, ok := tobj.(*ast.Ident); ok {
					typ = t.Name
				} else if t, ok := tobj.(*ast.SelectorExpr); ok {
					typ = t.Sel.Name
				} else {
					logv.Warn().Msgf("not found %s %T %v", name, tobj, tobj)
				}
				if t, ok := typMap[typ]; ok {
					typ = t
				}
				fields = append(fields, [2]string{name, typ})
			}
		}
		// 生成请求opts
		tstpl.AddInterface(n, fields...)
	}
	return tstpl.Dump()
}

// generate api file
// tag标签 parse:"path/header/query/form/json" 可以追加为 path@alias_name
func gen_from_gen_file(fname string, fragments ...string) error {
	fast := logv.AssertFuncErr(tpls.NewAst(fname))
	tstpl, err := tpls.NewTsTpl(utils.PathJoin(append([]string{*tsDir}, fragments...)...) + ".ts")
	if err != nil {
		return err
	}
	if tstpl.Body() == "" {
		webapiPrefix := "./"
		if len(fragments) > 1 {
			webapiPrefix = strings.Repeat("../", len(fragments)-1)
		}
		err = tstpl.Init(tpls.Params().With("root", webapiPrefix), "ts", "api")
		if err != nil {
			return err
		}
	}
	for n, obj := range fast.GetAllStructs() {
		pathTags := make([]string, 0, 8)
		for _, p := range fragments {
			if len(pathTags) > 0 && pathTags[len(pathTags)-1] == p {
				// 连续相同资源名跳过
				continue
			}
			pathTags = append(pathTags, p)
		}
		objFunc := n
		objName := n
		method := "Get"
		ownerObj := utils.SnakeToCamel(fragments[len(fragments)-1])
		if temp := objReg.FindStringSubmatch(n); len(temp) == 3 {
			method = temp[2]
			if method == "List" {
				method = "Get"
			}
			objName = temp[1]
			if temp[1] != ownerObj {
				// 2级资源
				pathTags = append(pathTags, fmt.Sprintf("${%s_id}", utils.CamelToSnake(ownerObj)), utils.CamelToSnake(temp[1]))
			} else {
				objFunc = strings.Replace(objFunc, ownerObj, "", 1)
			}
		}
		fields := make(map[string][][2]string)
		for _, f := range obj.Fields.List {
			if len(f.Names) != 0 {
				tags := parseReg.FindStringSubmatch(f.Tag.Value)
				if len(tags) < 2 {
					continue
				}
				src := tags[1]
				typ := "any"
				name := utils.CamelToSnake(f.Names[0].String())
				if len(tags) > 2 && tags[2] != "" {
					name = tags[2][1:]
				}
				tobj := f.Type
				if t, ok := tobj.(*ast.StarExpr); ok {
					if src != "path" {
						// path 参数不允许为空
						name = name + "?"
					}
					tobj = t.X
				}
				if t, ok := tobj.(*ast.Ident); ok {
					typ = t.Name
				} else if t, ok := tobj.(*ast.SelectorExpr); ok {
					typ = t.Sel.Name
				} else {
					logv.Warn().Msgf("not found %s %T %v", name, tobj, tobj)
				}
				if t, ok := typMap[typ]; ok {
					typ = t
				}
				fields[src] = append(fields[src], [2]string{name, typ})
			}
		}
		// 生成请求opts
		args := make([]string, 0, 5)
		resp := make([]string, 0, 4)
		if fields["path"] != nil {
			for _, f := range fields["path"] {
				args = append(args, fmt.Sprintf("%s: %s", f[0], f[1]))
				tmpf := fmt.Sprintf("${%s}", f[0])
				if !utils.InList(tmpf, pathTags) {
					pathTags = append(pathTags, tmpf)
				}
			}
		}
		if fields["form"] != nil {
			tstpl.AddInterface(objFunc+"Opts", fields["form"]...)
			args = append(args, fmt.Sprintf("form: %sOpts", objFunc))
			resp = append(resp, "form")
		}
		if fields["json"] != nil {
			tstpl.AddInterface(objFunc+"Opts", fields["json"]...)
			args = append(args, fmt.Sprintf("json: %sOpts", objFunc))
			resp = append(resp, "json")
		}
		if fields["query"] != nil {
			tstpl.AddInterface(objFunc+"Query", fields["query"]...)
			args = append(args, fmt.Sprintf("query: %sQuery", objFunc))
			resp = append(resp, "query")
		}
		if fields["header"] != nil {
			tstpl.AddInterface(objFunc+"Header", fields["header"]...)
			args = append(args, fmt.Sprintf("header: %sOpts", objFunc))
			resp = append(resp, "header")
		}
		tstpl.AddFunc(objFunc, tpls.Params().
			With("method", method).
			With("name", objFunc).
			With("obj", objName).
			With("url", "/"+strings.Join(pathTags, "/")).
			With("args", strings.Join(args, ", ")).
			With("resp", strings.Join(resp, ", ")),
			"snaps", "ts", "api")
	}

	return tstpl.Dump()
}
