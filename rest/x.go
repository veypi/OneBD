//
// x.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-09 13:08
// Distributed under terms of the MIT license.
//

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/veypi/utils"
)

type X struct {
	writer  http.ResponseWriter
	Request *http.Request
	code    int
	Params  Params
	fcs     []any
	fid     int
}

var _ http.ResponseWriter = &X{}

// 从不同来源解析目标结构体一级字段
// tag标签 parse:"path/header/query/form/json" 可以追加为 path@alias_name
// tag标签 default:""
// 字段为指针类型时为可选参数,defalt标签不生效
// 字段为非指针类型时是必选参数，default标签生效，未设置该值且未发现参数时报参数缺失错误
// json 字段由json解码控制，没有default机制
func (x *X) Parse(obj any) error {
	v := reflect.ValueOf(obj).Elem()
	contentType := x.Request.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		err := x.Request.ParseForm()
		if errors.Is(err, io.EOF) {
		} else if err != nil {
			return fmt.Errorf("%w: %v", ErrParse, err)
		}
	} else if strings.Contains(contentType, "application/json") {
		err := json.NewDecoder(x.Request.Body).Decode(obj)
		if errors.Is(err, io.EOF) {
		} else if err != nil {
			return fmt.Errorf("%w: %v", ErrParse, err)
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
		if method == "header" {
			// header参数默认Title形式
			key = utils.ToTitle(field.Name)
		}
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
			if field.Type.Kind() != reflect.Ptr && fieldValue.IsZero() {
				return fmt.Errorf("%w: %s@%s", ErrArgMissing, key, method)
			}
			continue
		default:
			return fmt.Errorf("%w: unknown parse method %s", ErrParse)
		}
		ft := field.Type
		isPointer := false
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
				return fmt.Errorf("%w: %s@%s", ErrArgMissing, key, method)
			}
			fContent = defaultValue
		}

		if fContent[0] == '"' && len(fContent) > 2 && fContent[len(fContent)-1] == '"' {
			// 去掉字符串两边的引号
			fContent = fContent[1 : len(fContent)-1]
		}

		var invalidArg = fmt.Errorf("%w: %s: %s", ErrArgInvalid, key, fContent)
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
		case reflect.Bool:
			n, err := strconv.ParseBool(fContent)
			if err != nil && !isPointer {
				return invalidArg
			}
			fieldValue.SetBool(n)
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
			err = fmt.Errorf("%s: %w", ErrCrash, e)
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
		err = fc(x, x.Request)
	case fc2:
		fc(x, x.Request)
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

func (x *X) Write(p []byte) (n int, err error) {
	x.writer.WriteHeader(x.code)
	return x.writer.Write(p)
}

func (x *X) WriteHeader(statusCode int) {
	x.code = statusCode
}
func (x *X) Header() http.Header {
	return x.writer.Header()
}

func (x *X) JSON(data any) error {
	v, err := json.Marshal(data)
	if err != nil {
		return err
	}
	x.Header().Add("Content-Type", "application/json")
	_, err = x.Write(v)
	return err
}
func (x *X) GetRemoteIp() string {
	// 首先尝试从 X-Forwarded-For 获取 IP 地址
	ip := x.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能包含多个 IP 地址，以逗号分隔，
		// 这里我们取第一个 IP 地址作为客户端的 IP。
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}

	// 如果 X-Forwarded-For 不存在，则尝试从 X-Real-IP 获取 IP 地址
	ip = x.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 如果以上两个都没有，则直接从 RemoteAddr 获取 IP 地址
	ip, _, err := net.SplitHostPort(x.Request.RemoteAddr)
	if err != nil {
		return ""
	}
	return ip
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

var xPool = sync.Pool{
	New: func() any {
		return &X{
			Params: make(Params, 0),
			code:   200,
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
	x.writer = nil
	x.code = 200
	x.fcs = nil
	xPool.Put(x)
}
