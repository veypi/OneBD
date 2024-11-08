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
	"strings"

	"github.com/veypi/OneBD/obd/cmds"
	"github.com/veypi/OneBD/obd/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
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
	logv.Debug().Msgf("gen model from %v", root)
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
	fast := logv.AssertFuncErr(tpls.NewAst(fname))

	initPath := utils.PathJoin(*cmds.DirRoot, *cmds.DirModel, "init.go")
	initAst, err := tpls.NewAst(initPath)
	if err != nil {
		return err
	}
	initStructs := initAst.GetAllStructs()
	// repo := utils.PathJoin(*cmds.RepoName, filepath.Dir(strings.ReplaceAll(fname, *cmds.DirRoot, "")))
	migratePath := filepath.Join(filepath.Dir(fname), "init.go")
	var migrateAst *tpls.Ast
	if utils.FileExists(migratePath) {
		migrateAst, err = tpls.NewAst(migratePath)
		if err != nil {
			return err
		}
	} else {
		migrateAst = tpls.NewEmptyAst(fast.Name.Name)
	}
	fAbsPath := filepath.Join(filepath.Dir(fname), strings.ReplaceAll(filepath.Base(fname), ".go", ".gen.go"))
	genAst := logv.AssertFuncErr(tpls.NewFileOrEmptyAst(fAbsPath, filepath.Base(filepath.Dir(fname))))
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
			err = addMigrator(typeSpec.Name.Name, migrateAst)
			if err != nil {
				return err
			}

			// 遍历结构体字段，提取带tag的字段
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					sName := getStructName(field.Type)
					styp := utils.SliceGet(initStructs, func(s *tpls.SimpleStruct) bool {
						return s.Name == sName
					})
					if styp == nil {
						logv.Debug().Msgf("not found struct: %v", field.Type)
					} else {
						for _, subF := range (*styp).Fields {
							parseTag(subF, typeSpec.Name.Name, genAst)
						}
					}
				}
				parseTag(field, typeSpec.Name.Name, genAst)
			}
		}
	}
	err = migrateAst.Dump(migratePath)
	if err != nil {
		return err
	}

	return genAst.Dump(fAbsPath)
}

func parseTag(field *ast.Field, Obj string, genAst *tpls.Ast) {
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
			if m == "" {
				continue
			}
			method := m
			typ := field.Type
			if strings.HasPrefix(method, "*") {
				method = method[1:]
				if _, ok := typ.(*ast.StarExpr); !ok {
					typ = &ast.StarExpr{X: typ}
				}
			}
			method = utils.ToTitle(method)
			if !utils.InList(method, allowedMethods) {
				logv.Warn().Msgf("method %s not allowed", method)
				continue
			}
			tag := strings.ReplaceAll(field.Tag.Value, res[0], "")
			if field.Names[0].Name == "ID" {
				// 如果字段名是ID，若没有显性标注path别名，则自动将tag中的`parse:"path"`替换为`parse:"path@obj_id"`
				tag = strings.ReplaceAll(tag, `parse:"path"`, fmt.Sprintf(`parse:"path@%s_id"`, utils.CamelToSnake(Obj)))
			}
			if method == "Get" || method == "List" {
				// 如果请求是get，list，将json替换为query
				tag = strings.ReplaceAll(tag, `parse:"json`, `parse:"query`)
			}
			genAst.AddStructWithFields(Obj+method,
				&ast.Field{
					Names: []*ast.Ident{{Name: field.Names[0].Name}},
					Type:  typ,
					Tag:   &ast.BasicLit{Value: tag},
				})
			// 检查字段类型并添加相应的import路径
			checkAndAddImport(field.Type, genAst)
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

// checkAndAddImport 检查字段类型并根据需要添加相应的import路径
func checkAndAddImport(expr ast.Expr, fAst *tpls.Ast) {
	switch t := expr.(type) {
	case *ast.Ident:
		// 基本类型无需导入
	case *ast.SelectorExpr:
		// 如果是SelectorExpr，可能是像time.Time这样的类型
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			importPath := getImportPath(pkgIdent.Name)
			if importPath != "" {
				fAst.AddImport(importPath)
			}
		}
	case *ast.StarExpr:
		// 指针类型，递归处理
		checkAndAddImport(t.X, fAst)
	case *ast.ArrayType:
		// 数组类型，递归处理
		checkAndAddImport(t.Elt, fAst)
	case *ast.MapType:
		// map类型，递归处理键和值的类型
		checkAndAddImport(t.Key, fAst)
		checkAndAddImport(t.Value, fAst)
	case *ast.ChanType:
		// 通道类型，递归处理元素类型
		checkAndAddImport(t.Value, fAst)
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

func addMigrator(sName string, fAst *tpls.Ast) error {
	fcMigrate := fAst.GetMethod("init")
	if fcMigrate == nil {
		fcMigrate = &ast.FuncDecl{
			Name: ast.NewIdent("init"),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{},
			},
			Type: &ast.FuncType{},
		}
		fAst.Decls = append(fAst.Decls, fcMigrate)
	}
	strFc, err := tpls.Ast2Str(fcMigrate)
	if err != nil {
		return err
	}
	if strings.Contains(strFc, fmt.Sprintf("&%s{}", sName)) {
		return nil
	}
	logv.Debug().Msgf("add auto migrate object: %s", sName)
	var item ast.Stmt = &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("cfg.ObjList")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("append"),
				Args: []ast.Expr{
					ast.NewIdent("cfg.ObjList"),
					ast.NewIdent(fmt.Sprintf("&%s{}", sName)),
				},
			}},
	}
	fcMigrate.Body.List = utils.InsertAt(fcMigrate.Body.List, -1, item)
	fAst.AddImport(*cmds.RepoName + "/cfg")
	return nil
}
