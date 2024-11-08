//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//
package {{.package}}

import (
    "github.com/veypi/OneBD/rest"
)

func use{{.Obj}}(r rest.Router) {
	r.Get("/:{{.s_obj}}_id", {{.obj}}Get)
	r.Get("/", {{.obj}}List)
	r.Post("/", {{.obj}}Post)
	r.Patch("/:{{.s_obj}}_id", {{.obj}}Patch)
	r.Delete("/:{{.s_obj}}_id", {{.obj}}Delete)
}

func {{.obj}}Get(x *rest.X) error {
	return nil
}

func {{.obj}}List(x *rest.X) error {
	return nil
}

func {{.obj}}Post(x *rest.X) error {
	return nil
}

func {{.obj}}Patch(x *rest.X) error {
	return nil
}

func {{.obj}}Delete(x *rest.X) error {
	return nil
}

