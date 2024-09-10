//
// json.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-09-02 10:58
// Distributed under terms of the MIT license.
//

package middlewares

import "github.com/veypi/OneBD/rest"

func JsonResponse(x *rest.X, data any) error {
	return x.JSON(map[string]any{"code": 0, "data": data})
}

func JsonErrorResponse(x *rest.X, err error, code int) {
	x.JSON(map[string]any{"code": 1, "err": err.Error()})
}
