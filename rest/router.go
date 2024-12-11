// router.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-07 13:45
// Distributed under terms of the MIT license.
package rest

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

type fc0 = func(*X)
type fc1 = func(*X) error
type fc2 = func(http.ResponseWriter, *http.Request) error
type fc3 = func(http.ResponseWriter, *http.Request)

type fc4 = func(*X) (any, error)
type fc5 = func(*X, any) error
type fc6 = func(*X, any) (any, error)

type ApiHandler interface {
	fc0 | fc1 | fc2 | fc3 | fc4 | fc5 | fc6
}

type ErrHandle = func(x *X, err error)

var allowedMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodConnect,
	http.MethodOptions, http.MethodTrace, "ANY"}

func NewRouter() Router {
	r := &route{}
	r.errHandler = func(x *X, err error) {
		if errors.Is(err, ErrNotFound) {
			x.WriteHeader(404)
		} else {
			x.WriteHeader(500)
		}
		if err != nil {
			x.Write([]byte(err.Error()))
		}
	}
	return r
}

type Router interface {
	String() string
	Print()
	GetParamsList() []string
	ServeHTTP(http.ResponseWriter, *http.Request)
	SubRouter(prefix string) Router

	Set(url string, method string, handlers ...any) Router
	Get(url string, handlers ...any) Router
	Any(url string, handlers ...any) Router
	Post(url string, handlers ...any) Router
	Head(url string, handlers ...any) Router
	Put(url string, handlers ...any) Router
	Patch(url string, handlers ...any) Router
	Delete(url string, handlers ...any) Router

	Use(middleware ...any)
	SetErrFunc(fc ErrHandle)
}

type route struct {
	// just blank for root router
	fragment   string
	handlers   map[string][]any
	middleware []any

	root   *route
	parent *route

	subRouters map[string]*route
	colon      *route
	wildcard   *route

	errHandler ErrHandle
}

func (r *route) Print() {
	fmt.Printf("Router Table\n%s\n", strings.Join(r.tree(), "\n"))
}

