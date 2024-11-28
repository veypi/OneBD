//
// struct.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-26 17:08
// Distributed under terms of the GPL license.
//

package crud

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

func Parse(objs ...any) StructList {
	structs := []*StructInfo{}
	for _, obj := range objs {
		s := &StructInfo{}
		s.parse(obj)
		structs = append(structs, s)
	}
	return structs
}

type structHandler struct {
	ObjName string
	Action  string
	Method  string
	Suffix  string
	Fields  []FieldInfo
}

func (t structHandler) String() string {
	return fmt.Sprintf("%s%s", t.ObjName, t.Action)
}

type StructList []*StructInfo

func (s StructList) RegistRouter(r rest.Router) StructList {
	for _, s := range s {
		s.RegistRouter(r)
	}
	return s
}

type StructInfo struct {
	top      *StructInfo
	registed bool
	v        reflect.Value
	t        reflect.Type
	// CamelName
	Name       string
	TableName  string
	Fields     []FieldInfo
	SubStructs StructList
	handlers   []*structHandler
}

func (s *StructInfo) HasMany(objs ...any) StructList {
	structs := []*StructInfo{}
	for _, obj := range objs {
		s := &StructInfo{
			top: s,
		}
		s.parse(obj)
		structs = append(structs, s)
	}
	// crud(r, structs...)
	s.SubStructs = append(s.SubStructs, structs...)
	return structs
}

func (s *StructInfo) RegistRouter(r rest.Router) {
	if s.top != nil {
		s.top.RegistRouter(r)
		return
	}
	s.register(r)
}

func (s *StructInfo) register(r rest.Router) {
	if !s.registed {
		s.registed = true
		crud(r, s)
	}
	snakeName := utils.CamelToSnake(s.Name)
	subr := r.SubRouter(fmt.Sprintf("/%s/:%s_id", snakeName, snakeName))
	for _, sub := range s.SubStructs {
		sub.register(subr)
	}
}

func (s *StructInfo) String() string {
	str := fmt.Sprintf("struct %s {\n", s.Name)
	for _, f := range s.Fields {
		str += fmt.Sprintf("\t%s\n", f.String())
	}
	str += "}\n"
	return str
}

func (s *StructInfo) Handlers() []*structHandler {
	return s.handlers
}

func (s *StructInfo) GetHandler(name string) *structHandler {
	for _, h := range s.handlers {
		if h.Action == name {
			return h
		}
	}
	return nil
}

func (s *StructInfo) GetField(name string) *FieldInfo {
	for _, f := range s.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func (s *StructInfo) parse(obj any) {
	s.v = reflect.ValueOf(obj)
	s.t = s.v.Type()
	if s.t.Kind() == reflect.Ptr {
		s.t = s.t.Elem()
	}
	s.Name = s.t.Name()
	if method := s.v.MethodByName("TableName"); method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			s.TableName = results[0].String()
		}
	}
	if s.TableName == "" {
		s.TableName = utils.CamelToSnake(s.Name)
		if regexp.MustCompile(`(s|z|x|sh|zh)$`).MatchString(s.TableName) {
			s.TableName = s.TableName + "es"
		} else if regexp.MustCompile(`\d$`).MatchString(s.TableName) {
		} else if regexp.MustCompile(`y$`).MatchString(s.TableName) {
			s.TableName = s.TableName[:len(s.TableName)-1] + "ies"
		} else {
			s.TableName = s.TableName + "s"
		}
	}
	s.parseFields(s.t, utils.CamelToSnake(s.Name))
	s.parseHandlers()
}

