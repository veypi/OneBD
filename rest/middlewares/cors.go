//
// cors.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-04 21:18
// Distributed under terms of the GPL license.
//

package middlewares

import (
	"net/http"

	"github.com/veypi/OneBD/rest"
)

func CorsAllowAny(x *rest.X) {
	origin := x.Request.Header.Get("Origin")
	x.Header().Set("Access-Control-Allow-Origin", origin)
	x.Header().Set("Access-Control-Allow-Credentials", "true")
	x.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	x.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if x.Request.Method == http.MethodOptions && x.Request.Header.Get("Access-Control-Request-Method") != "" {
		x.Stop()
	}
}

func CorsAllow(domains ...string) func(x *rest.X) {
	return func(x *rest.X) {
		origin := x.Request.Header.Get("Origin")
		for _, d := range domains {
			if d == origin {
				x.Header().Set("Access-Control-Allow-Origin", origin)
				x.Header().Set("Access-Control-Allow-Credentials", "true")
				x.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
				x.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
				if x.Request.Method == http.MethodOptions && x.Request.Header.Get("Access-Control-Request-Method") != "" {
					x.Stop()
				}
				return
			}
		}
	}
}
