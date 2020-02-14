package router

import (
	"fmt"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/meta"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/utils"
	"github.com/lightjiang/utils/log"
	"net/http"
	"os"
	"strings"
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
		return urlAppend(r.root.AbsPrefix(), r.prefix)
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
	prefix = urlAppend(r.AbsPrefix(), prefix)
	for _, m := range allowedMethods {
		t, err := r.top.trie.Add(urlAppend(m, prefix), fc)
		if err != nil {
			log.HandlerErrs(err)
		} else {
			t.router = r
		}
	}
}

func (r *route) Static(prefix string, directory string) {
	dir, err := os.Stat(directory)
	if err != nil {
		panic(err)
	}
	if !dir.IsDir() {
		r.Set(prefix, func(m core.Meta) {
			f, err := os.Open(directory)
			if err != nil {
				panic(err)
			}
			info, err := f.Stat()
			if err != nil {
				m.WriteHeader(rfc.StatusNotFound)
				return
			}
			http.ServeContent(m, m.Request(), info.Name(), info.ModTime(), f)
		})
		return
	}
	if strings.Contains(prefix, "*") {
		panic("static prefix should not contain *")
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefix += "*path"
	var fs http.FileSystem = http.Dir(directory)
	r.Set(prefix, func(m core.Meta) {
		name := m.Params("path")
		log.Warn().Msg("static " + name)
		f, err := fs.Open(name)
		if err != nil {
			m.WriteHeader(rfc.StatusNotFound)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			m.WriteHeader(rfc.StatusNotFound)
			return
		}
		if info.IsDir() {
			// TODO:: dir list
			m.WriteHeader(rfc.StatusNotFound)
			return
		}
		http.ServeContent(m, m.Request(), info.Name(), info.ModTime(), f)
	}, rfc.MethodGet)
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
		log.Warn().Msg("123")
		t.router.fireOnStatus(m.Status(), m)

	} else {
		m = meta.Acquire(w, req, nil)
		m.WriteHeader(rfc.StatusNotFound)
		r.fireOnStatus(rfc.StatusNotFound, m)
	}
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
			m.WriteHeader(rfc.StatusInternalServerError)
		}
	}()
	switch fc := fc.(type) {
	case core.HandlerPool:
		hPoolCycle(fc, m)
	case func() core.Handler:
		handleCycle(fc(), m)
	case core.MetaFunc:
		fc(m)
	default:
		fmt.Printf("> %t \n", fc)
		log.Warn().Interface("fc", fc).Msg("unknown fc to handle data")
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
		m.WriteHeader(rfc.StatusNotImplemented)
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

func urlAppend(s1, s2 string) string {
	if len(s1) == 0 {
		s1 = "/"
	}
	if s1[len(s1)-1] != '/' {
		s1 += "/"
	}
	if len(s2) != 0 && s2[0] == '/' {
		s2 = s2[1:]
	}
	return s1 + s2
}
