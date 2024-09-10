//
// x.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-09 13:08
// Distributed under terms of the MIT license.
//

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

type X struct {
	http.ResponseWriter
	Request *http.Request
	Params  Params
	fcs     []any
	fid     int
}

// 从不同来源解析目标结构体一级字段
// tag标签 parse:"path/header/query/form/json" 可以追加为 path@alias_name
// tag标签 default:""
// 字段为指针类型时为可选参数,defalt标签不生效
// 字段为非指针类型时为必选参数，default标签生效，未设置该值且未发现参数时报参数缺失错误
func (x *X) Parse(obj any) error {
	v := reflect.ValueOf(obj).Elem()
	contentType := x.Request.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		err := x.Request.ParseForm()
		if err != nil {
			return fmt.Errorf("parse form error %v", err)
		}
	} else if contentType == "application/json" {
		err := json.NewDecoder(x.Request.Body).Decode(obj)
		if err != nil {
			return fmt.Errorf("parse json error %v", err)
		}
	}
	t := v.Type()
	var queryMap url.Values
	// 遍历结构体的字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		// 获取字段的tag
		method := field.Tag.Get("parse")
		key := utils.CamelToSnake(field.Name)
		if strings.Contains(method, "@") {
			tmps := strings.Split(method, "@")
			method = tmps[0]
			key = tmps[1]
		}

		fContent := ""
		fContentSet := false
		switch method {
		case "path":
			fContent, fContentSet = x.Params.Get(key)
		case "query":
			if queryMap == nil {
				queryMap = x.Request.URL.Query()
				logv.WithNoCaller.Warn().Msgf("%v", queryMap)
			}
			if tmps, ok := queryMap[key]; ok {
				fContentSet = true
				fContent = tmps[0]
			}
		case "header":
			if tmps, ok := x.Request.Header[key]; ok {
				fContentSet = true
				fContent = tmps[0]
			}
		case "form":
			if tmps, ok := x.Request.Form[key]; ok {
				fContentSet = true
				fContent = tmps[0]
			}
		case "json", "":
			continue
		default:
			return fmt.Errorf("unknown parse method %s", method)
		}
		ft := field.Type
		isPointer := false
		logv.WithNoCaller.Info().Msgf("|%v|%v|%v|%v|", key, method, ft.Kind(), fContent)
		if ft.Kind() == reflect.Ptr {
			if !fContentSet {
				// 指针类型没有参数则直接跳过
				continue
			}
			isPointer = true
			ft = ft.Elem()
			fieldValue.Set(reflect.New(ft))
			fieldValue = fieldValue.Elem()
		} else if !fContentSet {
			// 非指针类型没有参数根据默认值设置，没有则返回缺少参数
			defaultValue, ok := field.Tag.Lookup("default")
			if !ok {
				return fmt.Errorf("missing arg %s", key)
			}
			fContent = defaultValue
		}

		if fContent[0] == '"' {
			// 去掉字符串两边的引号
			fContent = fContent[1 : len(fContent)-1]
		}

		var invalidArg = fmt.Errorf("invalid arg %s: %s", key, fContent)
		switch ft.Kind() {
		case reflect.String:
			fieldValue.SetString(fContent)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(fContent, 10, 64)
			if err != nil && !isPointer {
				return invalidArg
			}
			fieldValue.SetInt(n)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, err := strconv.ParseUint(fContent, 10, 64)
			if err != nil && !isPointer {
				return invalidArg
			}
			fieldValue.SetUint(n)
		case reflect.Slice:
			err := json.Unmarshal([]byte(fContent), fieldValue.Addr().Interface())
			if err != nil {
				return err
			}
		case reflect.Map:
			err := json.Unmarshal([]byte(fContent), fieldValue.Addr().Interface())
			if err != nil {
				return err
			}
		case reflect.Struct:
			fmethod, ok := fieldValue.Addr().Interface().(json.Unmarshaler)
			if ok {
				// 对于字段是json.Unmarshaler接口的，直接调用其UnmarshalJSON方法，需要包含""引号
				err := fmethod.UnmarshalJSON([]byte("\"" + fContent + "\""))
				if err != nil {
					return err
				}
			} else {
				if err := json.Unmarshal([]byte(fContent), fieldValue.Addr().Interface()); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("not support arg %s type %s %s", key, ft.Kind(), ft.Name())
		}
		// // 由基本类型转换成可能的自定义类型 int => type A int
		// val := reflect.ValueOf(fValue).Convert(ft)
		// if isPointer {
		// 	ptr := reflect.New(val.Type())
		// 	ptr.Elem().Set(val)
		// 	fieldValue.Set(reflect.ValueOf(ptr.Interface()))
		// } else {
		// 	fieldValue.Set(val)
		// }
	}
	return nil
}

func trans[T any](t T) *T {
	return &t
}

func (x *X) SetParam(k string, v string) {
	x.Params = append(x.Params, [2]string{k, v})
}

func (x *X) Next(args ...any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s", e)
			debug.PrintStack()
		}
	}()
	if x.fid >= len(x.fcs) {
		return nil
	}
	fc := x.fcs[x.fid]
	x.fid++
	switch fc := fc.(type) {
	case fc0:
		err = fc(x)
	case fc1:
		err = fc(x.ResponseWriter, x.Request)
	case fc2:
		fc(x.ResponseWriter, x.Request)
	case fc3:
		var arg any
		arg, err = fc(x)
		args = append(args, arg)
	case fc4:
		if len(args) == 0 {
			err = fc(x, nil)
		} else {
			err = fc(x, args[0])
		}
	case fc5:
		err = fc(x, args...)
	}
	if err != nil {
		return err
	}
	return x.Next(args...)
}

func (x *X) JSON(data any) error {
	v, err := json.Marshal(data)
	if err != nil {
		return err
	}
	x.ResponseWriter.Header().Add("Content-Type", "application/json")
	x.ResponseWriter.WriteHeader(http.StatusOK)
	_, err = x.ResponseWriter.Write(v)
	return err
}

var xPool = sync.Pool{
	New: func() any {
		return &X{
			Params: make(Params, 0),
		}
	},
}

func acquire() *X {
	x := xPool.Get().(*X)
	return x
}

func release(x *X) {
	x.fid = 0
	x.Params = x.Params[0:0]
	x.Request = nil
	x.ResponseWriter = nil
	xPool.Put(x)
}

type Params [][2]string

func (ps *Params) Get(key string) (string, bool) {
	for _, p := range *ps {
		if key == p[0] {
			return p[1], true
		}
	}
	return "", false
}

func (ps *Params) GetStr(k string) string {
	v, _ := ps.Get(k)
	return v
}

func (ps *Params) GetInt(k string) int {
	v, _ := ps.Get(k)
	vv, _ := strconv.Atoi(v)
	return vv
}
