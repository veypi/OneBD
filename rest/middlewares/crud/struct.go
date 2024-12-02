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

func New() *StructGraph {
	return &StructGraph{
		nodes: make(map[string]*StructInfo),
		edges: make(map[string]map[string]*StructRelation),
	}
}

type StructGraph struct {
	edges map[string]map[string]*StructRelation
	// 结构体同名视为同一结构体
	nodes map[string]*StructInfo
}

func (s *StructGraph) RegistRouter(r rest.Router, fn func(rest.Router, *StructInfo)) {
	s.calculateDegree()
	// 从入度为0的节点开始深度遍历注册
	for _, n := range s.nodes {
		if n.degree[0] == 0 {
			s.registRouterFrom(n, r, fn)
		}
	}
}
func (s *StructGraph) registRouterFrom(from *StructInfo, r rest.Router, fn func(rest.Router, *StructInfo)) {
	fn(r, from)
	snakeName := utils.CamelToSnake(from.Name)
	subr := r.SubRouter(fmt.Sprintf("/%s/:%s_id", snakeName, snakeName))
	for to, sr := range s.edges[from.Name] {
		if sr.Typ == SROne2Many {
			s.registRouterFrom(s.Get(to), subr, fn)
		}
	}
	// RegistRouter(subr, fn, s.SubStructs...)
}

func (s *StructGraph) Append(objs ...any) {
	// *s = append(*s, Parse(objs...)...)
	for _, obj := range objs {
		s.Add(obj)
	}
}

func (s *StructGraph) Add(obj any) *StructInfo {
	// *s = append(*s, Parse(obj)...)
	sObj := s.Get(obj)
	if sObj == nil {
		sObj = &StructInfo{
			root: s,
		}
		sObj.parse(obj)
		s.nodes[sObj.Name] = sObj
		logv.WithNoCaller.Debug().Msgf("regist obj %s", sObj.Name)
	}
	return sObj
}

func (s *StructGraph) Get(t any) *StructInfo {
	if tt, ok := t.(string); ok {
		return s.nodes[tt]
	}
	var tt reflect.Type
	if temp, ok := t.(reflect.Type); ok {
		tt = temp
	} else {
		tt = reflect.TypeOf(t)
	}
	if tt.Kind() == reflect.Ptr {
		tt = tt.Elem()
	}
	return s.nodes[tt.Name()]
}

func (s *StructGraph) calculateDegree() {
	for _, n := range s.nodes {
		n.degree = [2]int{0, 0}
	}
	for out, outMap := range s.edges {
		for in := range outMap {
			s.nodes[out].degree[1]++
			s.nodes[in].degree[0]++
		}
	}
}

func (s *StructGraph) One2Many(from any, to ...any) {
	f := s.Add(from)
	var t *StructInfo
	for _, tObj := range to {
		t = s.Add(tObj)
		if s.edges[f.Name] == nil {
			s.edges[f.Name] = make(map[string]*StructRelation)
		}
		s.edges[f.Name][t.Name] = &StructRelation{
			Typ: SROne2Many,
		}
		f = t
	}
}

func (s *StructGraph) Many2Many(middle any, Others ...any) {
	m := s.Add(middle)
	rs := &StructRelation{
		Typ:          SRMany2Many,
		Associations: make([]reflect.Type, 0, len(Others)),
	}
	for _, o := range Others {
		t := s.Add(o)
		if s.edges[t.Name] == nil {
			s.edges[t.Name] = make(map[string]*StructRelation)
		}
		s.edges[t.Name][m.Name] = rs
		rs.Associations = append(rs.Associations, t.t)
	}
}

var (
	SRMany2Many = "many2many"
	SROne2Many  = "one2many"
)

// Out --> In
type StructRelation struct {
	Typ          string
	Associations []reflect.Type
}

func (s *StructRelation) RegistRouter(r rest.Router, fn func(rest.Router, *StructInfo)) {
}

