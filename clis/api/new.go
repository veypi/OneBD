//
// new.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-23 12:03
// Distributed under terms of the MIT license.
//

package api

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

func new_api() error {
	if !nameRegex.MatchString(*nameObj) {
		panic("invalid name")
	}
	fragment := strings.Split(*nameObj, ".")
	name := fragment[len(fragment)-1]
	fname := utils.CamelToSnake(name)
	fragment[len(fragment)-1] = fname + ".go"

	fObj := tpls.OpenFile(fragment...)
	defer fObj.Close()
	packageName := *cmds.DirApi
	if len(fragment) > 1 {
		packageName = utils.CamelToSnake(fragment[len(fragment)-2])
	}
	err := tpls.T("api", "new").Execute(fObj, tpls.Params().
		With("package", packageName).
		With("s_obj", fname).
		With("obj", utils.SnakeToPrivateCamel(fname)).
		With("Obj", utils.SnakeToCamel(fname)),
	)
	logx.AssertError(err)
	fragment[len(fragment)-1] = fname
	logx.AssertError(addRouter(false, fragment))

	return err
}

func addRouter(isDir bool, fragments []string) error {
	if len(fragments) <= 0 {
		return nil
	}
	name := fragments[len(fragments)-1]
	fragments[len(fragments)-1] = "init.go"
	initPath := utils.PathJoin(append([]string{*cmds.DirRoot, *cmds.DirApi}, fragments...)...)
	if !utils.FileExists(initPath) {
		logx.AssertError(newRouterInInit(fragments...))
	}
	logx.AssertError(appendRouterInInit(isDir, name, fragments...))
	return addRouter(true, fragments[:len(fragments)-1])
}

// 在 init.go 中添加路由 xxx.Use(r) or useXxx(r)
// isDir: 是否是目录
// fcName: 资源名 snake_case
// fragments: 文件路径片段，相对于api dir
func appendRouterInInit(isDir bool, fcName string, fragments ...string) error {
	fAbsPath := utils.PathJoin(append([]string{*cmds.DirRoot, *cmds.DirApi}, fragments...)...)
	packageName := `"` + utils.PathJoin(*cmds.RepoName, *cmds.DirApi, strings.Join(fragments[:len(fragments)-1], "/"), fcName) + `"`

	fAst := logx.AssertFuncErr(tpls.NewAst(fAbsPath))

	// 标记是否找到并添加了 xxx.Use(r)
	shouldAddUse := false
	importAdded := false

	//  查找 `Use` 函数
	fAst.Inspect(func(n ast.Node) bool {
		// 查找函数声明
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "Use" {
				if isDir {
					//  检查 `Use` 函数内部是否有 `xxx.Use(r)`
					hasXxxUse := false
					for _, stmt := range funcDecl.Body.List {
						if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
							if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
								if selector, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
									if xIdent, ok := selector.X.(*ast.Ident); ok && xIdent.Name == fcName && selector.Sel.Name == "Use" {
										hasXxxUse = true
										break
									}
								}
							}
						}
					}

					// 如果没有 `xxx.Use(r)`，则添加
					if !hasXxxUse {
						newCall := &ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent(fcName),
									Sel: ast.NewIdent("Use"),
								},
								Args: []ast.Expr{ast.NewIdent("r")},
							},
						}
						funcDecl.Body.List = append(funcDecl.Body.List, newCall)
						shouldAddUse = true
					}
				} else {
					//  检查 `Use` 函数内部是否有 `useXxx(r)`
					hasUseRouter := false
					useXxx := fmt.Sprintf("use%s", utils.SnakeToCamel(fcName))
					for _, stmt := range funcDecl.Body.List {
						if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
							if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
								if funIdent, ok := callExpr.Fun.(*ast.Ident); ok && funIdent.Name == useXxx {
									hasUseRouter = true
									break
								}
							}
						}
					}

					//  如果没有 `useXxx(r)`，则添加
					if !hasUseRouter {
						arg := "r"
						if fcName != fAst.Name.Name {
							arg = fmt.Sprintf(`r.SubRouter(":%s_id/%s")`, fAst.Name.Name, fcName)
						}
						newCall := &ast.ExprStmt{
							X: &ast.CallExpr{
								Fun:  ast.NewIdent(useXxx),
								Args: []ast.Expr{ast.NewIdent(arg)},
							},
						}
						funcDecl.Body.List = append(funcDecl.Body.List, newCall)
						shouldAddUse = true
					}
				}
			}
		}

		// 检查和添加import
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT && isDir {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					if importSpec.Path.Value == packageName {
						importAdded = true
						break
					}
				}
			}
			if !importAdded {
				// 添加 import "xxx"
				newImport := &ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: packageName,
					},
				}
				genDecl.Specs = append(genDecl.Specs, newImport)
				importAdded = true
			}
		}

		return true
	})

	// 如果没有找到现有的 import block，则需要创建一个新的
	if !importAdded && isDir {
		newImport := &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: packageName,
					},
				},
			},
		}
		fAst.Decls = append([]ast.Decl{newImport}, fAst.Decls...)
	}

	// 覆写文件
	if shouldAddUse {
		return fAst.Dump(fAbsPath)
	}

	return nil
}

func newRouterInInit(fragments ...string) error {
	fObj := tpls.OpenFile(append([]string{*cmds.DirApi}, fragments...)...)
	defer fObj.Close()
	packageName := *cmds.DirApi
	if len(fragments) > 1 {
		packageName = fragments[len(fragments)-2]
	}
	return tpls.T("api", "init").Execute(fObj, tpls.Params().With("package", packageName))
}
