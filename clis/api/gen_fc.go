//
// gen_fc.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-09-05 17:03
// Distributed under terms of the MIT license.
//

package api

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

func create_api_func(fAst *tpls.Ast, obj string, method string, objFields []*ast.Field) *ast.FuncDecl {

	// 构建变量声明：opts := &M.ObjMethod{}
	optsDecl := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("opts")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.UnaryExpr{
				Op: token.AND, // 取地址符号 &
				X: &ast.CompositeLit{
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("M"),
						Sel: ast.NewIdent(obj + method),
					},
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
				Fun:  ast.NewIdent("x.Parse"),
				Args: []ast.Expr{ast.NewIdent("opts")},
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

	// 构建变量声明：data := &M.Obj{}
	dataDecl := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("data")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.UnaryExpr{
				Op: token.AND, // 取地址符号 &
				X: &ast.CompositeLit{
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("M"),
						Sel: ast.NewIdent(obj),
					},
				},
			},
		},
	}
	if method == "List" {
		dataDecl = &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("data")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("make"),
					Args: []ast.Expr{
						ast.NewIdent(fmt.Sprintf("[]*M.%s, 0, 10", obj)),
					},
				},
			},
		}
	}

	// 构建 return 语句：return data, nil
	returnStmt := &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("data"),
			ast.NewIdent("err"),
		},
	}
	customs := &bytes.Buffer{}
	fields := make([]map[string]string, 0, 10)
	for _, f := range objFields {
		temp := map[string]string{}
		if len(f.Names) > 0 {
			temp["name"] = f.Names[0].Name
			temp["snake"] = utils.CamelToSnake(f.Names[0].Name)
		}
		if ft, ok := f.Type.(*ast.StarExpr); ok {
			temp["is_pointer"] = "true"
			temp["type"] = fmt.Sprintf("%s", ft.X)
		} else {
			temp["type"] = fmt.Sprintf("%s", f.Type)
		}
		if f.Tag != nil {
			temp["tag"] = f.Tag.Value
		}
		fields = append(fields, temp)
	}
	logv.AssertError(tpls.T("snaps", "rest", strings.ToLower(method)).Execute(customs, tpls.Params().With("fields", fields)))
	customStmts := []ast.Stmt{
		&ast.ExprStmt{X: ast.NewIdent("")},
		&ast.ExprStmt{X: ast.NewIdent("")},
	}
	if customs.Len() > 0 {
		for _, s := range strings.Split(customs.String(), "\n") {
			if strings.HasPrefix(s, "import ") {
				fAst.AddImport(strings.Split(s, " ")[1])
			} else if s != "" && s != "\n" && strings.Trim(s, " ") != "" {
				// logv.WithNoCaller.Warn().Msgf("|%s", s)
				customStmts = append(customStmts, &ast.ExprStmt{X: ast.NewIdent(s)})
			}
		}
	}
	customStmts = append(customStmts, &ast.ExprStmt{X: ast.NewIdent("")})

	body :=
		[]ast.Stmt{
			optsDecl,
			parseCall,
			errifStmt,
			dataDecl,
			// &ast.ExprStmt{X: ast.NewIdent("// custom code")},
		}
	body = append(body, customStmts...)
	body = append(body, returnStmt)
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
		Body: &ast.BlockStmt{
			List: body,
		},
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
	callMethod := method
	if method == "Get" || method == "Patch" || method == "Delete" || method == "Put" {
		u = fmt.Sprintf(`"/:%s_id"`, utils.CamelToSnake(name))
	}
	if method == "List" {
		callMethod = "Get"
	}
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("r"),
				Sel: ast.NewIdent(callMethod),
			},
			Args: []ast.Expr{
				ast.NewIdent(u),
				ast.NewIdent(utils.ToLowerFirst(name) + method),
			},
		},
	}
}
