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
	tpls.GoFmt(".")
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
		} else {
			err = gen_from_file(fullPath)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func gen_from_file(fname string) error {
	if strings.HasSuffix(fname, ".gen.go") || strings.HasSuffix(fname, "init.go") {
		return nil
	}
	fast := logx.AssertFuncErr(tpls.NewAst(fname))
	// 遍历AST，找到所有结构体
	newStructs := make(map[string][]*ast.Field)

	// 用于存储需要的import路径
	imports := map[string]bool{
		"encoding/json":               true,
		"github.com/veypi/OneBD/rest": true,
	}
	initStructs := getStructs(utils.PathJoin(*cmds.DirRoot, *cmds.DirModel, "init.go"))
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
				if len(field.Names) == 0 {
					styp := initStructs[getStructName(field.Type)]
					if styp == nil {
						logx.Debug().Msgf("not found struct: %v", field.Type)
					} else {
						for _, subF := range styp.Fields.List {
							parseTag(subF, typeSpec.Name.Name, newStructs, imports)
						}
					}
				}
				parseTag(field, typeSpec.Name.Name, newStructs, imports)
			}
		}
	}
	fAbsPath := filepath.Join(filepath.Dir(fname), strings.ReplaceAll(filepath.Base(fname), ".go", ".gen.go"))
	fAst := logx.AssertFuncErr(tpls.NewFileOrEmptyAst(fAbsPath, filepath.Base(filepath.Dir(fname))))
	for a := range imports {
		fAst.AddImport(a)
	}
	structNames := make([]string, 0, len(newStructs))
	for k := range newStructs {
		structNames = append(structNames, k)
	}
	sort.Strings(structNames)

	for _, t := range structNames {
		fAst.AddStructWithFields(t, newStructs[t]...)
		parseFc := createParseBody(t, newStructs[t])
		if parseFc != nil {
			fAst.AddOrReplaceStructMethods(t, "Parse", parseFc)
		}
		// break
	}
	return fAst.Dump(fAbsPath)
}

func parseTag(field *ast.Field, Obj string, newStructs map[string][]*ast.Field, imports map[string]bool) {
	if field.Tag == nil {
		return
	}
	res := methodReg.FindStringSubmatch(field.Tag.Value)
	if len(res) == 0 {
		return
	}
	for _, tag := range res[1:] {
		methods := strings.Split(strings.ReplaceAll(tag, " ", ""), ",")
		for _, m := range methods {
			method := utils.ToTitle(m)
			if !utils.InList(method, allowedMethods) {
				logx.Warn().Msgf("method %s not allowed", method)
				continue
			}
			if newStructs[Obj+method] == nil {
				newStructs[Obj+method] = make([]*ast.Field, 0, 4)
			}
			newStructs[Obj+method] = append(newStructs[Obj+method], &ast.Field{
				Names: []*ast.Ident{{Name: field.Names[0].Name}},
				Type:  field.Type,
				Tag:   &ast.BasicLit{Value: strings.ReplaceAll(field.Tag.Value, res[0], "")},
			})
			// 检查字段类型并添加相应的import路径
			checkAndAddImport(field.Type, imports)
		}
	}
}

func getStructName(expr ast.Expr) string {
	sName := ""
	switch t := expr.(type) {
	case *ast.Ident:
		sName = t.Name
	case *ast.SelectorExpr:
		sName = t.Sel.Name
	}
	return sName
}

func getStructs(fPath string) map[string]*ast.StructType {
	res := make(map[string]*ast.StructType)
	fAst, err := tpls.NewAst(fPath)
	if err != nil {
		return nil
	}
	fAst.Inspect(func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if styp, ok := typeSpec.Type.(*ast.StructType); ok {
				res[typeSpec.Name.Name] = styp
			}
		}
		return true
	})
	return res
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
