//
// router.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-07 13:45
// Distributed under terms of the MIT license.
//
package router

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/veypi/utils"
	"github.com/veypi/utils/logx"
)

type fc0 = func(*X) error
type fc1 = func(http.ResponseWriter, *http.Request) error
type fc2 = func(http.ResponseWriter, *http.Request)

type ErrHandle = func(http.ResponseWriter, *http.Request, error, int)

var (
	ErrNotFound = errors.New("not found")
)

var allowedMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodConnect,
	http.MethodOptions, http.MethodTrace}

func NewRouter() Router {
	r := &route{}
	r.errHandler = func(w http.ResponseWriter, r *http.Request, err error, code int) {
		w.WriteHeader(code)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	}
	return r
}

type Router interface {
	String() string
	Print()
	ServeHTTP(http.ResponseWriter, *http.Request)
	SubRouter(prefix string) Router

	Set(url string, method string, handlers ...any) Router
	Get(url string, handlers ...any) Router
	Post(url string, handlers ...any) Router
	Head(url string, handlers ...any) Router
	Put(url string, handlers ...any) Router
	Patch(url string, handlers ...any) Router
	Delete(url string, handlers ...any) Router

	Use(middleware ...any)
	SetErrFunc(fc ErrHandle)
	Static(prefix string, directory string)
	EmbedFile(prefix string, f []byte)
	EmbedDir(prefix string, fs embed.FS, fsPrefix string)
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
			item += " " + m
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

func (r *route) String() string {
	if r.parent != nil {
		return r.parent.String() + "/" + r.fragment
	}
	return r.fragment
}

func (r *route) match(u string, m string, x *X) *route {
	if u == "/" || u == "" {
		if len(r.handlers[m]) > 0 {
			return r
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
		if len(r.wildcard.handlers[m]) > 0 {
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
	x.ResponseWriter = w
	start := time.Now()

	if subR := r.match(req.URL.Path[1:], req.Method, x); subR != nil {
		x.fcs = subR.handlers[req.Method]
		err := x.Next()
		if err != nil {
			subR.fire_err(w, req, err, 500)
		}
	} else {
		subR.fire_err(w, req, ErrNotFound, 404)
	}
	logx.WithNoCaller.Debug().Int("ms", int(time.Since(start).Milliseconds())).Str("method", req.Method).Msg(req.RequestURI)
}

func (r *route) fire_err(w http.ResponseWriter, req *http.Request, err error, code int) {
	logx.Info().Msgf("%s: %v", r, err)
	if r.errHandler != nil {
		r.errHandler(w, req, err, code)
	} else {
		r.parent.fire_err(w, req, err, code)
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
				logx.Assert(false, "url path can not has //")
			} else if next.fragment[0] == '*' {
				if last.wildcard != nil {
					return r.wildcard
				}
				last.wildcard = next
				return next
			} else if next.fragment[0] == ':' {
				if last.colon != nil {
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
	logx.Assert(prefix != "", "path must begin with '/'")
	logx.Assert(utils.InList(method, allowedMethods), fmt.Sprintf("not support HTTP method: %v", method))
	logx.Assert(len(handlers) > 0, "there must be at least one handler")

	var tmp *route
	if len(r.fragment) > 0 && r.fragment[0] == '*' {
		tmp = r
	} else {
		tmp = r.get_subrouter(prefix)
	}
	if tmp.handlers == nil {
		tmp.handlers = make(map[string][]any)
	}
	logx.Assert(tmp.handlers[method] == nil, "url defined duplicate")
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
	for _, fc := range handlers {
		switch fc := fc.(type) {
		case fc0:
		case fc1:
		case fc2:
		default:
			logx.Fatal().Msgf("handler type not support: %T", fc)
		}
	}
	fcs = append(fcs, handlers...)
	tmp.handlers[method] = fcs
	return tmp
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
		case fc0, fc1, fc2:
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

func (r *route) Static(prefix string, directory string) {
	dir, err := os.Stat(directory)
	if err != nil {
		panic(err)
	}
	if !dir.IsDir() {
		r.Set(prefix, http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
			f, err := os.Open(directory)
			if err != nil {
				panic(err)
			}
			info, err := f.Stat()
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(info.Name())))
			http.ServeContent(w, req, info.Name(), info.ModTime(), f)
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
	r.Set(prefix, http.MethodGet, func(w http.ResponseWriter, req *http.Request, x *X) {
		name := x.Params.GetStr("path")
		f, err := fs.Open(name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if info.IsDir() {
			// TODO:: dir list
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(info.Name())))
		http.ServeContent(w, req, info.Name(), info.ModTime(), f)
	}, http.MethodGet)
}

func (r *route) EmbedFile(prefix string, f []byte) {
	r.Set(prefix, http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(req.URL.Path)))
		_, err := w.Write(f)
		if err != nil {
			logx.Warn().Msgf("write file failed: %s", err.Error())
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
	r.Set(prefix, http.MethodGet, func(w http.ResponseWriter, req *http.Request, x *X) {
		name := fsPrefix + x.Params.GetStr("path")
		f, err := fs.Open(name)
		if err != nil {
			logx.Info().Msgf("serve file failed: %s", err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			logx.Info().Msgf("serve file failed: %s", err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if info.IsDir() {
			// TODO:: dir list
			logx.Info().Msgf("serve file failed: %s", err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(info.Name())))
		http.ServeContent(w, req, info.Name(), info.ModTime(), f.(io.ReadSeeker))
	}, http.MethodGet)
}

func (r *route) SubRouter(prefix string) Router {
	logx.Assert(prefix != "" && prefix != "/", "subrouter path can not be '' or '/'")
	return r.get_subrouter(prefix)
}

func (r *route) SetErrFunc(fc ErrHandle) {
	r.errHandler = fc
}
