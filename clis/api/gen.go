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
	"github.com/veypi/utils/logx"
)

func gen_api() error {
	absPath := utils.PathJoin(*cmds.DirRoot, *cmds.DirModel, *fromObj)
	if !utils.FileExists(absPath) {
		return fmt.Errorf("file or dir not exists: %v", absPath)
	}
	var err error
	logx.Info().Msgf("auto generate api files from model: %s", absPath)
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
	return err
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
	if strings.HasSuffix(fname, ".gen.go") || strings.HasSuffix(fname, "init.go") {
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
	if utils.FileExists(genFname) {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, genFname, nil, parser.AllErrors)
		if err != nil {
			return err
		}
		objs_gen := make(map[string]*ast.StructType)
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
				objs_gen[typeSpec.Name.Name] = st

			}
		}
	}
	for name := range objs {
		tPath := utils.PathJoin(*cmds.DirRoot, *cmds.DirApi, strings.Join(fragments, "/"), utils.CamelToSnake(name)+".go")
		if !utils.FileExists(tPath) {
			packageName := filepath.Base(filepath.Dir(tPath))
			objName := utils.CamelToSnake(name)
			fObj := tpls.OpenAbsFile(tPath)
			defer fObj.Close()
			mimport := fmt.Sprintf(`"%s"`, strings.Join(append([]string{*cmds.RepoName, *cmds.DirModel}, fragments[:len(fragments)-1]...), "/"))
			err = tpls.T("api", "new").Execute(fObj, tpls.Params().
				With("mimport", mimport).
				With("package", packageName).
				With("s_obj", objName).
				With("obj", utils.SnakeToPrivateCamel(objName)).
				With("Obj", utils.SnakeToCamel(objName)))
			logx.AssertError(err)
			err = addRouter(false, append(fragments, objName))

		} else {
		}
		if err != nil {
			return err
		}
		// packagesName := utils.PathJoin(*cmds.DirRoot, *cmds.DirApi, strings.Join(fragments, "/"))
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
