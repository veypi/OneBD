//
// ast.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-09-02 13:47
// Distributed under terms of the MIT license.
//

package tpls

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/veypi/utils"
)

func Ast2Str(decls any) (string, error) {
	f := &bytes.Buffer{}
	err := printer.Fprint(f, token.NewFileSet(), decls)
	return f.String(), err
}

func NewAst(fPath string) (*Ast, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &Ast{node, fset}, nil
}

func NewEmptyAst(packageName string) *Ast {

	packageDecl := &ast.File{
		Name:  ast.NewIdent(packageName),
		Decls: []ast.Decl{},
	}

	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{},
	}

	packageDecl.Decls = append(packageDecl.Decls, importDecl)
	return &Ast{packageDecl, token.NewFileSet()}

}

func NewFileOrEmptyAst(fPath string, packageName string) (*Ast, error) {
	if utils.FileExists(fPath) {
		return NewAst(fPath)
	}
	return NewEmptyAst(packageName), nil
}

type Ast struct {
	*ast.File
	fset *token.FileSet
}

func (a *Ast) Inspect(fn func(ast.Node) bool) {
	ast.Inspect(a.File, fn)
}

func (a *Ast) Dump(fPath string) error {
	f, err := utils.MkFile(fPath)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer f.Close()

	if err := printer.Fprint(f, a.fset, a.File); err != nil {
		return fmt.Errorf("Error writing to file: %v", err)
	}
	return nil
}

func (a *Ast) AddStructWithFields(name string, fields ...*ast.Field) error {
	found := false
	a.Inspect(func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if styp, ok := typeSpec.Type.(*ast.StructType); ok {
				if typeSpec.Name.Name == name {
					for _, f := range fields {
						field_found := false
						for _, tf := range styp.Fields.List {
							if len(tf.Names) > 0 && len(f.Names) > 0 && tf.Names[0].Name == f.Names[0].Name {
								tf.Type = f.Type
								tf.Tag = f.Tag
								field_found = true
							}
						}
						if !field_found {
							styp.Fields.List = append(styp.Fields.List, f)
						}
					}
					found = true
					return false
				}
			}
		}
		return true
	})
	if !found {
		return a.AddStruct(name, &ast.StructType{
			Fields: &ast.FieldList{
				List: fields,
			},
		})
	}
	return nil
}

func (a *Ast) AddStruct(name string, styp *ast.StructType) error {
	found := false
	a.Inspect(func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				if typeSpec.Name.Name == name {
					found = true
					return false
				}
			}
		}
		return true
	})
	if !found {
		newStruct := &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(name),
					Type: styp,
				},
			},
			Doc: &ast.CommentGroup{
				List: []*ast.Comment{
					{
						Slash: token.NoPos,
						Text:  "//",
					},
				},
			},
		}
		a.Decls = append(a.Decls, newStruct)
	}
	return nil
}

func (a *Ast) AddImport(repo string) error {
	repo = fmt.Sprintf(`"%s"`, repo)
	importAdded := false
	a.Inspect(func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					if importSpec.Path.Value == repo {
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
						Value: repo,
					},
				}
				genDecl.Specs = append(genDecl.Specs, newImport)
				importAdded = true
			}
			return false
		}
		return true
	})
	if !importAdded {
		newImport := &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: repo,
					},
				},
			},
		}
		a.Decls = append([]ast.Decl{newImport}, a.Decls...)
	}
	return nil
}

func NewAstField(name, typ, tag string) *ast.Field {
	r := &ast.Field{
		Type: ast.NewIdent(typ),
	}
	if name != "" {
		r.Names = []*ast.Ident{ast.NewIdent(name)}
	}
	if tag != "" {
		r.Tag = &ast.BasicLit{Kind: token.STRING, Value: tag}
	}
	return r
}
