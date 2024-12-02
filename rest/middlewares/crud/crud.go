//
// crud.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-26 17:22
// Distributed under terms of the GPL license.
//

package crud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/veypi/OneBD/rest"
	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
	"gorm.io/gorm"
)

var db *gorm.DB

func SetDB(d *gorm.DB) {
	db = d
}

func CRUD(r rest.Router, obj *StructInfo) {
	idcheck := r.GetParamsList()
	for _, h := range obj.Handlers() {
		haction := h.Action
		var fn any
		switch haction {
		case "Get":
			fn = handleGetReq(h, obj, idcheck)
		case "List":
			fn = handleListReq(h, obj, idcheck)
		case "Post":
			fn = handlePostReq(h, obj, idcheck)
		case "Patch":
			fn = handlePatchReq(h, obj, idcheck)
		case "Put":
			fn = handlePutReq(h, obj, idcheck)
		case "Delete":
			fn = handleDelReq(h, obj, idcheck)
		default:
			logv.Debug().Msgf("ignore custom handle %s", h.String())
			continue
		}
		if fn != nil {
			r.Set(utils.CamelToSnake(obj.Name)+"/"+h.Suffix, h.Method, fn)
		}
	}
}

func handleGetReq(_ *StructHandler, s *StructInfo, idCheck []string) func(x *rest.X) (any, error) {
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
		err := db.Raw(sqlRaw, ids...).First(&data).Error
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
		idCon[i] = idCheck[i][1:] + " = ?"
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
				fsql := f.Key + " = ?"
				fk := f.Key
				if f.Alias != "" {
					fk = f.Alias
				}
				if optv, ok := args[fk+"_opt"].(string); ok {
					optv = strings.ToLower(optv)
					switch optv {
					case "like":
						fsql = fmt.Sprintf("%s LIKE ?", f.Key)
					case "in":
						fsql = fmt.Sprintf("%s IN (?)", f.Key)
					case "gt":
						fsql = fmt.Sprintf("%s > ?", f.Key)
					case "lt":
						fsql = fmt.Sprintf("%s < ?", f.Key)
					case "between":
						fsql = fmt.Sprintf("%s BETWEEN ? AND ?", f.Key)
						sqlArgs = append(sqlArgs, args[fk+"_opt_max"])
					}
				}
				sqlCon = append(sqlCon, fsql)
			} else if f.Type == "gorm.DeletedAt" {
				sqlCon = append(sqlCon, fmt.Sprintf("%s IS NULL", f.Key))
			}
		}
		sqlRaw := sqlRawOrigin
		if len(sqlCon) > 0 {
			// logv.Warn().Msgf("%v %s", strings.Join(sqlCon, "|"), sqlRaw)
			sqlRaw = fmt.Sprintf("%s WHERE %s ", sqlRaw, sqlCon[0])
			for _, con := range sqlCon[1:] {
				sqlRaw += fmt.Sprintf(" AND %s ", con)
			}
		}
		data := make([]map[string]interface{}, 0, 10)
		err := db.Raw(sqlRaw, sqlArgs...).Find(&data).Error
		if err != nil {
			return nil, err
		}
		return &data, nil
	}
}