func (s *StructInfo) parseHandlers() {
	s.handlers = make([]*structHandler, 0, 6)
	for _, f := range s.Fields {
		for _, m := range f.Methods {
			h := s.GetHandler(m.Name)
			if h == nil {
				h = &structHandler{
					ObjName: s.Name,
					Action:  m.Name,
					Suffix:  m.Suffix,
					Method:  m.Method,
				}
				s.handlers = append(s.handlers, h)
			} else {
				if h.Method == "" && m.Method != "" {
					h.Method = m.Method
				}
				if h.Suffix == "" && m.Suffix != "" {
					h.Suffix = m.Suffix
				}
			}
			h.Fields = append(h.Fields, f)
		}
	}
	records := make(map[string]string)
	for _, h := range s.handlers {
		uniurl := h.Method + "_" + h.Suffix
		if tmp, ok := records[uniurl]; ok {
			logv.Warn().Msgf("duplicate %s handler: %s and %s: /%s", h.Method, h, tmp, h.Suffix)
		} else {
			records[uniurl] = h.String()
		}
	}
}
func (s *StructInfo) parseFields(t reflect.Type, obj_name string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// n := t.Field(i).Name

		// 如果字段是结构体类型，则递归调用
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			s.parseFields(f.Type, obj_name)
		} else {
			key := utils.CamelToSnake(f.Name)
			if jsonTag := f.Tag.Get("json"); jsonTag == "-" {
				// ignore this field
				continue
			} else if jsonTag != "" {
				key = jsonTag
			}
			// default methodstag: post,
			methodsTag := f.Tag.Get("methods")
			parseTag := f.Tag.Get("parse")
			if parseTag == "-" {
				// ignore this field
				continue
			}
			if key == "id" {
				if methodsTag == "" {
					// 否则，添加字段信息
					methodsTag = "get,patch,delete"
				}
				if parseTag == "" {
					parseTag = fmt.Sprintf("path@%s_id", obj_name)
				}
			} else {
				if methodsTag == "" {
					if f.Type.Kind() == reflect.Ptr {
						methodsTag = "*list,*post,*patch,*put"
					} else {
						methodsTag = "*list,post,*patch,put"
					}
				}
				if parseTag == "" {
					parseTag = "json"
				}
			}
			resF := FieldInfo{
				Name:    f.Name,
				Type:    f.Type.String(),
				Tag:     string(f.Tag),
				Key:     key,
				Methods: nil,
			}
			resF.ParseMethods(methodsTag, obj_name)
			resF.ParseParse(parseTag)
			s.Fields = append(s.Fields, resF)
		}
	}
}

// 定义一个结构体来存储字段信息
type FieldInfo struct {
	Name     string
	Key      string
	Type     string
	Tag      string
	Methods  []FieldMethod
	Src      string
	SrcAlias string
}

type FieldMethod struct {
	// CamelName
	Name    string
	HasStar bool
	Method  string
	Suffix  string
}

func (m *FieldMethod) String() string {
	txt := m.Name
	if defaultMethods[txt] == nil {
		txt += "@" + m.Method
		if m.Suffix != "" {
			txt += "@/" + m.Suffix
		}
	}
	if m.HasStar {
		txt = "*" + txt
	}
	return txt
}

func (f *FieldInfo) String() string {
	m := ""
	for _, method := range f.Methods {
		m += "," + method.String()
	}
	src := f.Src
	if f.SrcAlias != "" {
		src += "@" + f.SrcAlias
	}
	return fmt.Sprintf("%s:  %s  `json:\"%s\" method:\"%s\" parse:\"%s\"`", f.Name, f.Type, f.Key, m[1:], src)
}

// *Name@Method@/urlsuffix
var methodRex = regexp.MustCompile(`(\*?)([^@/,]+)(@([^@/,]+))?(@/([^@,]+))?`)

var defaultMethods = map[string]*FieldMethod{
	"List":   {Name: "List", Method: http.MethodGet, Suffix: ""},
	"Get":    {Name: "Get", Method: http.MethodGet, Suffix: ":#id"},
	"Post":   {Name: "Post", Method: http.MethodPost, Suffix: ""},
	"Patch":  {Name: "Patch", Method: http.MethodPatch, Suffix: ":#id"},
	"Put":    {Name: "Put", Method: http.MethodPut, Suffix: ""},
	"Delete": {Name: "Delete", Method: http.MethodDelete, Suffix: ":#id"},
}

func (f *FieldInfo) ParseMethods(tag string, obj_name string) {
	// f.Methods = make(map[]FieldMethod)
	tag = strings.Replace(tag, " ", "", -1)
	matches := methodRex.FindAllStringSubmatch(tag, -1)
	for _, match := range matches {
		hasStar := len(match[1]) > 0
		name := utils.SnakeToCamel(match[2])
		method := http.MethodGet
		suffix := ""
		m := defaultMethods[name]
		if m != nil {
			method = m.Method
			name = m.Name
			suffix = m.Suffix
		}
		if len(match[4]) > 0 {
			method = strings.ToUpper(match[4])
		}
		if len(match[6]) > 0 {
			suffix = match[6]
		}
		suffix = strings.ReplaceAll(suffix, "#id", obj_name+"_id")
		f.Methods = append(f.Methods, FieldMethod{
			Name:    name,
			HasStar: hasStar,
			Method:  method,
			Suffix:  suffix,
		})
	}
}

func (f *FieldInfo) ParseParse(tag string) {
	tag = strings.Replace(tag, " ", "", -1)
	f.Src = tag
	tags := strings.Split(tag, "@")
	f.Src = utils.CamelToSnake(tags[0])
	if !utils.InList(f.Src, []string{"path", "query", "header", "form", "json"}) {
		logv.Warn().Msgf("parse tag: %s not support, use default json", f.Src)
		f.Src = "json"
	}
	if len(tags) > 1 {
		f.SrcAlias = tags[1]
	}
}
