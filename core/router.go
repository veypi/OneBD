package core

import (
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
)

type Router interface {
	Set(prefix string, hp HandlerPool, allowedMethods ...rfc.Method)
	SubRouter(name string) Router
	ServeHTTP(http.ResponseWriter, *http.Request)
}