func (r *route) tree() []string {
	fc := func(res []string, subt *route) []string {
		if subt != nil {
			for _, s := range subt.tree() {
				res = append(res, s)
			}
		}
		return res
	}
	res := make([]string, 0, 10)
	if r.handlers != nil && len(r.handlers) > 0 {
		item := r.String()
		if item == "" {
			item = "/"
		}
		item = "\033[32m" + item + "\033[0m"
		for m := range r.handlers {
			item += "\n    " + m
			for _, h := range r.handlers[m] {
				op := reflect.ValueOf(h).Pointer()
				fnName := strings.Split(runtime.FuncForPC(op).Name(), "/")
				item += fmt.Sprintf(" %s", fnName[len(fnName)-1])
			}
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

func (r *route) GetParamsList() []string {
	var res []string
	tr := r
	for tr != nil {
		if strings.HasPrefix(tr.fragment, ":") || strings.HasPrefix(tr.fragment, "*") {
			res = append(res, tr.fragment)
		}
		tr = tr.parent
	}
	return res
}

func (r *route) String() string {
	if r.parent != nil {
		return r.parent.String() + "/" + r.fragment
	}
	return r.fragment
}

func (r *route) match(u string, m string, x *X) *route {
	if u == "/" || u == "" {
		if len(r.handlers[m]) > 0 || len(r.handlers["ANY"]) > 0 {
			return r
		}
		if r.wildcard != nil {
			if len(r.wildcard.handlers[m]) > 0 || len(r.wildcard.handlers["ANY"]) > 0 {
				x.SetParam(r.wildcard.fragment[1:], "")
				return r.wildcard
			}
		}
		return nil
	}
	idx := 0
	for i, v := range u {
		if v == '/' {
			break
		} else {
			idx = i + 1
		}
	}
	nexts := u[idx:]
	if len(nexts) > 0 && nexts[0] == '/' {
		nexts = nexts[1:]
	}
	if subr := r.subRouters[u[:idx]]; subr != nil {
		temp := subr.match(nexts, m, x)
		if temp != nil {
			return temp
		}
	}
	if r.colon != nil {
		temp := r.colon.match(nexts, m, x)
		if temp != nil {
			x.SetParam(r.colon.fragment[1:], u[:idx])
			return temp
		}
	}
	if r.wildcard != nil {
		if len(r.wildcard.handlers[m]) > 0 || len(r.wildcard.handlers["ANY"]) > 0 {
			x.SetParam(r.wildcard.fragment[1:], u)
			return r.wildcard
		}
	}
	return nil
}

func (r *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	x := acquire()
	defer release(x)
	x.Request = req
	x.writer = w
	start := time.Now()

	if subR := r.match(req.URL.Path[1:], req.Method, x); subR != nil {
		x.fcs = subR.handlers[req.Method]
		if len(x.fcs) == 0 {
			x.fcs = subR.handlers["ANY"]
		}
		err := x.Next()
		if err != nil {
			subR.fire_err(x, err)
		}
	} else {
		r.fire_err(x, ErrNotFound)
	}
	logv.WithNoCaller.Debug().Int("ms", int(time.Since(start).Milliseconds())).Str("method", req.Method).Int("code", x.code).Msg(req.RequestURI)
}

func (r *route) fire_err(x *X, err error) {
	if r.errHandler != nil {
		r.errHandler(x, err)
	} else {
		r.parent.fire_err(x, err)
	}
}

func (r *route) get_subrouter(url string) *route {
	if url == "" || url == "/" {
		return r
	}
	startIdx := 0
	if url[0] == '/' {
		url = url[1:]
	}
	if url[len(url)-1] != '/' {
		url += "/"
	}
	var next *route
	last := r
	for i, c := range url {
		if c == '/' {
			next = &route{
				fragment: url[startIdx:i],
				parent:   last,
				root:     last.root,
			}
			startIdx = i + 1
			if next.fragment == "" {
				logv.Assert(false, "url path can not has //")
			} else if next.fragment[0] == '*' {
				if last.wildcard != nil {
					if last.wildcard.fragment != next.fragment {
						logv.Warn().Msgf("variable path conflict: %s %s", last.colon.String(), next.String())
					}
					return last.wildcard
				}
				last.wildcard = next
				return next
			} else if next.fragment[0] == ':' {
				if last.colon != nil {
					if last.colon.fragment != next.fragment {
						logv.Warn().Msgf("variable path conflict: %s %s", last.colon.String(), next.String())
					}
					last = last.colon
				} else {
					last.colon = next
					last = next
				}
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
}

func (r *route) Set(prefix string, method string, handlers ...any) Router {
	method = strings.ToUpper(method)
	logv.Assert(utils.InList(method, allowedMethods), fmt.Sprintf("not support HTTP method: %v", method))
	logv.Assert(len(handlers) > 0, "there must be at least one handler")

	var tmp *route
	if len(r.fragment) > 0 && r.fragment[0] == '*' {
		tmp = r
	} else {
		tmp = r.get_subrouter(prefix)
	}
	if tmp.handlers == nil {
		tmp.handlers = make(map[string][]any)
	}
	for _, fc := range handlers {
		if reflect.ValueOf(fc).IsNil() {
			logv.WithNoCaller.Fatal().Caller(1).Msgf("set nil handler for %s/%s: %T", r.String(), prefix, fc)
		}
		switch fc := fc.(type) {
		case fc0, fc1, fc2, fc3, fc4, fc5, fc6:
		default:
			logv.WithNoCaller.Fatal().Caller(1).Msgf("handler type not support: %T", fc)
		}
	}
	if tmp.handlers[method] != nil {
		tmp.handlers[method] = append(tmp.handlers[method], handlers...)
		return tmp
	}
	fcs := make([]any, 0, 10)
	var tmp_route = r
	for {
		if len(tmp_route.middleware) != 0 {
			fcs = append(tmp_route.middleware, fcs...)
		}
		if tmp_route.parent != nil {
			tmp_route = tmp_route.parent
		} else {
			break
		}
	}
	fcs = append(fcs, handlers...)
	tmp.handlers[method] = fcs
	return tmp
}
func (r *route) Any(url string, handlers ...any) Router {
	return r.Set(url, "ANY", handlers...)
}
func (r *route) Get(url string, handlers ...any) Router {
	return r.Set(url, http.MethodGet, handlers...)
}
func (r *route) Post(url string, handlers ...any) Router {
	return r.Set(url, http.MethodPost, handlers...)
}
func (r *route) Head(url string, handlers ...any) Router {
	return r.Set(url, http.MethodHead, handlers...)
}
func (r *route) Put(url string, handlers ...any) Router {
	return r.Set(url, http.MethodPut, handlers...)
}
func (r *route) Patch(url string, handlers ...any) Router {
	return r.Set(url, http.MethodPatch, handlers...)
}
func (r *route) Delete(url string, handlers ...any) Router {
	return r.Set(url, http.MethodDelete, handlers...)
}

func (r *route) Use(middleware ...any) {
	for _, m := range middleware {
		switch m := m.(type) {
		case fc0, fc1, fc2, fc3, fc4, fc5, fc6:
			r.use(m)
		default:
			panic(fmt.Sprintf("not support middleware %T", m))
		}
	}
}

func (r *route) use(m any) {
	if r == nil {
		return
	}
	r.middleware = append(r.middleware, m)
	for _, sub := range r.subRouters {
		sub.use(m)
	}
	r.colon.use(m)
	r.wildcard.use(m)
	for method := range r.handlers {
		r.handlers[method] = append(r.handlers[method], m)
	}
}

func (r *route) SubRouter(prefix string) Router {
	logv.Assert(prefix != "" && prefix != "/", "subrouter path can not be '' or '/'")
	return r.get_subrouter(prefix)
}

func (r *route) SetErrFunc(fc ErrHandle) {
	r.errHandler = fc
}
