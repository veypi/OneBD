//
// crud.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-26 17:22
// Distributed under terms of the GPL license.
//

package crud

import (
	"fmt"
	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
	"gorm.io/gorm"
)

type StructConnect uint

var db *gorm.DB

func SetDB(d *gorm.DB) {
	db = d
}

func crud(r rest.Router, objs ...*StructInfo) {
	idcheck := r.GetParamsList()
	for _, t := range objs {
		logv.Warn().Msgf("|%s\n%s", t.TableName, t)
		for _, h := range t.Handlers() {
			haction := h.Action
			var fn any
			switch haction {
			case "Get":
				fn = handleGetReq(h, t, idcheck)
			case "List":
				fn = handleListReq(h, t, idcheck)
			case "Post":
			case "Patch":
			case "Put":
			case "Delete":
			default:
				logv.Debug().Msgf("ignore custom handle %s", h.String())
				continue
			}
			if fn != nil {
				r.Set(utils.CamelToSnake(t.Name)+"/"+h.Suffix, h.Method, fn)
			}
		}
	}
}

func handleGetReq(h *StructHandler, s *StructInfo, idCheck []string) func(x *rest.X) (any, error) {
	feilds := ""
	for _, f := range s.Fields {
		feilds += "," + f.Key
	}
	sqlRaw := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", feilds[1:], s.TableName)
	for _, idc := range idCheck {
		sqlRaw = fmt.Sprintf("%s AND %s = ?", sqlRaw, idc[1:])
	}
	plen := len(idCheck) + 1
	return func(x *rest.X) (any, error) {
		if len(x.Params) != plen {
			return nil, ErrMissArg.Fmt("path id")
		}
		data := make(map[string]interface{})
		ids := make([]any, plen)
		for i := range plen {
			ids[i] = x.Params[i][1]
		}
		err := db.Debug().Raw(sqlRaw, ids...).First(&data).Error
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}

func handleListReq(h *StructHandler, s *StructInfo, idCheck []string) func(*rest.X, any) (any, error) {
	feilds := ""
	for _, f := range s.Fields {
		feilds += "," + f.Key
	}
	sqlRawOrigin := fmt.Sprintf("SELECT %s FROM %s", feilds[1:], s.TableName)
	idCon := make([]string, len(idCheck))
	for i := range idCheck {
		idCon[i] = idCheck[i][1:]
	}
	plen := len(idCheck)
	return func(x *rest.X, argsBody any) (any, error) {
		args, ok := argsBody.(map[string]interface{})
		if !ok {
			logv.Warn().Msgf("args not map[string]interface{}: %T", argsBody)
			return nil, Err500
		}
		if len(x.Params) != plen {
			return nil, ErrMissArg.Fmt("path id")
		}
		sqlArgs := make([]any, plen, plen+5)
		for i := range plen {
			sqlArgs[i] = x.Params[i][1]
		}
		sqlCon := idCon[:]
		for _, f := range h.Fields {
			if v, ok := args[f.Key]; ok {
				sqlArgs = append(sqlArgs, v)
				sqlCon = append(sqlCon, f.Key)
			}
		}
		sqlRaw := sqlRawOrigin
		if len(sqlCon) > 0 {
			// logv.Warn().Msgf("%v %s", strings.Join(sqlCon, "|"), sqlRaw)
			sqlRaw = fmt.Sprintf("%s WHERE %s = ? ", sqlRaw, sqlCon[0])
			for _, con := range sqlCon[1:] {
				sqlRaw += fmt.Sprintf("AND %s = ? ", con)
			}
		}
		data := make([]map[string]interface{}, 0, 10)
		err := db.Debug().Raw(sqlRaw, sqlArgs...).Find(&data).Error
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}
