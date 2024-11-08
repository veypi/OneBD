//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//
package {{.package}}

import (
    "github.com/veypi/OneBD/rest"
    M {{.mimport}}
)

func use{{.Obj}}(r rest.Router) {
	r.Get("/:{{.s_obj}}_id", {{.obj}}Get)
	r.Get("/", {{.obj}}List)
	r.Post("/", {{.obj}}Post)
	r.Put("/:{{.s_obj}}_id", {{.obj}}Put)
	r.Patch("/:{{.s_obj}}_id", {{.obj}}Patch)
	r.Delete("/:{{.s_obj}}_id", {{.obj}}Delete)
}

func {{.obj}}Get(x *rest.X) (any, error) {
	opts := M.{{.Obj}}Get{}
	err := opts.Parse(x)
	if err != nil {
		return nil, err
	}
	data := M.{{.Obj}}{}
	// edit here

	return data, nil
}

func {{.obj}}List(x *rest.X) (any, error) {
	opts := M.{{.Obj}}List{}
	err := opts.Parse(x)
	if err != nil {
		return nil, err
	}
    data := []M.{{.Obj}}{}
	// edit here

	return data, nil
}

func {{.obj}}Post(x *rest.X) (any, error) {
	opts := M.{{.Obj}}Post{}
	err := opts.Parse(x)
	if err != nil {
		return nil, err
	}
    data := M.{{.Obj}}{}
	return data, nil
}

func {{.obj}}Put(x *rest.X) (any, error) {
	opts := M.{{.Obj}}Put{}
	err := opts.Parse(x)
	if err != nil {
		return nil, err
	}
    data := M.{{.Obj}}{}
	return data, nil
}

func {{.obj}}Patch(x *rest.X) (any, error) {
	opts := M.{{.Obj}}Patch{}
	err := opts.Parse(x)
	if err != nil {
		return nil, err
	}
    data := M.{{.Obj}}{}
	return data, nil
}

func {{.obj}}Delete(x *rest.X) error {
	opts := M.{{.Obj}}Delete{}
	err := opts.Parse(x)
	if err != nil {
		return err
	}
	return nil
}
