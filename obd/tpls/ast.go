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
	"strings"

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
	GoFmt(fPath)
	return nil
}

// add or replace struct methods
func (a *Ast) AddStructMethods(name string, fcName string, fc *ast.FuncDecl, reWrite bool) {
	sIndex := -1
	for i, decl := range a.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name == name {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						sIndex = i + 1
					}
				}
			}
		}
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if recvT, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident); ok {
					if recvT.Name == name && funcDecl.Name.Name == fcName {
						// found and replace method
						if reWrite {
							funcDecl.Body = fc.Body
							funcDecl.Type = fc.Type
						}
						return
					}
				}
			}
		}
	}
	if sIndex == -1 {
		sIndex = len(a.Decls)
	}
	newDecls := append(a.Decls[:sIndex], append([]ast.Decl{fc}, a.Decls[sIndex:]...)...)
	a.Decls = newDecls
	return
}

func (a *Ast) AppendStmtInMethod(t ast.Stmt, fc *ast.FuncDecl, keywords ...string) {
	fcStr, _ := Ast2Str(fc)
	tStr, _ := Ast2Str(t)
	if len(keywords) > 0 {
		for _, k := range keywords {
			if strings.Contains(fcStr, k) {
				return
			}
		}
	} else if strings.Contains(fcStr, tStr) {
		return
	}
	fc.Body.List = append(fc.Body.List, t)
	return
}

// add or replace methods
func (a *Ast) AddMethod(fc *ast.FuncDecl, reWrite bool) bool {
	for _, decl := range a.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == fc.Name.Name {
				// found and replace method
				if funcDecl.Body == nil || len(funcDecl.Body.List) == 0 || reWrite {
					funcDecl.Doc = fc.Doc
					funcDecl.Body = fc.Body
					funcDecl.Type = fc.Type
					return true
				}
				return false
			}
		}
	}
	a.Decls = append(a.Decls, fc)
	return true
}

type SimpleStruct struct {
	Name   string
	Fields []*ast.Field
}

func (a *Ast) GetAllStructs() []*SimpleStruct {
	res := make([]*SimpleStruct, 0)
	a.Inspect(func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if styp, ok := typeSpec.Type.(*ast.StructType); ok {
				// res[typeSpec.Name.Name] = styp
				res = append(res, &SimpleStruct{Name: typeSpec.Name.Name, Fields: styp.Fields.List})
			}
		}
		return true
	})
	return res
}

func (a *Ast) GetMethod(fcName string) *ast.FuncDecl {
	for _, decl := range a.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == fcName {
				return funcDecl
			}
		}
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
								field_found = true
							}
							if len(tf.Names) == 0 && len(f.Names) == 0 {
								ffName := f.Type.(*ast.Ident).Name
								if strings.Contains(ffName, ".") {
									ffName = strings.Split(ffName, ".")[1]
								}
								if ttf, ok := tf.Type.(*ast.Ident); ok && ffName == ttf.Name {
									field_found = true
								}
								if ttf, ok := tf.Type.(*ast.SelectorExpr); ok && ffName == ttf.Sel.Name {
									field_found = true
								}
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

func (a *Ast) AddImport(repo string, tags ...string) error {
	repo = fmt.Sprintf(`"%s"`, repo)
	var tag *ast.Ident
	if len(tags) > 0 {
		tag = ast.NewIdent(tags[0])
	}

	importAdded := false
	a.Inspect(func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					if importSpec.Path.Value == repo {
						if tag != nil && importSpec.Name != nil {
							if tag.Name == importSpec.Name.Name {
								importAdded = true
								break
							}
						} else if tag == nil && importSpec.Name == nil {
							importAdded = true
							break
						}
					}
				}
			}
			if !importAdded {
				// 添加 import "xxx"
				newImport := &ast.ImportSpec{
					Name: tag,
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
					Name: tag,
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

func stmtEqual(a, b ast.Stmt) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
