// router.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-07 13:45
// Distributed under terms of the MIT license.
package rest

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/veypi/utils"
	"github.com/veypi/utils/logv"
)

type func00 = func(*X)
type func01 = func(*X) any
type func02 = func(*X) error
type func03 = func(*X) (any, error)
type func10 = func(*X, any)
type func11 = func(*X, any) any
type func12 = func(*X, any) error
type func13 = func(*X, any) (any, error)

type func20 = func(http.ResponseWriter, *http.Request)
type func21 = func(http.ResponseWriter, *http.Request) any
type func22 = func(http.ResponseWriter, *http.Request) error
type func23 = func(http.ResponseWriter, *http.Request) (any, error)
type funcErr = func(*X, error) error

var allowedMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodConnect,
	http.MethodOptions, http.MethodTrace, "ANY"}

func NewRouter() Router {
	r := &route{
		funcBefore: make([]any, 0, 10),
		funcAfter:  make([]any, 0, 10),
	}
	return r
}

type Router interface {
	String() string
	Print()
	GetParamsList() []string
	ServeHTTP(http.ResponseWriter, *http.Request)
	SubRouter(prefix string) Router

	Clear(url string, method string)
	Set(url string, method string, handlers ...any) Router
	Get(url string, handlers ...any) Router
	Any(url string, handlers ...any) Router
	Post(url string, handlers ...any) Router
	Head(url string, handlers ...any) Router
	Put(url string, handlers ...any) Router
	Patch(url string, handlers ...any) Router
	Delete(url string, handlers ...any) Router

	UseBefore(middleware ...any) Router
	UseAfter(middleware ...any) Router
	Replace(Router) Router
	Extend(string, Router) Router
}