func handlePostReq(h *StructHandler, s *StructInfo, idCheck []string) func(*rest.X, any) (any, error) {
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
		createdMap := make(map[string]interface{})
		for _, f := range h.Fields {
			if v, ok := args[f.Key]; ok {
				createdMap[f.Name] = v
			}
		}
		// path arg 优先级更高 避免越权写入
		for i := range plen {
			createdMap[utils.SnakeToCamel(idCheck[i][1:])] = x.Params[i][1]
		}
		data := reflect.New(s.v.Type()).Interface()
		dataElem := reflect.ValueOf(data).Elem()
		for k, v := range createdMap {
			fv := dataElem.FieldByName(k)
			if fv.IsValid() {
				if fv.CanSet() {
					fv.Set(reflect.ValueOf(v))
				}
			}
		}
		err := db.Create(data).Error
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func handlePatchReq(h *StructHandler, s *StructInfo, idCheck []string) func(*rest.X, any) (any, error) {
	sqlRaw := "id = ?"
	for _, idc := range idCheck {
		sqlRaw = fmt.Sprintf("%s AND %s = ?", sqlRaw, idc[1:])
	}
	plen := len(idCheck) + 1
	return func(x *rest.X, argsBody any) (any, error) {
		args, ok := argsBody.(map[string]interface{})
		if !ok {
			logv.Warn().Msgf("args not map[string]interface{}: %T", argsBody)
			return nil, Err500
		}
		ids := make([]any, plen)
		for i := range plen {
			ids[i] = x.Params[i][1]
		}
		data := reflect.New(s.v.Type()).Interface()
		err := db.Where(sqlRaw, ids...).First(data).Error
		if err != nil {
			return nil, err
		}
		updatedMap := make(map[string]interface{})
		for _, f := range h.Fields {
			if v, ok := args[f.Key]; ok {
				updatedMap[f.Name] = v
			}
		}
		if len(updatedMap) != 0 {
			err = db.Model(data).Updates(updatedMap).Error
		}
		return data, err
	}
}

// 批量更新或创建
func handlePutReq(_ *StructHandler, s *StructInfo, idCheck []string) func(*rest.X) (any, error) {
	sqlRaw := ""
	for _, idc := range idCheck {
		sqlRaw += fmt.Sprintf("AND %s = ? ", idc[1:])
	}
	if sqlRaw == "" {
		sqlRaw = "id = ?"
	} else {
		sqlRaw = sqlRaw[4:] + " AND id = ?"
	}
	return func(x *rest.X) (any, error) {
		items := make([]map[string]interface{}, 0, 4)
		err := json.NewDecoder(x.Request.Body).Decode(&items)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrParse, err)
		}
		ids := make([]any, len(idCheck))
		for i := range len(idCheck) {
			ids[i] = x.Params[i][1]
		}
		// created := make([]any, 0, 4)
		sliceType := reflect.SliceOf(s.v.Type())
		// 创建切片对象
		created := reflect.MakeSlice(sliceType, 0, 4)

		updated := make([]map[string]any, 0, 4)
		for _, item := range items {
			if _, ok := item["id"]; ok {
				updated = append(updated, item)
			} else {
				for i := range len(idCheck) {
					item[idCheck[i][1:]] = x.Params[i][1]
				}
				data := reflect.New(s.v.Type())
				dataElem := data.Elem()
				for k, v := range item {
					fv := dataElem.FieldByName(utils.SnakeToCamel(k))
					if fv.IsValid() {
						if fv.CanSet() {
							fv.Set(reflect.ValueOf(v))
						}
					}
				}
				created = reflect.Append(created, dataElem)
			}
		}
		data := reflect.New(s.v.Type()).Elem().Interface()
		err = db.Transaction(func(tx *gorm.DB) error {
			if created.Len() > 0 {
				err := tx.Create(created.Interface()).Error
				if err != nil {
					return err
				}
			}
			for _, item := range updated {
				err := tx.Model(data).Where(sqlRaw, append(ids, item["id"])...).Updates(item).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
		return "ok", err
	}
}

func handleDelReq(_ *StructHandler, s *StructInfo, idCheck []string) func(x *rest.X) (any, error) {
	sqlRaw := "id = ?"
	for _, idc := range idCheck {
		sqlRaw = fmt.Sprintf("%s AND %s = ?", sqlRaw, idc[1:])
	}
	plen := len(idCheck) + 1
	return func(x *rest.X) (any, error) {
		ids := make([]any, plen)
		for i := range plen {
			ids[i] = x.Params[i][1]
		}
		data := reflect.New(s.v.Type()).Interface()
		res := db.Where(sqlRaw, ids...).Delete(data)
		return res.RowsAffected, res.Error
	}
}
