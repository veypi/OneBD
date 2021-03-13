package router

import (
	"embed"
	"github.com/veypi/OneBD/core"
	"github.com/veypi/OneBD/libs/meta"
	"github.com/veypi/OneBD/rfc"
	"github.com/veypi/utils"
	"github.com/veypi/utils/log"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// NewRouter allowedMethods 为空时默认所有方法皆允许
func NewMainRouter() core.Router {
	baseRouter := &route{
		cycle: DefaultCycle,
	}
	baseRouter.top = baseRouter
	baseRouter.statusFuncCache = make(map[rfc.Status]bool)
	return baseRouter
}

type route struct {
	utils.FastLocker
	fragment string
	depth    uint
	// 记录变量偏移，以/计数
	params map[string]uint
	// params 和handler 同时存在
	handler       map[rfc.Method]interface{}
	parent        *route
	wildcard      *route
	colon         *route
	top           *route
	cycle         core.RequestLifeCycle
	subRouters    map[string]*route
	otherHandlers map[rfc.Status]core.MetaFunc
	// 触发缓存 避免每次请求都递归查询触发函数
	// 仅 testR.top.statusFuncCache 有用
	statusFuncCache map[rfc.Status]bool
}

func (r *route) SetRequestLifeCycle(cycle core.RequestLifeCycle) {
	r.cycle = cycle
}

func (r *route) String() string {
	return strings.Join(r.string(), "\n")
}

func (r *route) string() []string {
	fc := func(res []string, subt *route) []string {
		if subt != nil {
			for _, s := range subt.string() {
				res = append(res, s)
			}
		}
		return res
	}
	res := make([]string, 0, 10)
	if r.handler != nil && len(r.handler) > 0 {
		item := r.AbsPath()
		item = "\033[32m" + item + "\033[0m"
		for m := range r.handler {
			item += " " + m
		}
		for p, v := range r.params {
			item += " " + p + "(" + strconv.Itoa(int(v)) + ")"
		}
		res = append(res, item)
	}
	for _, subT := range r.subRouters {
		res = fc(res, subT)
	}
	res = fc(res, r.colon)
	res = fc(res, r.wildcard)
	return res
}

func (r *route) AbsPath() string {
	if r.parent != nil {
		return r.parent.AbsPath() + "/" + r.fragment
	}
	return r.fragment
}

// Add add sub route, wildcardRouter and subRouters[''] do not have sub router
func (r *route) Add(url string) *route {
	last := r
	if url == "" {
		return r
	}
	startIdx := 0
	if url[0] == '/' {
		url = url[1:]
	}
	url += "/"
	var next *route
	for i, c := range url {
		if c == '/' {
			next = &route{
				fragment: url[startIdx:i],
				depth:    last.depth + 1,
				params:   nil,
				parent:   last,
				top:      last.top,
				cycle:    last.cycle,
			}
			startIdx = i + 1
			next.copyParams()
			if next.fragment == "" {
			} else if next.fragment[0] == '*' {
				if last.wildcard != nil {
					return r.wildcard
				}
				//next.fragment = "*"
				last.wildcard = next
				return next
			} else if next.fragment[0] == ':' {
				if last.colon != nil {
					last = last.colon
				} else {
					last.colon = next
					last = next
				}
				//next.fragment = "@"
				continue
			}

			if last.subRouters == nil {
				last.subRouters = make(map[string]*route)
			}
			if tmp := last.subRouters[next.fragment]; tmp != nil {
				last = tmp
			} else {
				last.subRouters[next.fragment] = next
				last = next
			}
		}
	}
	return last
	//return testR.add(strings.Split(strings.TrimPrefix(url, "/"), "/"))
}

func (r *route) copyParams() {
	r.params = make(map[string]uint)
	for k, v := range r.parent.params {
		r.params[k] = v
	}
	if r.fragment == "" {
		return
	}
	if r.fragment[0] == '*' || r.fragment[0] == ':' {
		r.params[r.fragment] = r.depth
	}
}

func (r *route) Set(prefix string, fc interface{}, allowedMethods ...rfc.Method) core.Router {
	if len(allowedMethods) == 0 {
		allowedMethods = []rfc.Method{rfc.MethodGet}
	}
	var tmp *route
	if r.fragment == "" && r.parent != nil {
		tmp = r.parent.Add(prefix)
	} else if len(r.fragment) > 0 && r.fragment[0] == '*' {
		tmp = r
	} else {
		tmp = r.Add(prefix)
	}
	if tmp.handler == nil {
		tmp.handler = make(map[rfc.Method]interface{})
	}
	for _, m := range allowedMethods {
		if t := tmp.handler[m]; t != nil {
			panic("UrlDefinedDuplicate")
		}
		tmp.handler[m] = fc
	}
	return tmp
}

func (r *route) WS(prefix string, upgrader core.WebSocketFunc) core.Router {
	return r.Set(prefix, func(m core.Meta) {
		m.DisableSelfWriter()
		c, err := upgrader(m)
		if err != nil {
			log.Warn().Msg(err.Error())
			return
		}
		defer c.Close()
		c.Wait()
	})
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

func (r *route) EmbedFile(prefix string, path []byte) {
	r.Set(prefix, func(m core.Meta) {
		_, err := m.Write(path)
		if err != nil {
			log.Info().Msgf("write file failed: %s", err.Error())
			m.WriteHeader(rfc.StatusBadRequest)
		} else {
			m.WriteHeader(rfc.StatusOK)
		}
	})
}

func (r *route) EmbedDir(prefix string, fs embed.FS, fsPrefix string) {
	if strings.Contains(prefix, "*") {
		panic("static prefix should not contain *")
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	prefix += "*path"
	r.Set(prefix, func(m core.Meta) {
		name := fsPrefix + m.Params("path")
		f, err := fs.Open(name)
		if err != nil {
			log.Info().Msgf("serve file failed: %s", err.Error())
			m.WriteHeader(rfc.StatusNotFound)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			log.Info().Msgf("serve file failed: %s", err.Error())
			m.WriteHeader(rfc.StatusNotFound)
			return
		}
		if info.IsDir() {
			// TODO:: dir list
			log.Info().Msgf("serve file failed: %s", err.Error())
			m.WriteHeader(rfc.StatusNotFound)
			return
		}
		http.ServeContent(m, m.Request(), info.Name(), info.ModTime(), f.(io.ReadSeeker))
	}, rfc.MethodGet)
}

func (r *route) SubRouter(prefix string) core.Router {
	if len(prefix) > 1 && prefix[len(prefix)-1] == '/' {
		prefix = prefix[:len(prefix)-1]
	}
	return r.Add(prefix)
}

/*
 ResponseWriter 生命周期 由HTTP协议决定
	- 1 : w.Header().Set(key, value)
	- 2 : w.WriteHeader(status_code)
	- 3 : w.Write(data)
*/
func (r *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var m core.Meta
	defer func() {
		// 以防各种statusFunc出问题
		if err := recover(); err != nil {
			log.Error().Err(nil).Msgf("%s", err)
			m.WriteHeader(rfc.StatusInternalServerError)
		}
		if m != nil {
			log.WithNoCaller.Debug().Int("status", m.Status()).
				Str("addr", req.RemoteAddr).
				Int64("delta/ms", m.AliveTime().Milliseconds()).
				Str("method", req.Method).
				Msg(req.URL.Path)
			meta.Release(m)
		}
	}()
	subR := r
	startIdx := 0
	p := req.URL.Path[1:] + "/"
	for i, v := range p {
		if v == '/' {
			if subR.subRouters != nil && subR.subRouters[p[startIdx:i]] != nil {
				subR = subR.subRouters[p[startIdx:i]]
			} else if subR.colon != nil {
				subR = subR.colon
			} else if subR.wildcard != nil {
				subR = subR.wildcard
				break
			} else {
				m = meta.Acquire(w, req, nil)
				m.WriteHeader(rfc.StatusNotFound)
				subR.fireOnStatus(rfc.StatusNotFound, m)
				return
			}
			startIdx = i + 1
		}
	}
	if subR.handler != nil && subR.handler[req.Method] != nil {
		m = meta.Acquire(w, req, subR.params)
		r.cycle(subR.handler[req.Method], m)
		subR.fireOnStatus(m.Status(), m)
	} else {
		m = meta.Acquire(w, req, nil)
		m.WriteHeader(rfc.StatusNotFound)
		subR.fireOnStatus(rfc.StatusNotFound, m)
	}
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
		if fc := r.otherHandlers[status]; fc != nil {
			fc(m)
			return
		}
	}
	if r.parent != nil {
		r.parent.fireOnStatus(status, m)
	}
}

func DefaultCycle(fc interface{}, m core.Meta) {
	defer func() {
		if err := recover(); err != nil {
			m.WriteHeader(rfc.StatusInternalServerError)
			log.Error().Err(nil).Msgf("%s", err)
			//log.Error().Interface("panic", err).Msg("default cycle panic")
		}
	}()
	switch fc := fc.(type) {
	case core.HandlerPool:
		HPoolCycle(fc, m)
	case func() core.Handler:
		HandleCycle(fc(), m)
	case core.MetaFunc:
		fc(m)
	case core.OriginFunc:
		fc(m.ResponseWriter(), m.Request())
	default:
		log.Warn().Interface("fc", fc).Msg("unknown fc to handle data")
		m.WriteHeader(rfc.StatusNotFound)
	}
}

func HandleCycle(handler core.Handler, m core.Meta) {
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

func HPoolCycle(hp core.HandlerPool, m core.Meta) {
	h := hp.Acquire()
	HandleCycle(h, m)
	hp.Release(h)
}
