// init.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-10-09 15:19
// Distributed under terms of the MIT license.
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
	nameRegex     = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\._]*$`)
	parseReg      = regexp.MustCompile(`parse:"(path|query|header|json|form)(@\w+)?"`)
	jsonTagReg    = regexp.MustCompile(`json:"([\w-]+)?`)
	objReg        = regexp.MustCompile(`(\w+)?(Get|List|Post|Put|Patch|Delete)$`)
	globalStructs = make(map[string][]*ast.Field)
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
	initAst := logv.AssertFuncErr(tpls.NewAst(utils.PathJoin(*cmds.DirRoot, *cmds.DirModel, "init.go")))
	for _, obj := range initAst.GetAllStructs() {
		globalStructs[obj.Name] = obj.Fields
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
	if !utils.FileExists(utils.PathJoin(*tsDir, "webapi.ts")) {
		apiF := tpls.OpenAbsFile(*tsDir, "webapi.ts")
		defer apiF.Close()
		return tpls.T("ts", "webapi").Execute(apiF, tpls.Params())
	}
	return nil
}

func gen_from_dir(dir string, fragments ...string) error {
	absPath := utils.PathJoin(append([]string{*cmds.DirRoot, *cmds.DirModel}, fragments...)...)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}
	indexPath := utils.PathJoin(append(append([]string{*tsDir}, fragments...), "index.ts")...)
	indextpl := logv.AssertFuncErr(tpls.NewTsTpl(indexPath))

	for _, entry := range entries {
		ename := entry.Name()
		fullPath := filepath.Join(absPath, ename)
		if entry.IsDir() {
			err = gen_from_dir(fullPath, append(fragments, ename)...)
			indextpl.AddImport(fmt.Sprintf(`import %s from "./%s"`, ename, ename))
			indextpl.AddDefaultExport(ename)
		} else {
			if strings.HasSuffix(fullPath, ".gen.go") {
				obj := ename[:len(ename)-7]
				err = gen_from_gen_file(fullPath, append(fragments, obj)...)
				indextpl.AddImport(fmt.Sprintf(`import * as %s from "./%s"`, obj, obj))
				indextpl.AddDefaultExport(obj)
			} else if strings.HasSuffix(fullPath, "init.go") {
			} else if strings.HasSuffix(fullPath, ".go") {
				err = gen_from_model_file(fullPath, append(fragments, ename[:len(ename)-3])...)
			}
		}
		if err != nil {
			return err
		}
	}
	return indextpl.Dump()
}

var typMap = map[string]string{
	"uint":   "number",
	"int":    "number",
	"float":  "number",
	"string": "string",
	"[]byte": "string",
	"Time":   "Date",
}

func get_ts_fields_from_struct(origin []*ast.Field, src string) [][2]string {
	fields := make([][2]string, 0, 8)
	for _, f := range origin {
		if len(f.Names) != 0 {
			typ := "any"
			name := utils.CamelToSnake(f.Names[0].String())
			// src 为空时构造返回数据， 不为空时构造请求数据
			if src != "" {
				tags := parseReg.FindStringSubmatch(f.Tag.Value)
				if len(tags) < 2 || tags[1] != src {
					continue
				}
				if len(tags) > 2 && tags[2] != "" {
					name = tags[2][1:]
				}
			} else {
				tags := jsonTagReg.FindStringSubmatch(f.Tag.Value)
				if len(tags) > 1 && tags[1] != "" {
					if tags[1] == "-" {
						continue
					} else {
						name = tags[1]
					}
				}
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
			fields = append(fields, [2]string{name, typ})
		} else if t, ok := f.Type.(*ast.Ident); ok {
			if gobj, ok := globalStructs[t.Name]; ok {
				fields = append(fields, get_ts_fields_from_struct(gobj, src)...)
			}
		}
	}
	return fields
}

// generate model file
func gen_from_model_file(fname string, fragments ...string) error {
	fast := logv.AssertFuncErr(tpls.NewAst(fname))
	tstpl, err := tpls.NewTsTpl(utils.PathJoin(*tsDir, "models.ts"))
	if err != nil {
		return err
	}
	for _, obj := range fast.GetAllStructs() {
		// 生成请求opts
		tstpl.AddInterface(obj.Name, get_ts_fields_from_struct(obj.Fields, "")...)
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
	for _, obj := range fast.GetAllStructs() {
		pathTags := make([]string, 0, 8)
		for _, p := range fragments {
			if len(pathTags) > 0 && pathTags[len(pathTags)-1] == p {
				// 连续相同资源名跳过
				continue
			}
			pathTags = append(pathTags, p)
		}
		objFunc := obj.Name
		objName := obj.Name
		method := "Get"
		ownerObj := utils.SnakeToCamel(fragments[len(fragments)-1])
		isListReponse := false
		if temp := objReg.FindStringSubmatch(obj.Name); len(temp) == 3 {
			method = temp[2]
			if method == "List" {
				isListReponse = true
				method = "Get"
			}
			objName = temp[1]
			if temp[1] != ownerObj {
				// 2级资源
				pathTags = append(pathTags, fmt.Sprintf("${%s_id}", utils.CamelToSnake(ownerObj)), utils.CamelToSnake(temp[1]))
			} else {
				objFunc = strings.Replace(objFunc, ownerObj, "", 1)
			}
		} else {
			pathTags = append(pathTags, obj.Name)
		}
		args := make([]string, 0, 5)
		resp := make([]string, 0, 4)
		if fields := get_ts_fields_from_struct(obj.Fields, "path"); len(fields) > 0 {
			for _, f := range fields {
				args = append(args, fmt.Sprintf("%s: %s", f[0], f[1]))
				tmpf := fmt.Sprintf("${%s}", f[0])
				if !utils.InList(tmpf, pathTags) {
					pathTags = append(pathTags, tmpf)
				}
			}
		}
		if fields := get_ts_fields_from_struct(obj.Fields, "form"); len(fields) > 0 {
			tstpl.AddInterface(objFunc+"Opts", fields...)
			args = append(args, fmt.Sprintf("form: %sOpts", objFunc))
			resp = append(resp, "form")
		}
		if fields := get_ts_fields_from_struct(obj.Fields, "json"); len(fields) > 0 {
			tstpl.AddInterface(objFunc+"Opts", fields...)
			args = append(args, fmt.Sprintf("json: %sOpts", objFunc))
			resp = append(resp, "json")
		}
		if fields := get_ts_fields_from_struct(obj.Fields, "query"); len(fields) > 0 {
			tstpl.AddInterface(objFunc+"Query", fields...)
			args = append(args, fmt.Sprintf("query: %sQuery", objFunc))
			resp = append(resp, "query")
		}
		if fields := get_ts_fields_from_struct(obj.Fields, "header"); len(fields) > 0 {
			tstpl.AddInterface(objFunc+"Header", fields...)
			args = append(args, fmt.Sprintf("header: %sOpts", objFunc))
			resp = append(resp, "header")
		}
		tstpl.AddFunc(objFunc, tpls.Params().
			With("method", method).
			With("name", objFunc).
			With("obj", objName).
			With("is_list", isListReponse).
			With("url", "/"+strings.Join(pathTags, "/")).
			With("args", strings.Join(args, ", ")).
			With("resp", strings.Join(resp, ", ")),
			"snaps", "ts", "api")
	}

	return tstpl.Dump()
}
