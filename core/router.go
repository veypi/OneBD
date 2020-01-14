package core

import (
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
)

type Router interface {
	SubRouter(name string) Router
	Set(prefix string, hp HandlerPool, allowedMethods ...rfc.Method)
	ServeHTTP(http.ResponseWriter, *http.Request)
}
