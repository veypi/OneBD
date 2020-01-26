package router

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/hpool"
	"github.com/lightjiang/OneBD/libs/meta"
	"github.com/lightjiang/OneBD/libs/oerr"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"github.com/lightjiang/OneBD/utils/log"
	"go.uber.org/zap"
	"net/http"
)

var app core.AppInfo
var cfg *core.Config
var logger = log.DefaultLogger
var baseRouter *route

// 触发缓存 避免每次请求都递归查询触发函数
var statusFuncCache = make(map[rfc.Status]bool)

type route struct {
	utils.FastLocker
	prefix        string // asd/{id:type}/
	trie          *trie
	root          *route
	top           *route
	subRouters    []*route
	otherHandlers map[rfc.Status]core.MetaFunc
}

func (r *route) String() string {
	return r.trie.String()
}

func (r *route) AbsPrefix() string {
	if r.root != nil {
		return r.root.AbsPrefix() + r.prefix
	}
	return r.prefix
}

// NewRouter allowedMethods 为空时默认所有方法皆允许
func NewMainRouter(_app core.AppInfo) core.Router {
	if baseRouter == nil {
		baseRouter = &route{
			trie: &trie{},
		}
		baseRouter.top = baseRouter
	}
	app = _app
	cfg = _app.Config()
	logger = app.Logger()
	return baseRouter
}

func (r *route) Set(prefix string, fc func() core.Handler, allowedMethods ...rfc.Method) {
	if len(allowedMethods) == 0 {
		allowedMethods = []rfc.Method{rfc.MethodGet}
	}
	prefix = r.AbsPrefix() + prefix
	for _, m := range allowedMethods {
		r.top.trie.Add("/"+m+prefix, hpool.NewHandlerPool(fc))
	}
}

func (r *route) SubRouter(prefix string) core.Router {
	tmp := &route{
		prefix: prefix,
		root:   r,
		top:    r.top,
	}
	return tmp
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
	//logger.Debug(req.RequestURI + " <- " + req.RemoteAddr)
	//now := time.Now()
	var t *trie
	var m core.Meta
	t = r.trie.Match("/" + req.Method + req.URL.Path)
	if t != nil && t.handler != nil {
		m = meta.Acquire(w, req, t.params, app)
		handler := t.handler.(core.HandlerPool).Acquire()
		requestCircle(handler, m)
		t.handler.(core.HandlerPool).Release(handler)
		r.fireOnStatus(m.Status(), m)

	} else {
		m = meta.Acquire(w, req, nil, app)
		m.SetStatus(rfc.StatusNotFound)
		r.fireOnStatus(rfc.StatusNotFound, m)
	}
	m.Flush()
	meta.Release(m)
	//logger.Debug(req.Method+":"+req.URL.Path,
	//	zap.Int64("delta/us", time.Now().Sub(now).Microseconds()),
	//	zap.String("addr", req.RemoteAddr))
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
		if err := recover(); err != nil {
			logger.Error("panic error", zap.Any("err", err))
			m.SetStatus(rfc.StatusInternalServerError)
		}
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
	handler.TryReset()
}
