package libs

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"net/http"
)

type Router struct {
	app *Application
}

func NewRouter(app *Application) core.Router {
	r := &Router{app: app}
	return r
}
func (r *Router) SubRouter(name string) core.Router {
	return r
}

func (r *Router) DisableMethod(methods ...rfc.Method) {
}
func (r *Router) ServeHTTP(http.ResponseWriter, *http.Request) {}
