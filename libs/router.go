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
func (router *Router) SubRouter(name string) core.Router {
	return router
}

func (router *Router) DisableMethod(methods ...rfc.Method) {
}
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.app.Logger().Warn(r.RequestURI)
	w.Write([]byte(r.RequestURI))
}