type route struct {
	// just blank for root router
	fragment      string
	funcBefore    []any
	funcAfter     []any
	handlers      map[string][]any
	handlersCache map[string][]any

	parent *route

	subRouters map[string]*route
	colon      *route
	wildcard   *route
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
			for _, h := range r.handlersCache[m] {
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

func (r *route) match(u string, m string, x *X) (*route, []any) {
	if u == "/" || u == "" {
		if len(r.handlers[m]) > 0 {
			return r, r.handlersCache[m]
		} else if len(r.handlers["ANY"]) > 0 {
			return r, r.handlersCache["ANY"]
		}
		if r.wildcard != nil {
			if len(r.wildcard.handlers[m]) > 0 {
				x.SetParam(r.wildcard.fragment[1:], "")
				return r.wildcard, r.wildcard.handlersCache[m]
			} else if len(r.wildcard.handlers["ANY"]) > 0 {
				x.SetParam(r.wildcard.fragment[1:], "")
				return r.wildcard, r.wildcard.handlersCache["ANY"]
			}
		}
		return nil, nil
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
		temp, fcs := subr.match(nexts, m, x)
		if temp != nil {
			return temp, fcs
		}
	}
	if r.colon != nil {
		temp, fcs := r.colon.match(nexts, m, x)
		if temp != nil {
			x.SetParam(r.colon.fragment[1:], u[:idx])
			return temp, fcs
		}
	}
	if r.wildcard != nil {
		if len(r.wildcard.handlers[m]) > 0 {
			x.SetParam(r.wildcard.fragment[1:], u)
			return r.wildcard, r.wildcard.handlersCache[m]
		} else if len(r.wildcard.handlers["ANY"]) > 0 {
			x.SetParam(r.wildcard.fragment[1:], u)
			return r.wildcard, r.wildcard.handlersCache["ANY"]
		}
	}
	return nil, nil
}

func (r *route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	x := acquire()
	defer release(x)
	x.Request = req
	x.writer = w
	start := time.Now()
	_ = start

	if subR, fcs := r.match(req.URL.Path[1:], req.Method, x); subR != nil && len(fcs) > 0 {
		x.fcs = fcs
		x.Next()
		logv.WithNoCaller.Debug().Int("ms", int(time.Since(start).Milliseconds())).Str("method", req.Method).Int("code", x.code).Msg(req.RequestURI)
	} else {
		logv.WithNoCaller.Warn().Str("method", req.Method).Str("path", req.URL.Path).Msg("Not Handled")
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

func (r *route) Clear(prefix string, method string) {
	var tmp *route
	if len(r.fragment) > 0 && r.fragment[0] == '*' {
		tmp = r
	} else {
		tmp = r.get_subrouter(prefix)
	}
	if method == "*" {
		tmp.handlers = nil
		tmp.subRouters = nil
		tmp.funcAfter = nil
		tmp.funcBefore = nil
	} else {
		delete(tmp.handlers, method)
	}
	tmp.syncCache()
}

func (r *route) Set(prefix string, method string, handlers ...any) Router {
	method = strings.ToUpper(method)
	logv.Assert(utils.InList(method, allowedMethods), fmt.Sprintf("not support HTTP method: %v", method))
	// logv.Assert(len(handlers) > 0, "there must be at least one handler")

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
		if fc == nil || reflect.ValueOf(fc).IsNil() {
			logv.WithNoCaller.Fatal().Caller(1).Msgf("set nil handler for %s/%s: %T", r.String(), prefix, fc)
		}
		switch fc := fc.(type) {
		case func00, func01, func02, func03, func10, func11, func12, func13, func20, func21, func22, func23, funcErr:
		default:
			logv.WithNoCaller.Fatal().Caller(2).Msgf("handler type not support: %T", fc)
		}
	}
	if tmp.handlers[method] != nil {
		tmp.handlers[method] = append(tmp.handlers[method], handlers...)
	} else {
		tmp.handlers[method] = handlers
	}
	tmp.syncCache()
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

func (r *route) UseAfter(middleware ...any) Router {
	for _, m := range middleware {
		switch m := m.(type) {
		case func00, func01, func02, func03, func10, func11, func12, func13, func20, func21, func22, func23, funcErr:
			r.use(m, false)
		default:
			panic(fmt.Sprintf("not support middleware %T", m))
		}
	}
	return r
}

func (r *route) UseBefore(middleware ...any) Router {
	for _, m := range middleware {
		switch m := m.(type) {
		case func00, func01, func02, func03, func10, func11, func12, func13, func20, func21, func22, func23, funcErr:
			r.use(m, true)
		default:
			panic(fmt.Sprintf("not support middleware %T", m))
		}
	}
	return r
}

func (r *route) use(m any, before bool) {
	if before {
		r.funcBefore = append(r.funcBefore, m)
	} else {
		r.funcAfter = append(r.funcAfter, m)
	}
	r.syncCache()
}

func (r *route) syncCache() {
	r.handlersCache = make(map[string][]any)
	before := make([]any, 0, 10)
	after := make([]any, 0, 10)
	tmpr := r
	for tmpr != nil {
		// ! slice 陷阱
		// before = append(tmpr.funcBefore[:], before...)
		before = append(before[:0], append(tmpr.funcBefore, before...))
		after = append(after, tmpr.funcAfter...)
		tmpr = tmpr.parent
	}
	for k := range r.handlers {
		r.handlersCache[k] = append(append([]any{}, before...), r.handlers[k]...)
		r.handlersCache[k] = append(r.handlersCache[k], after...)
	}

	for _, sub := range r.subRouters {
		sub.syncCache()
	}
	if r.colon != nil {
		r.colon.syncCache()
	}
	if r.wildcard != nil {
		r.wildcard.syncCache()
	}
}

func (r *route) Extend(prefix string, subr Router) Router {
	return r.get_subrouter(prefix).Replace(subr)
}
func (r *route) Replace(subr Router) Router {
	// r.parent = parent.(*route)
	logv.Assert(r.parent != nil, "root router can not replace")
	name := r.fragment
	sub := subr.(*route)
	sub.fragment = name
	sub.parent = r.parent
	if name[0] == '*' {
		r.parent.wildcard = sub
	} else if name[0] == ':' {
		r.parent.colon = sub
	} else {
		r.parent.subRouters[name] = sub
	}
	sub.syncCache()
	return sub
}

func (r *route) SubRouter(prefix string) Router {
	logv.Assert(prefix != "" && prefix != "/", "subrouter path can not be '' or '/'")
	return r.get_subrouter(prefix)
}
