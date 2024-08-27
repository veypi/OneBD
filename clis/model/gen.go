//
// gen.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-23 18:29
// Distributed under terms of the MIT license.
//

package model

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

var (
	targetObj = genCmd.String("t", "", "target model path or dir")
)

func gen_model() error {
	root := utils.PathJoin(*cmds.DirRoot, *cmds.DirModel)
	if *targetObj != "" {
		root = utils.PathJoin(root, *targetObj)
	}
	if !utils.FileExists(root) {
		return fmt.Errorf("file or dir not exists: %v", root)
	}
	var err error
	if utils.PathIsDir(root) {
		err = gen_from_dir(root)
	} else {
		err = gen_from_file(root)
	}
	return err
}

func gen_from_dir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err = gen_from_dir(fullPath)
		} else if !strings.HasSuffix(fullPath, "_gen.go") {
			err = gen_from_file(fullPath)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func gen_from_file(fname string) error {
	fast := logx.AssertFuncErr(tpls.NewAst(fname))
	// 遍历AST，找到所有结构体
	newStructs := make(map[string]*ast.StructType)

	// 用于存储需要的import路径
	imports := map[string]bool{}
	for _, decl := range fast.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// 遍历结构体字段，提取带tag的字段
			for _, field := range structType.Fields.List {
				res := methodReg.FindStringSubmatch(field.Tag.Value)
				if len(res) == 0 {
					continue
				}
				for _, tag := range res[1:] {
					field.Tag.Value = strings.ReplaceAll(field.Tag.Value, res[0], "")
					methods := strings.Split(strings.ReplaceAll(tag, " ", ""), ",")
					for _, m := range methods {
						method := utils.ToTitle(m)
						if !utils.InList(method, allowedMethods) {
							logx.Warn().Msgf("method %s not allowed", method)
							continue
						}
						if newStructs[typeSpec.Name.Name+method] == nil {
							newStructs[typeSpec.Name.Name+method] = &ast.StructType{Fields: &ast.FieldList{}}
						}
						newStructs[typeSpec.Name.Name+method].Fields.List = append(newStructs[typeSpec.Name.Name+method].Fields.List, field)
						// 检查字段类型并添加相应的import路径
						checkAndAddImport(field.Type, imports)
					}
				}
			}
		}
	}
	importsList := []string{}
	for a := range imports {
		importsList = append(importsList, a)
	}
	structNames := make([]string, 0, len(newStructs))
	for k := range newStructs {
		structNames = append(structNames, k)
	}
	sort.Strings(structNames)

	var decls []ast.Decl
	for _, t := range structNames {
		decls = append(decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{&ast.TypeSpec{
				Name: ast.NewIdent(t),
				Type: newStructs[t],
			}},
		})
	}
	structsBody := logx.AssertFuncErr(tpls.Ast2Str(decls))

	fObj := tpls.OpenAbsFile(filepath.Dir(fname), strings.ReplaceAll(filepath.Base(fname), ".go", "_gen.go"))
	defer fObj.Close()
	return tpls.T("models/gen").Execute(fObj, tpls.Params().With("body", structsBody).With("imports", importsList).With("package", filepath.Base(filepath.Dir(fname))))
}

// checkAndAddImport 检查字段类型并根据需要添加相应的import路径
func checkAndAddImport(expr ast.Expr, imports map[string]bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		// 基本类型无需导入
	case *ast.SelectorExpr:
		// 如果是SelectorExpr，可能是像time.Time这样的类型
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			importPath := getImportPath(pkgIdent.Name)
			if importPath != "" {
				imports[importPath] = true
			}
		}
	case *ast.StarExpr:
		// 指针类型，递归处理
		checkAndAddImport(t.X, imports)
	case *ast.ArrayType:
		// 数组类型，递归处理
		checkAndAddImport(t.Elt, imports)
	case *ast.MapType:
		// map类型，递归处理键和值的类型
		checkAndAddImport(t.Key, imports)
		checkAndAddImport(t.Value, imports)
	case *ast.ChanType:
		// 通道类型，递归处理元素类型
		checkAndAddImport(t.Value, imports)
	}
}

// getImportPath 根据类型名返回需要的import路径
func getImportPath(typeName string) string {
	// 你可以根据需要扩展这个映射
	importPaths := map[string]string{
		"time": "time",
		// 添加其他常用包的映射
	}

	return importPaths[typeName]
}