// 兼顾json 能从json导出导入结构体描述，字段描述和结构间关系
type StructInfo struct {
	v reflect.Value
	t reflect.Type
	// [in degree, out degree]
	degree [2]int
	root   *StructGraph
	// CamelName
	Name      string
	TableName string
	Fields    []StructField
	Relations []StructRelation
	handlers  []*StructHandler
}

func (s *StructInfo) HasMany(objs ...any) {
	for _, obj := range objs {
		s.root.One2Many(s.t, obj)
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

func (s *StructInfo) Handlers() []*StructHandler {
	return s.handlers
}

func (s *StructInfo) GetHandler(name string) *StructHandler {
	for _, h := range s.handlers {
		if h.Action == name {
			return h
		}
	}
	return nil
}

func (s *StructInfo) GetField(name string) *StructField {
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
	if s.t.Kind() != reflect.Struct {
		logv.Fatal().Msgf("obj must be a struct: %T: %v", obj, obj)
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
	s.handlers = make([]*StructHandler, 0, 6)
	for _, f := range s.Fields {
		for _, m := range f.Methods {
			h := s.GetHandler(m.Action)
			if h == nil {
				h = &StructHandler{
					ObjName: s.Name,
					Action:  m.Action,
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
			h.Fields = append(h.Fields, HandlerField{
				Name:    f.Name,
				Alias:   m.Alias,
				Key:     f.Key,
				Type:    f.Type,
				HasStar: m.HasStar,
				Src:     m.Src,
			})
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
			if methodsTag == "-" {
				// ignore this field
				continue
			}
			if key == "id" {
				// id 默认忽视，不存储在argParser中，由crud自动在get,patch,delete做sql条件查询处理
				if methodsTag == "" {
					methodsTag = "get@path,delete@path"
				}
			} else {
				if methodsTag == "" {
					// 其余自动默认在list 做条件查询，在post,patch，put做字段更新
					// post, put做数据创建时，会做*参数检查是否存在
					if f.Type.Kind() == reflect.Ptr {
						methodsTag = "*list,*post,*patch,*put"
					} else {
						methodsTag = "*list,post,*patch,put"
					}
				}
			}
			resF := StructField{
				Name:    f.Name,
				Type:    f.Type.String(),
				Tag:     string(f.Tag),
				Key:     key,
				Methods: nil,
			}
			resF.ParseParse(parseTag)
			resF.ParseMethods(methodsTag, obj_name)
			s.Fields = append(s.Fields, resF)
		}
	}
}

type StructHandler struct {
	ObjName string
	Action  string
	Method  string
	Suffix  string
	Fields  []HandlerField
}

// 存储不同请求方式里参数包含的字段
type HandlerField struct {
	// 原始参数名 CamelName
	Name string
	// 请求参数名,为空时为key
	Alias string
	// 存储字段名
	Key     string
	Type    string
	HasStar bool
	Src     string
}

func (t StructHandler) String() string {
	return fmt.Sprintf("%s%s", t.ObjName, t.Action)
}

// 定义一个结构体来存储字段信息
type StructField struct {
	// 原始名 CamelName
	Name string
	// 通信名， 默认为snake_name, 可以通过json tag自定义
	Key     string
	Type    string
	Tag     string
	Methods []FieldMethod
	// path header query form json
	Src      string
	SrcAlias string
}

// *Action@Get@/urlsuffix@json@:argname
type FieldMethod struct {
	// 默认支持 Get,List,Post,Patch,Put,Delete
	Action string
	Suffix string
	// :argname定义 默认为空，设置时覆盖field.key
	Alias   string
	HasStar bool
	Method  string
	// json path query form header
	Src string
}

func (m *FieldMethod) String() string {
	txt := m.Action
	if tpl := defaultActions[txt]; tpl != nil {
		if m.Suffix != tpl.Suffix {
			txt += "@/" + m.Suffix
		}
		if m.Src != m.Src {
			txt += "@" + m.Src
		}
	} else {
		if m.Method != "GET" {
			txt += "@" + m.Method
		}
		if m.Suffix != "" {
			txt += "@/" + m.Suffix
		}
		if m.Src != "json" {
			txt += "@" + m.Src
		}
	}
	if m.HasStar {
		txt = "*" + txt
	}
	if m.Alias != "" {
		txt += "@:" + m.Alias
	}
	return txt
}

func (f *StructField) String() string {
	m := ""
	for _, method := range f.Methods {
		m += "," + method.String()
	}
	return fmt.Sprintf("%s:  %s  `json:\"%s\" method:\"%s\"", f.Name, f.Type, f.Key, m[1:])
}

var defaultActions = map[string]*FieldMethod{
	"List":   {Action: "List", Method: http.MethodGet, Suffix: "", Src: "query"},
	"Get":    {Action: "Get", Method: http.MethodGet, Suffix: ":#id", Src: "query"},
	"Post":   {Action: "Post", Method: http.MethodPost, Suffix: "", Src: "json"},
	"Patch":  {Action: "Patch", Method: http.MethodPatch, Suffix: ":#id", Src: "json"},
	"Put":    {Action: "Put", Method: http.MethodPut, Suffix: "", Src: "json"},
	"Delete": {Action: "Delete", Method: http.MethodDelete, Suffix: ":#id", Src: "json"},
}
var defaultSources = []string{"path", "query", "header", "form", "json"}
var defaultMethods = []string{"get", "post", "patch", "put", "delete"}

// *Action@Get@/urlsuffix@json@:argname
func (f *StructField) ParseMethods(tag string, obj_name string) {
	// f.Methods = make(map[]FieldMethod)
	tag = strings.Replace(tag, " ", "", -1)
	for _, tags := range strings.Split(tag, ",") {
		matches := strings.Split(tags, "@")
		if len(matches) == 0 || matches[0] == "" {
			continue
		}
		fm := FieldMethod{
			Method: "GET",
			Src:    "json",
		}
		if matches[0][0] == '*' {
			fm.HasStar = true
			fm.Action = utils.SnakeToCamel(matches[0][1:])
		} else {
			fm.Action = utils.SnakeToCamel(matches[0])
		}
		m := defaultActions[fm.Action]
		if m != nil {
			fm.Method = m.Method
			fm.Suffix = m.Suffix
			fm.Src = m.Src
		}
		if f.Src != "" {
			fm.Src = f.Src
		}
		if f.SrcAlias != "" {
			fm.Alias = f.SrcAlias
		}
		for _, subMatch := range matches[1:] {
			if subMatch == "" {
				continue
			} else if subMatch[0] == ':' {
				fm.Alias = subMatch[1:]
			} else if subMatch[0] == '/' {
				fm.Suffix = subMatch[1:]
			} else if utils.InList(subMatch, defaultSources) {
				fm.Src = subMatch
			} else if utils.InList(subMatch, defaultMethods) {
				fm.Method = strings.ToUpper(subMatch)
			} else {
				logv.Warn().Msgf("method tag: %s not support, %s", subMatch, tag)
			}
		}
		fm.Suffix = strings.ReplaceAll(fm.Suffix, "#id", obj_name+"_id")
		f.Methods = append(f.Methods, fm)
	}
}

// parseTag:  src@alias, 优先级低，为此参数默认来源
func (f *StructField) ParseParse(tag string) {
	if tag == "" || tag == "-" {
		return
	}
	tag = strings.Replace(tag, " ", "", -1)
	f.Src = tag
	tags := strings.Split(tag, "@")
	f.Src = utils.CamelToSnake(tags[0])
	if !utils.InList(f.Src, defaultSources) {
		logv.Warn().Msgf("parse tag: %s not support, use default json", f.Src)
		f.Src = "json"
	}
	if len(tags) > 1 {
		f.SrcAlias = tags[1]
	}
}
