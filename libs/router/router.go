package router

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/hpool"
	"github.com/lightjiang/OneBD/libs/meta"
	"github.com/lightjiang/OneBD/libs/oerr"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var app core.AppInfo
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
func NewMainRouter(_app core.AppInfo) core.Router {
	if baseRouter == nil {
		baseRouter = &route{}
	}
	app = _app
	logger = app.Logger()
	return baseRouter
}
func (r *route) Set(prefix string, fc func() core.Handler, allowedMethods ...rfc.Method) {
	tmp := route{
		prefix:      prefix,
		root:        r,
		mainHandler: hpool.NewHandlerPool(fc, app),
	}
	if len(allowedMethods) == 0 {
		allowedMethods = []rfc.Method{rfc.MethodAll}
	}
	tmp.enableMethod(allowedMethods...)
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
	defer func() {
		// 以防各种statusFunc出问题
		if err := recover(); err != nil {
			logger.Error("panic error", zap.Any("err", err))
		}
	}()
	logger.Info(req.RequestURI + " <- " + req.RemoteAddr)
	now := time.Now()
	hp := r.match(req.URL.Path)
	m := meta.Acquire(w, req, app)
	if hp != nil {
		handler := hp.mainHandler.Acquire()
		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic error", zap.Any("err", err))
					handler.Meta().SetStatus(rfc.StatusInternalServerError)
				}
			}()
			requestCircle(handler, m)
		}()
		r.fireOnStatus(handler.Meta().Status(), handler.Meta())
		hp.mainHandler.Release(handler)

	} else {
		r.fireOnStatus(rfc.StatusNotFound, m)
		w.WriteHeader(int(rfc.StatusNotFound))
	}
	meta.Release(m)
	logger.Info(req.RequestURI+" -> "+req.RemoteAddr, zap.Int64("delta/us", time.Now().Sub(now).Microseconds()))
}

func (r *route) enableMethod(methods ...rfc.Method) {
	if r.allowedMethods == nil {
		r.allowedMethods = make(map[rfc.Method]bool)
	}
	for _, m := range methods {
		r.allowedMethods[m] = true
	}
}

func (r *route) disableMethod(methods ...rfc.Method) {
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

func (r *route) SetInternalErrorFunc(fc core.MetaFunc) {
	r.SetStatusFunc(rfc.StatusInternalServerError, fc)
}

// 递归向上直到触发或到树顶
func (r *route) fireOnStatus(status rfc.Status, m core.Meta) {
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
		r.root.fireOnStatus(status, m)
	}
}

func requestCircle(handler core.Handler, m core.Meta) {
	defer func() {
		handler.TryReset()
	}()
	var data interface{}
	var err error
	err = handler.Init(m)
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
