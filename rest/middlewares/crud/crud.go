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
		logv.Warn().Msgf("|%s", t.TableName)
		for _, h := range t.Handlers() {
			haction := h.Action
			var fn func(x *rest.X) (any, error)
			switch haction {
			case "Get":
				fn = handleGetReq(h, t, h.Suffix[1:], idcheck)
			case "List":
			case "Post":
			case "Patch":
			case "Put":
			case "Delete":
			default:
				logv.Debug().Msgf("ignore custom handle %s", h.String())
				continue
			}
			r.Set(utils.CamelToSnake(t.Name)+"/"+h.Suffix, h.Method, fn)
		}
	}
}

//	func crud(r rest.Router, parents []*StructInfo) crudfn {
//		return func(idx uint, connect StructConnect, objs ...any) crudfn {
//			p := parents[idx]
//			snaker := utils.CamelToSnake(p.Name)
//			subr := r.SubRouter(fmt.Sprintf("/%s/:%s", snaker, snaker))
//			targents := Parse(objs...)
//			return crud(subr, targents)
//		}
//	}
func handleGetReq(h *structHandler, s *StructInfo, id string, idCheck []string) func(x *rest.X) (any, error) {
	feilds := ""
	for _, f := range s.Fields {
		feilds += "," + f.Key
	}
	sqlRaw := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", feilds[1:], s.TableName)
	for _, idc := range idCheck {
		sqlRaw = fmt.Sprintf("%s AND %s = ?", sqlRaw, idc[1:])
	}
	// if len(idCheck) > 0 {
	// 	sqlRaw += " AND id = ?"
	// } else {
	// 	sqlRaw += " id = ?"
	// }
	plen := len(idCheck) + 1
	return func(x *rest.X) (any, error) {
		if len(x.Params) != plen {
			return nil, ErrMissParam.Fmt("path id")
		}
		data := make(map[string]interface{})
		ids := make([]any, len(idCheck)+1)
		for i := range plen {
			ids[i] = x.Params[i][1]
		}
		err := db.Debug().Raw(sqlRaw, ids...).First(&data).Error
		if err != nil {
			return nil, err
		}
		logv.WithNoCaller.Debug().Msgf("auto handle %s: %s /%s", h.String(), h.Method, h.Suffix)
		return &data, nil
	}
}
