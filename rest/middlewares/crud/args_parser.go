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

	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

func argParser(r rest.Router, objs ...*StructInfo) {
	for _, t := range objs {
		for _, h := range t.Handlers() {
			if len(h.Fields) == 0 {
				continue
			}
			var fn = func(x *rest.X) (any, error) {
				args := make(map[string]any)
				jsonData := make(map[string]any)
				formParsed := false
				jsonParsed := false
				var queryMap url.Values
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
					case "header":
						if tmps, ok := x.Request.Header[fk]; ok {
							fset = true
							fbody = tmps[0]
						}
					case "query":
						if queryMap == nil {
							queryMap = x.Request.URL.Query()
						}
						if tmps, ok := queryMap[fk]; ok {
							fset = true
							fbody = tmps[0]
						}
					case "form":
						if !formParsed {
							formParsed = true
							err := x.Request.ParseForm()
							if errors.Is(err, io.EOF) {
							} else if err != nil {
								return nil, fmt.Errorf("%w: %v", ErrParse, err)
							}
						}
						if tmps, ok := x.Request.Form[fk]; ok {
							fset = true
							fbody = tmps[0]
						}
					default:
						// json
						if !jsonParsed {
							jsonParsed = true
							err := json.NewDecoder(x.Request.Body).Decode(&jsonData)
							if errors.Is(err, io.EOF) {
							} else if err != nil {
								return nil, fmt.Errorf("%w: %v", ErrParse, err)
							}
						}
						if tmps, ok := jsonData[fk]; ok {
							fset = true
							fbody = tmps
						}
					}
					logv.Debug().Msgf("field %v %s: %s, value: %v", f.HasStar, f.Src, fk, fbody)
					if fset {
						// do not use srcAlias
						args[f.Key] = fbody
					} else if !f.HasStar {
						return nil, ErrMissArg.Fmt(fk + " from " + f.Src)
					}
				}
				return args, nil
			}
			// haction := h.Action
			// switch haction {
			// case "List":
			// case "Post":
			// case "Patch":
			// case "Put":
			// case "Get", "Delete":
			// default:
			// }
			r.Set(utils.CamelToSnake(t.Name)+"/"+h.Suffix, h.Method, fn)
		}
	}
}
