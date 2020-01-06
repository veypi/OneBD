package libs

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

func Defualt404(ctx core.Context) {
}

type Router struct {
	mu             sync.Mutex
	app            *Application
	name           string
	root           *Router
	subRouters     []Router
	mainHandler    core.Handler
	statusHandlers map[rfc.Status]core.BaseHandler
	logger         *zap.Logger
}

func NewRouter(app *Application) core.Router {
	r := &Router{app: app, logger: app.Logger()}
	return r
}
func (router *Router) SubRouter(name string) core.Router {
	return router
}

func (router *Router) DisableMethod(methods ...rfc.Method) {
}

func (router *Router) SetStatusHandler(status rfc.Status, fc core.BaseHandler) {
	if router.statusHandlers == nil {
		router.mu.Lock()
		router.statusHandlers = make(map[rfc.Status]core.BaseHandler, 0)
		router.mu.Unlock()
	}
	router.mu.Lock()
	router.statusHandlers[status] = fc
	router.mu.Unlock()
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.logger.Warn(r.RequestURI)
	w.Write([]byte(r.RequestURI))
}
