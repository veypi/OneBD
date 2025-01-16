//
// args_parser.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-28 22:34
// Distributed under terms of the GPL license.
//

package crud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils"
)

func ArgParser(r rest.Router, obj *StructInfo) {
	for _, h := range obj.Handlers() {
		if len(h.Fields) == 0 {
			continue
		}
		var fn = func(x *rest.X) (any, error) {
			args := make(map[string]any)
			var queryMap url.Values
			contentType := x.Request.Header.Get("Content-Type")
			if contentType == "application/x-www-form-urlencoded" {
				err := x.Request.ParseForm()
				if errors.Is(err, io.EOF) {
				} else if err != nil {
					return nil, fmt.Errorf("%w: %v", ErrParse, err)
				}
				for tk := range x.Request.Form {
					args[tk] = x.Request.Form[tk][0]
				}
			} else if strings.Contains(contentType, "application/json") {
				err := json.NewDecoder(x.Request.Body).Decode(&args)
				if errors.Is(err, io.EOF) {
				} else if err != nil {
					return nil, fmt.Errorf("%w: %v", ErrParse, err)
				}
			}
			var err error
			for _, f := range h.Fields {
				var fbody any
				fset := false
				fk := f.Key
				if f.Alias != "" {
					fk = f.Alias
				}
				switch f.Src {
				case "path":
					fbody, fset = x.Params.Get(fk)
					if fset {
						args[f.Key] = fbody
					}
				case "header":
					if tmps, ok := x.Request.Header[fk]; ok {
						fset = true
						args[f.Key] = tmps[0]
					}
				case "query":
					if queryMap == nil {
						queryMap = x.Request.URL.Query()
						for tk := range queryMap {
							args[tk] = queryMap[tk][0]
						}
					}
					if tmps, ok := queryMap[fk]; ok {
						fset = true
						args[f.Key] = tmps[0]
					}
				case "form":
					if tmps, ok := x.Request.Form[fk]; ok {
						fset = true
						args[f.Key] = tmps[0]
					}
				default:
					// for json
					if tmps, ok := args[fk]; ok {
						fset = true
						args[f.Key] = tmps
					}
				}
				// logv.Debug().Msgf("field %v %s: %s, value: %v", f.HasStar, f.Src, fk, args[f.Key])
				if fset {
					// do not use srcAlias
				} else if !f.HasStar {
					err = ErrMissArg.Fmt(fk + " from " + f.Src)
					break
				}
			}
			if queryMap == nil {
				queryMap = x.Request.URL.Query()
			}
			if queryMap.Get("api_info") != "" {
				x.Skip()
				return h.Fields, nil
			}
			return args, err
		}
		haction := h.Action
		switch haction {
		case "List", "Post", "Patch":
		case "Put", "Get", "Delete":
			// get, delete无需解析参数，仅需要path参数
			// put 由crud解析json slice数据包
			continue
		default:
			// ignore custom actor
			// continue
		}
		r.Set(utils.CamelToSnake(obj.Name)+"/"+h.Suffix, h.Method, fn)
	}
}
