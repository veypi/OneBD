//
// x.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-09 13:08
// Distributed under terms of the MIT license.
//

package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

type X struct {
	http.ResponseWriter
	Request *http.Request
	Params  Params
	fcs     []any
	fid     int
}

func (x *X) SetParam(k string, v string) {
	x.Params = append(x.Params, [2]string{k, v})
}

func (x *X) Next() error {
	if x.fid >= len(x.fcs) {
		return nil
	}
	fc := x.fcs[x.fid]
	x.fid++
	var err error
	switch fc := fc.(type) {
	case fc0:
		err = fc(x)
	case fc1:
		err = fc(x.ResponseWriter, x.Request)
	case fc2:
		fc(x.ResponseWriter, x.Request)
	}
	if err != nil {
		return err
	}
	return x.Next()
}

func (x *X) next(attach any) error {
	if x.fid >= len(x.fcs) {
		return nil
	}
	fc := x.fcs[x.fid]
	x.fid++
	var err error
	switch fc := fc.(type) {
	case fc0:
		err = fc(x)
	case fc1:
		err = fc(x.ResponseWriter, x.Request)
	case fc2:
		fc(x.ResponseWriter, x.Request)
	}
	if err != nil {
		return err
	}
	return x.Next()
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
	v, ok := ps.Get(k)
	if !ok {
		return 0
	}
	vv, _ := strconv.Atoi(v)
	return vv
}
