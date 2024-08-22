//
// model.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 17:12
// Distributed under terms of the MIT license.
//

package model

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"
	"sort"
	"strings"

	"github.com/veypi/OneBD/clis/cmds"
	"github.com/veypi/OneBD/clis/tpls"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

var (
	newModel = cmds.Model.SubCommand("new", "generate new model")
	nameObj  = newModel.String("n", "user", "target model name")
)

var (
	nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
)

func init() {
	newModel.Command = new_model
	cmds.Model.Command = gen_model
}

func new_model() error {
	name := []byte(*nameObj)
	name[0] = bytes.ToLower(name[:1])[0]
	fname := string(name)
	name[0] = bytes.ToUpper(name[:1])[0]
	objName := string(name)

	fObj := tpls.OpenFile("models", fname+".go")
	defer fObj.Close()
	return tpls.T("models/obj").Execute(fObj, tpls.Params().With("obj", objName))
}

var methodReg = regexp.MustCompile(`methods:"([^"]+)"`)
var allowedMethods = []string{
	"Get", "List", "Post", "Put",
	"Patch", "Delete"}

func gen_model() error {

	fset := token.NewFileSet()
	filename := utils.PathJoin(*cmds.DirPath, "models", "user.go")
	node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return err
	}
	// 遍历AST，找到所有结构体
	newStructs := make(map[string]*ast.StructType)

	// 用于存储需要的import路径
	imports := map[string]bool{}
	for _, decl := range node.Decls {
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
				res := methodReg.FindStringSubmatch(field.Tag.Value)
				if len(res) == 0 {
					continue
				}
				for _, tag := range res[1:] {
					field.Tag.Value = strings.ReplaceAll(field.Tag.Value, res[0], "")
					methods := strings.Split(strings.ReplaceAll(tag, " ", ""), ",")
					for _, m := range methods {
						method := utils.ToTitle(m)
						if !utils.InList(method, allowedMethods) {
							logx.Warn().Msgf("method %s not allowed", method)
							continue
						}
						if newStructs[typeSpec.Name.Name+method] == nil {
							newStructs[typeSpec.Name.Name+method] = &ast.StructType{Fields: &ast.FieldList{}}
						}
						newStructs[typeSpec.Name.Name+method].Fields.List = append(newStructs[typeSpec.Name.Name+method].Fields.List, field)
						// 检查字段类型并添加相应的import路径
						checkAndAddImport(field.Type, imports)
					}
				}
			}
		}
	}
	importsList := []string{}
	for a := range imports {
		importsList = append(importsList, a)
	}
	structNames := make([]string, 0, len(newStructs))
	for k := range newStructs {
		structNames = append(structNames, k)
	}
	sort.Strings(structNames)

	var decls []ast.Decl
	for _, t := range structNames {
		decls = append(decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{&ast.TypeSpec{
				Name: ast.NewIdent(t),
				Type: newStructs[t],
			}},
		})
	}
	f := &bytes.Buffer{}
	if err := printer.Fprint(f, fset, decls); err != nil {
		return err
	}

	fObj := tpls.OpenFile("models", "user_gen.go")
	defer fObj.Close()
	return tpls.T("models/gen").Execute(fObj, tpls.Params().With("structs", f.String()).With("imports", importsList))
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
