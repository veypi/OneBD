//
// gen.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-23 17:48
// Distributed under terms of the MIT license.
//

package api

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

func gen_api() error {
	absPath := utils.PathJoin(*cmds.DirRoot, *cmds.DirModel, *fromObj)
	if !utils.FileExists(absPath) {
		return fmt.Errorf("file or dir not exists: %v", absPath)
	}
	var err error
	logv.Info().Msgf("auto generate api files from model: %s", absPath)
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
		fragments[len(fragments)-1] = fname[:len(fname)-3]
		err = gen_from_file(absPath, fragments...)
	}
	if err != nil {
		return err
	}
	return tpls.GoModtidy()
}

func gen_from_dir(dir string, fragments ...string) error {
	absPath := utils.PathJoin(append([]string{*cmds.DirRoot, *cmds.DirModel}, fragments...)...)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		ename := entry.Name()
		fullPath := filepath.Join(absPath, ename)
		if entry.IsDir() {
			err = gen_from_dir(fullPath, append(fragments, ename)...)
		} else {
			err = gen_from_file(fullPath, append(fragments, ename[:len(ename)-3])...)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func gen_from_file(fname string, fragments ...string) error {
	if !strings.HasSuffix(fname, ".go") || strings.HasSuffix(fname, ".gen.go") || strings.HasSuffix(fname, "init.go") {
		return nil
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fname, nil, parser.AllErrors)
	if err != nil {
		return err
	}
	objs := make(map[string]*ast.StructType)
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			st, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			objs[typeSpec.Name.Name] = st

		}
	}
	genFname := fname[:len(fname)-3] + ".gen.go"
	objs_gen := make(map[string]map[string]*ast.StructType)
	if utils.FileExists(genFname) {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, genFname, nil, parser.AllErrors)
		if err != nil {
			return err
		}
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				st, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				if objs := objReg.FindStringSubmatch(typeSpec.Name.Name); len(objs) > 1 {
					obj := objs[1]
					method := objs[2]
					if objs_gen[obj] == nil {
						objs_gen[obj] = make(map[string]*ast.StructType)
					}
					objs_gen[obj][method] = st
				}
			}
		}
	}
	for name := range objs {
		snakeName := utils.CamelToSnake(name)
		tPath := utils.PathJoin(*cmds.DirRoot, *cmds.DirApi, strings.Join(fragments, "/"), snakeName+".go")
		packageName := filepath.Base(filepath.Dir(tPath))
		fAst := tpls.NewEmptyAst(utils.CamelToSnake(packageName))
		if utils.FileExists(tPath) {
			fAst, err = tpls.NewAst(tPath)
			if err != nil {
				return err
			}
		}
		fAst.AddMethod(create_use_func("use"+name), false)
		fAst.AddImport("github.com/veypi/OneBD/rest")
		mimport := strings.Join(append([]string{*cmds.RepoName, *cmds.DirModel}, fragments[:len(fragments)-1]...), "/")
		fAst.AddImport(mimport, "M")
		fAst.AddImport(*cmds.RepoName + "/cfg")
		useFunc := fAst.GetMethod("use" + name)
		for m := range objs_gen[name] {
			fAst.AddMethod(create_api_func(fAst, name, m, objs_gen[name][m].Fields.List), false)
			fAst.AppendStmtInMethod(create_use_stmt(name, m), useFunc)
		}
		err = fAst.Dump(tPath)
		if err != nil {
			return err
		}
		err = addRouter(false, append(fragments, snakeName))
		if err != nil {
			return err
		}
	}

	// fObj := tpls.OpenFile(*cmds.DirApi, )
	// defer fObj.Close()
	// return tpls.T("models/gen").Execute(fObj, tpls.Params().With("structs", f.String()).With("imports", importsList).With("package", filepath.Base(filepath.Dir(fname))))
	return nil
}

// checkAndAddImport 检查字段类型并根据需要添加相应的import路径

func gen_from_struct(fragments ...string) error {
	return nil
}

func newApiFile(fragments ...string) error {
	fObj := tpls.OpenFile(fragments...)
	defer fObj.Close()
	return tpls.T("api", "init").Execute(fObj, tpls.Params().With("package", fragments[len(fragments)-2]))
	// return tpls.T("api", "new").Execute(fObj, tpls.Params().With("package", fragment[len(fragment)-2]).With("obj", utils.SnakeToPrivateCamel(fname)).With("Obj", utils.SnakeToCamel(fname)))
}
