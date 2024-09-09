//
// gen_fc.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-09-05 17:03
// Distributed under terms of the MIT license.
//

package api

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/veypi/utils"
)

func create_api_func(obj, method string) *ast.FuncDecl {

	// 构建变量声明：opts := M.ObjMethod{}
	optsDecl := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("opts")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CompositeLit{
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("M"),
					Sel: ast.NewIdent(obj + method),
				},
			},
		},
	}
	// 构建函数调用：err := opts.Parse(x)
	parseCall := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun:  ast.NewIdent("opts.Parse"),
				Args: []ast.Expr{ast.NewIdent("x")},
			},
		},
	}
	// 构建 if 语句：if err != nil { return nil, err }
	errifStmt := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("err"),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						ast.NewIdent("nil"),
						ast.NewIdent("err"),
					},
				},
			},
		},
	}

	// 构建变量声明：data := M.Obj{}
	dataDecl := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("data")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CompositeLit{
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("M"),
					Sel: ast.NewIdent(obj),
				},
			},
		},
	}

	// 构建 return 语句：return data, nil
	returnStmt := &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("data"),
			ast.NewIdent("nil"),
		},
	}

	body := &ast.BlockStmt{
		List: []ast.Stmt{
			optsDecl,
			parseCall,
			errifStmt,
			dataDecl,
			&ast.ExprStmt{X: ast.NewIdent("// edit here")},
			&ast.ExprStmt{X: ast.NewIdent("")},
			returnStmt,
		},
	}
	// 生成方法体 func objMethod(x *rest.X) error
	return &ast.FuncDecl{
		Name: ast.NewIdent(utils.ToLowerFirst(obj + method)),
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
					{Type: ast.NewIdent("any")},
					{Type: ast.NewIdent("error")},
				},
			},
		},
		Body: body,
	}
}

func create_use_func(name string) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("r")},
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "rest"},
							Sel: &ast.Ident{Name: "Router"},
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{},
	}
}

func create_use_stmt(name, method string) ast.Stmt {
	u := `"/"`
	if method == "Get" || method == "Patch" || method == "Delete" {
		u = fmt.Sprintf(`"/:%s_id"`, utils.CamelToSnake(name))
	}
	if method == "List" {
		method = "Get"
	}
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("r"),
				Sel: ast.NewIdent(method),
			},
			Args: []ast.Expr{
				ast.NewIdent(u),
				ast.NewIdent(utils.ToLowerFirst(name) + method),
			},
		},
	}
}
