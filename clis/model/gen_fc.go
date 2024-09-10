//
// gen_fc.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-09-04 18:56
// Distributed under terms of the MIT license.
//

package model

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

// 创建Parse方法函数体
/*
func (m *ModelObj) Parse(x *rest.X) error {
	contentType := x.Request.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
        err := x.Request.ParseForm()
		if err != nil {
			return err
		}
	} else {
        err := json.NewDecoder(x.Request.Body).Decode(m)
		if err != nil {
			return err
		}
	}
	return nil
}
*/
func getArgsTransFunc(f *ast.Field, arg ast.Expr, imports map[string]bool) ast.Expr {
	typ := f.Type
	if sTyp, ok := typ.(*ast.StarExpr); ok {
		typ = sTyp.X
	}
	switch t := typ.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return arg
		case "int":
			imports[fmt.Sprintf("%s/vtools", *cmds.RepoName)] = true
			return &ast.CallExpr{
				Fun:  ast.NewIdent("vtools.Str2Int"),
				Args: []ast.Expr{arg},
			}
		case "uint":
			imports[fmt.Sprintf("%s/vtools", *cmds.RepoName)] = true
			return &ast.CallExpr{
				Fun:  ast.NewIdent("vtools.Str2Uint"),
				Args: []ast.Expr{arg},
			}
		}
		return arg
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				imports[fmt.Sprintf("%s/vtools", *cmds.RepoName)] = true
				return &ast.CallExpr{
					Fun:  ast.NewIdent("vtools.Str2Time"),
					Args: []ast.Expr{arg},
				}
			}
		}
	default:
		logv.Warn().Msgf("getArgsTransFunc not support: %s %T", f.Names[0].Name, f.Type)
	}
	return arg
}
func createParseBody(name string, fields []*ast.Field, imports map[string]bool) *ast.FuncDecl {
	imports["fmt"] = true
	var paramstmts = make([]ast.Stmt, 0)
	var formStmts = make([]ast.Stmt, 0)
	hasQuery := false
	for _, f := range fields {
		if f.Tag == nil {
			continue
		}
		res := parseReg.FindStringSubmatch(f.Tag.Value)
		// form and json depend on the request content-type
		source := "form"
		if len(res) > 1 {
			// path query header
			source = res[1]
		}
		key := f.Names[0].Name
		if key == "ID" {
			if objTag := objReg.FindStringSubmatch(name); len(objTag) > 1 {
				key = objTag[1] + "ID"
			}
		}
		if len(res) > 2 && res[2] != "" {
			key = res[2]
		}
		key = "\"" + utils.CamelToSnake(key) + "\""
		if source == "form" {
			formStmts = append(formStmts, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("m." + f.Names[0].Name)},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					getArgsTransFunc(
						f,
						&ast.CallExpr{
							Fun: ast.NewIdent("x.Request.Form.Get"),
							Args: []ast.Expr{
								ast.NewIdent(key),
							},
						}, imports),
				},
			})
		} else if source == "path" {
			paramstmts = append(paramstmts, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("m." + f.Names[0].Name)},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					getArgsTransFunc(f,
						&ast.CallExpr{
							Fun: ast.NewIdent("x.Params.GetStr"),
							Args: []ast.Expr{
								ast.NewIdent(key),
							},
						}, imports),
				},
			})
		} else if source == "header" {
			paramstmts = append(paramstmts, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("m." + f.Names[0].Name)},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					getArgsTransFunc(f,
						&ast.CallExpr{
							Fun: ast.NewIdent("x.Request.Header.Get"),
							Args: []ast.Expr{
								ast.NewIdent(key),
							},
						}, imports),
				},
			})
		} else {
			// query
			if !hasQuery {
				hasQuery = true
				queryAssign := &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("queryMap")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("x.Request.URL.Query"),
						},
					},
				}
				paramstmts = append([]ast.Stmt{queryAssign}, paramstmts...)
			}
			paramstmts = append(paramstmts, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("m." + f.Names[0].Name)},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					getArgsTransFunc(f,
						&ast.CallExpr{
							Fun: ast.NewIdent("queryMap.Get"),
							Args: []ast.Expr{
								ast.NewIdent(key),
							},
						}, imports),
				},
			})

		}
	}

	// 构建表达式：contentType := x.Request.Header.Get("Content-Type")
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("contentType")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("x.Request.Header.Get"),
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "\"Content-Type\"",
					},
				},
			},
		},
	}

	// 构建 if 语句
	ifStmt := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("contentType"),
			Op: token.EQL,
			Y: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"application/x-www-form-urlencoded\"",
			},
		},
		Body: &ast.BlockStmt{
			List: append([]ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("x"),
									Sel: ast.NewIdent("Request"),
								},
								Sel: ast.NewIdent("ParseForm"),
							},
						},
					},
				},
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{ast.NewIdent(`fmt.Errorf("parse form error %v", err)`)},
							},
						},
					},
				},
			}, formStmts...),
		},
		Else: &ast.IfStmt{
			Cond: ast.NewIdent(`contentType == "application/json"`),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent("err")},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("json"),
											Sel: ast.NewIdent("NewDecoder"),
										},
										Args: []ast.Expr{
											&ast.SelectorExpr{
												X: &ast.SelectorExpr{
													X:   ast.NewIdent("x"),
													Sel: ast.NewIdent("Request"),
												},
												Sel: ast.NewIdent("Body"),
											},
										},
									},
									Sel: ast.NewIdent("Decode"),
								},
								Args: []ast.Expr{ast.NewIdent("m")},
							},
						},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X:  ast.NewIdent("err"),
							Op: token.NEQ,
							Y:  ast.NewIdent("nil"),
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{ast.NewIdent(`fmt.Errorf("parse json error %v", err)`)},
								},
							},
						},
					},
				},
			},
		},
	}
	body := &ast.BlockStmt{
		List: append([]ast.Stmt{
			// &ast.ExprStmt{X: ast.NewIdent("// Auto Generate")},
			assignStmt,
			ifStmt,
		}, paramstmts...),
	}
	body.List = append(body.List, &ast.ReturnStmt{
		Results: []ast.Expr{ast.NewIdent("nil")},
	})
	// 生成方法体 func (m *Name) Parse(x *rest.X) error
	return &ast.FuncDecl{
		Name: ast.NewIdent("Parse"),
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("m")},
					Type:  &ast.StarExpr{X: ast.NewIdent(name)},
				},
			},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("x")},
						Type: &ast.StarExpr{X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "rest"},
							Sel: &ast.Ident{Name: "X"},
						}},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: body,
	}
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

func printAst(src string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			ast.Print(fset, funcDecl.Body.List[0].End())
		}
		return true
	})
}
