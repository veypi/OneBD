package router

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs"
	"github.com/lightjiang/OneBD/rfc"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

var app *libs.Application
var logger *zap.Logger
var baseRouter *route

type route struct {
	mu             sync.Mutex
	prefix         string // asd/{id:type}/
	root           *route
	subRouters     []*route
	allowedMethods map[rfc.Method]bool
	mainHandler    core.HandlerPool
}

// NewRouter allowedMethods 为空时默认所有方法皆允许
func NewMainRouter(application *libs.Application, allowedMethods ...rfc.Method) core.Router {
	if baseRouter == nil {
		baseRouter = &route{}
	}
	app = application
	logger = application.Logger()
	baseRouter.EnableMethod(allowedMethods...)
	return baseRouter
}
func (router *route) Set(prefix string, hp core.HandlerPool, allowedMethods ...rfc.Method) {
	r := route{
		prefix:      prefix,
		root:        router,
		mainHandler: hp,
	}
	r.EnableMethod(allowedMethods...)
	router.subRouters = append(router.subRouters, &r)
}

func (router *route) SubRouter(prefix string) core.Router {
	return router
}

func (router *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Warn(r.RequestURI)
	w.Write([]byte(r.RequestURI))
}

func (router *route) EnableMethod(methods ...rfc.Method) {
	if router.allowedMethods == nil {
		router.allowedMethods = make(map[rfc.Method]bool)
	}
	for _, m := range methods {
		router.allowedMethods[m] = true
	}
}

func (router *route) DisableMethod(methods ...rfc.Method) {
	if router.allowedMethods == nil {
		router.allowedMethods = make(map[rfc.Method]bool)
	}
	for _, m := range methods {
		router.allowedMethods[m] = false
	}
}
