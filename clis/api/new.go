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
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

func new_model() error {
	if !nameRegex.MatchString(*nameObj) {
		panic("invalid name")
	}
	fragment := strings.Split(*nameObj, ".")
	name := fragment[len(fragment)-1]
	fragment = append([]string{*cmds.DirApi}, fragment...)
	fname := utils.CamelToSnake(name)
	fragment[len(fragment)-1] = fname + ".go"

	fObj := tpls.OpenFile(fragment...)
	defer fObj.Close()
	err := tpls.T("api", "new").Execute(fObj, tpls.Params().With("package", fragment[len(fragment)-2]).With("obj", utils.SnakeToPrivateCamel(fname)).With("Obj", utils.SnakeToCamel(fname)))
	logx.AssertError(err)
	fragment[len(fragment)-1] = fname
	logx.AssertError(addRouter(false, fragment))

	return err
}

func addRouter(isDir bool, fragments []string) error {
	if len(fragments) <= 1 {
		return nil
	}
	name := fragments[len(fragments)-1]
	fragments[len(fragments)-1] = "init.go"
	initPath := utils.PathJoin(append([]string{*cmds.DirRoot}, fragments...)...)
	if !utils.FileExists(initPath) {
		logx.AssertError(newRouterInInit(fragments...))
	}
	logx.AssertError(appendRouterInInit(isDir, name, fragments...))
	return addRouter(true, fragments[:len(fragments)-1])
}

func appendRouterInInit(isDir bool, fcName string, fragments ...string) error {
	fAbsPath := utils.PathJoin(append([]string{*cmds.DirRoot}, fragments...)...)
	packageName := fmt.Sprintf(`"%s/%s/%s"`, *cmds.RepoName, strings.Join(fragments[:len(fragments)-1], "/"), fcName)
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fAbsPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("Error parsing file: %v", err)
	}

	// 标记是否找到并添加了 xxx.Use(r)
	shouldAddUse := false
	importAdded := false

	//  查找 `Use` 函数
	ast.Inspect(node, func(n ast.Node) bool {
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
						newCall := &ast.ExprStmt{
							X: &ast.CallExpr{
								Fun:  ast.NewIdent(useXxx),
								Args: []ast.Expr{ast.NewIdent("r")},
							},
						}
						funcDecl.Body.List = append(funcDecl.Body.List, newCall)
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
		fmt.Println("Creating new import block for xxx")
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
		node.Decls = append([]ast.Decl{newImport}, node.Decls...)
	}

	// 5. 覆写文件
	if shouldAddUse || !importAdded {
		outputFilename := fAbsPath
		f, err := os.Create(outputFilename)
		if err != nil {
			return fmt.Errorf("Error creating file: %v", err)
		}
		defer f.Close()

		if err := printer.Fprint(f, fset, node); err != nil {
			return fmt.Errorf("Error writing to file: %v", err)
		}
	}

	return nil
}

func newRouterInInit(fragments ...string) error {
	fObj := tpls.OpenFile(fragments...)
	defer fObj.Close()
	return tpls.T("api", "init").Execute(fObj, tpls.Params().With("package", fragments[len(fragments)-2]))
}
