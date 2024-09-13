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
	// 遍历AST，找到所有结构体
	newStructs := make(map[string][]*ast.Field)

	// 用于存储需要的import路径
	imports := map[string]bool{}
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
					styp := initStructs[getStructName(field.Type)]
					if styp == nil {
						logv.Debug().Msgf("not found struct: %v", field.Type)
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
	err = migrateAst.Dump(migratePath)
	if err != nil {
		return err
	}
	fAbsPath := filepath.Join(filepath.Dir(fname), strings.ReplaceAll(filepath.Base(fname), ".go", ".gen.go"))
	fAst := logv.AssertFuncErr(tpls.NewFileOrEmptyAst(fAbsPath, filepath.Base(filepath.Dir(fname))))
	structNames := make([]string, 0, len(newStructs))
	for k := range newStructs {
		structNames = append(structNames, k)
	}
	sort.Strings(structNames)

	for _, t := range structNames {
		fAst.AddStructWithFields(t, newStructs[t]...)
		// parseFc := createParseBody(t, newStructs[t], imports)
		// if parseFc != nil {
		// 	fAst.AddStructMethods(t, "Parse", parseFc, true)
		// }
		// break
	}
	for a := range imports {
		fAst.AddImport(a)
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
			if newStructs[Obj+method] == nil {
				newStructs[Obj+method] = make([]*ast.Field, 0, 4)
			}
			tag := strings.ReplaceAll(field.Tag.Value, res[0], "")
			if field.Names[0].Name == "ID" {
				// 如果字段名是ID，若没有显性标注path别名，则自动将tag中的`parse:"path"`替换为`parse:"path@obj_id"`
				tag = strings.ReplaceAll(tag, `parse:"path"`, fmt.Sprintf(`parse:"path@%s_id"`, utils.CamelToSnake(Obj)))
			}
			newStructs[Obj+method] = append(newStructs[Obj+method], &ast.Field{
				Names: []*ast.Ident{{Name: field.Names[0].Name}},
				Type:  typ,
				Tag:   &ast.BasicLit{Value: tag},
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
