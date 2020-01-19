package router

import (
	"fmt"
	"github.com/lightjiang/OneBD/config"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/meta"
	"github.com/lightjiang/OneBD/libs/oerr"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

var cfg *config.Config
var logger *zap.Logger
var baseRouter *route

// 触发缓存 避免每次请求都递归查询触发函数
var statusFuncCache = make(map[rfc.Status]bool)

type route struct {
	utils.FastLocker
	prefix         string // asd/{id:type}/
	root           *route
	subRouters     []*route
	allowedMethods map[rfc.Method]bool
	mainHandler    core.HandlerPool
	otherHandlers  map[rfc.Status]core.MetaFunc
}

// NewRouter allowedMethods 为空时默认所有方法皆允许
func NewMainRouter(_cfg *config.Config, allowedMethods ...rfc.Method) core.Router {
	if baseRouter == nil {
		baseRouter = &route{}
	}
	cfg = _cfg
	logger = cfg.Logger
	baseRouter.EnableMethod(allowedMethods...)
	return baseRouter
}
func (r *route) Set(prefix string, hp core.HandlerPool, allowedMethods ...rfc.Method) {
	tmp := route{
		prefix:      prefix,
		root:        r,
		mainHandler: hp,
	}
	tmp.EnableMethod(allowedMethods...)
	r.subRouters = append(r.subRouters, &tmp)
}

func (r *route) SubRouter(prefix string) core.Router {
	tmp := &route{
		prefix: prefix,
		root:   r,
	}
	r.subRouters = append(r.subRouters, tmp)
	return r
}

func (r *route) match(url string) *route {
	// TODO:: 是否有必要对路由查询结果做缓存
	if url == r.prefix {
		return r
	}
	var tmp *route
	for _, item := range r.subRouters {
		tmp = item.match(url)
		if tmp != nil {
			return tmp
		}
	}
	return nil
}

/*
 ResponseWriter 生命周期 由HTTP协议决定
	- 1 : w.Header().Set(key, value)
	- 2 : w.WriteHeader(status_code)
	- 3 : w.Write(data)
*/
func (r *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger.Info(strings.Join([]string{req.RequestURI, " <- ", req.RemoteAddr}, ""))
	hp := r.match(req.URL.Path)
	if hp != nil {
		handler := hp.mainHandler.Acquire()
		requestCircle(handler, w, req)
		r.FireOnStatus(handler.Meta().Status(), handler.Meta())
	} else {
		r.FireOnStatus(rfc.StatusNotFound, meta.New(w, req))
		w.WriteHeader(int(rfc.StatusNotFound))
	}
	w.Write([]byte(fmt.Sprintf("%#v", req.URL)))
}

func (r *route) EnableMethod(methods ...rfc.Method) {
	if r.allowedMethods == nil {
		r.allowedMethods = make(map[rfc.Method]bool)
	}
	for _, m := range methods {
		r.allowedMethods[m] = true
	}
}

func (r *route) DisableMethod(methods ...rfc.Method) {
	if r.allowedMethods == nil {
		r.allowedMethods = make(map[rfc.Method]bool)
	}
	for _, m := range methods {
		r.allowedMethods[m] = false
	}
}

func (r *route) SetStatusFunc(status rfc.Status, fc core.MetaFunc) {
	r.Lock()
	if r.otherHandlers == nil {
		r.otherHandlers = make(map[rfc.Status]core.MetaFunc)
	}
	r.otherHandlers[status] = fc
	statusFuncCache[status] = true
	r.Unlock()
}

func (r *route) SetNotFoundFunc(fc core.MetaFunc) {
	r.SetStatusFunc(rfc.StatusNotFound, fc)
}

// 递归向上直到触发或到树顶
func (r *route) FireOnStatus(status rfc.Status, m core.Meta) {
	if _, ok := statusFuncCache[status]; !ok {
		return
	}
	if r.otherHandlers != nil {
		fc, ok := r.otherHandlers[status]
		if ok && fc != nil {
			fc(m)
			return
		}
	}
	if r.root != nil {
		r.root.FireOnStatus(status, m)
	}
}

func requestCircle(handler core.Handler, w http.ResponseWriter, r *http.Request) {
	defer func() {
		handler.TryReset()
	}()
	var data interface{}
	var err error
	err = handler.Init(w, r)
	if err != nil {
		handler.OnError(err)
		return
	}
	switch handler.Meta().Method() {
	case rfc.MethodGet:
		data, err = handler.Get()
	case rfc.MethodPost:
		data, err = handler.Post()
	case rfc.MethodPut:
		data, err = handler.Put()
	case rfc.MethodDelete:
		data, err = handler.Delete()
	case rfc.MethodPatch:
		data, err = handler.Patch()
	case rfc.MethodHead:
		data, err = handler.Head()
	case rfc.MethodOptions:
		data, err = handler.Options()
	case rfc.MethodTrace:
		data, err = handler.Trace()
	default:
		panic(oerr.MethodNotSupport)
	}
	if err != nil {
		handler.OnError(err)
		return
	}
	err = handler.Finished()
	if err != nil {
		handler.OnError(err)
		return
	}
	handler.OnResponse(data)
}
