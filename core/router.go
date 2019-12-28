package core

import (
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
)

type Router interface {
	SubRouter(name string) Router
	DisableMethod(methods ...rfc.Method)
	ServeHTTP(http.ResponseWriter, *http.Request)
}
