package router

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/meta"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/utils"
	"github.com/lightjiang/utils/log"
	"net/http"
)

type route struct {
	utils.FastLocker
	prefix        string // asd/{id:type}/
	trie          *trie
	root          *route
	top           *route
	cycle         core.RequestLifeCycle
	subRouters    []*route
	otherHandlers map[rfc.Status]core.MetaFunc
	// 触发缓存 避免每次请求都递归查询触发函数
	// 仅 r.top.statusFuncCache 有用
	statusFuncCache map[rfc.Status]bool
}

func (r *route) SetRequestLifeCycle(cycle core.RequestLifeCycle) {
	r.cycle = cycle
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
func NewMainRouter() core.Router {
	baseRouter := &route{
		trie:  &trie{},
		cycle: DefaultCycle,
	}
	baseRouter.top = baseRouter
	baseRouter.statusFuncCache = make(map[rfc.Status]bool)
	return baseRouter
}

func (r *route) Set(prefix string, fc interface{}, allowedMethods ...rfc.Method) {
	if len(allowedMethods) == 0 {
		allowedMethods = []rfc.Method{rfc.MethodGet}
	}
	prefix = r.AbsPrefix() + prefix
	for _, m := range allowedMethods {
		r.top.trie.Add("/"+m+prefix, fc)
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
	var t *trie
	var m core.Meta
	defer func() {
		// 以防各种statusFunc出问题
		if err := recover(); err != nil {
			log.Error().Interface("panic", err).Msg("serve http panic")
		} else if m != nil {
			meta.Release(m)
		}
	}()
	//logger.Debug(req.RequestURI + " <- " + req.RemoteAddr)
	t = r.trie.Match("/" + req.Method + req.URL.Path)
	if t != nil && t.handler != nil {
		m = meta.Acquire(w, req, t.params)
		r.cycle(t.handler, m)
		r.fireOnStatus(m.Status(), m)

	} else {
		m = meta.Acquire(w, req, nil)
		m.SetStatus(rfc.StatusNotFound)
		r.fireOnStatus(rfc.StatusNotFound, m)
	}
	m.Flush()
	log.Debug().
		Str("addr", req.RemoteAddr).
		Int64("delta/ms", m.AliveTime().Microseconds()).
		Str("method", req.Method).
		Msg(req.URL.Path)
}

func (r *route) SetStatusFunc(status rfc.Status, fc core.MetaFunc) {
	r.Lock()
	if r.otherHandlers == nil {
		r.otherHandlers = make(map[rfc.Status]core.MetaFunc)
	}
	r.otherHandlers[status] = fc
	r.top.statusFuncCache[status] = true
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
	if _, ok := r.top.statusFuncCache[status]; !ok {
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

func DefaultCycle(fc interface{}, m core.Meta) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Interface("panic", err).Msg("default cycle panic")
			m.SetStatus(rfc.StatusInternalServerError)
		}
	}()
	switch fc := fc.(type) {
	case core.HandlerPool:
		hPoolCycle(fc, m)
	case func() core.Handler:
		handleCycle(fc(), m)
	case core.MetaFunc:
		fc(m)
	}
}

func handleCycle(handler core.Handler, m core.Meta) {
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
		m.SetStatus(rfc.StatusNotImplemented)
		return
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

func hPoolCycle(hp core.HandlerPool, m core.Meta) {
	h := hp.Acquire()
	handleCycle(h, m)
	hp.Release(h)
}
