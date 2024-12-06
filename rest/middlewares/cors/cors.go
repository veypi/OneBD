//
// cors.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-12-06 15:56
// Distributed under terms of the MIT license.
//

package cors

import (
	"net/http"

	"github.com/veypi/OneBD/rest"
)

func AllowAny(x *rest.X) {
	origin := x.Request.Header.Get("Origin")
	x.Header().Set("Access-Control-Allow-Origin", origin)
	x.Header().Set("Access-Control-Allow-Credentials", "true")
	x.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH, PROPFIND")
	x.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, depth")
	if x.Request.Method == http.MethodOptions && x.Request.Header.Get("Access-Control-Request-Method") != "" {
		x.Stop()
	}
}
