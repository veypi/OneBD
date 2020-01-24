package router

import (
	"fmt"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/hpool"
	"github.com/lightjiang/OneBD/libs/meta"
	"github.com/lightjiang/OneBD/libs/oerr"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"github.com/lightjiang/OneBD/utils/log"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var app core.AppInfo
var cfg *core.Config
var logger = log.DefaultLogger
var baseRouter *route

// 触发缓存 避免每次请求都递归查询触发函数
var statusFuncCache = make(map[rfc.Status]bool)

var getRouter, addRouter = func() (func(string) *route, func(*route)) {
	var allRouter = make(map[string]*route)
	mu := sync.RWMutex{}
	return func(fullPath string) *route {
			mu.RLock()
			res, _ := allRouter[fullPath]
			mu.RUnlock()
			return res
		}, func(r *route) {
			mu.Lock()
			tmp, _ := allRouter[r.fullPath]
			if tmp != nil {
				panic(oerr.UrlDefinedDuplicate.AttachStr(r.fullPath))
			}
			allRouter[r.fullPath] = r
			mu.Unlock()
			r.parsePrefixToFragment()
			logger.Debug("add router " + r.fullPath)
		}
}()

// url 验证规则
var urlPattern = regexp.MustCompile("^/[A-Za-z0-9/:]*$")
var urlParamPattern = regexp.MustCompile("([a-z]*):([A-Za-z0-9]*)")
var urlParamType = []string{"int", "str", "float"}

// parsePrefix 校验url合法性及初始化url参数
func parsePrefix(urlParams []string, prefix string) ([]string, string) {
	prefix = strings.Trim(prefix, "/")
	prefix = "/" + prefix
	if !urlPattern.MatchString(prefix) {
		logger.Fatal(oerr.UrlPatternNotSupport.AttachStr("'"+prefix+"' is not supported", utils.CallPath(1)).Error())
	}
	params := urlParamPattern.FindAllStringSubmatch(prefix, -1)
	if params != nil {
		if urlParams == nil {
			urlParams = make([]string, 0, 5)
		} else {
			urlParams = urlParams[:]
		}
		var vType, variable string
		for _, param := range params {
			vType = param[1]
			variable = param[2]
			if ok, _ := utils.ListContainStr(urlParams, variable); ok {
				logger.Fatal(oerr.UrlParamDuplicate.AttachStr("duplicate with " + variable).Error())
			} else {
				urlParams = append(urlParams, variable)
			}
			if ok, _ := utils.ListContainStr(urlParamType, vType); !ok {
				logger.Fatal(oerr.UrlPatternNotSupport.AttachStr(strings.Join(urlParamType, ", ") + " is supported, except " + vType + " of " + prefix).Error())
			}
			prefix = strings.Replace(prefix, param[0], "@"+vType, 1)
		}
	}
	if strings.Contains(prefix, ":") || strings.Contains(prefix, "//") {
		logger.Fatal(oerr.UrlPatternNotSupport.AttachStr(prefix).Error())
	}
	return urlParams, prefix
}

type urlFragment struct {
	IsComplex bool
	Type      string
	Var       string
}

func joinFragment(p1, p2 string) string {
	p1 = strings.TrimSuffix(p1, "/")
	p2 = strings.TrimPrefix(p2, "/")
	return p1 + "/" + p2
}

type route struct {
	utils.FastLocker
	fullPath       string
	prefix         string // asd/{id:type}/
	prefixRegexp   *regexp.Regexp
	piece          []urlFragment
	root           *route
	subRouters     []*route
	allowedMethods map[rfc.Method]bool
	caller         string
	mainHandler    core.HandlerPool
	otherHandlers  map[rfc.Status]core.MetaFunc
	urlParams      []string
}

func (r *route) String() string {
	return "\n" + strings.Join(r.string(), "\n")
}

func (r *route) string() []string {
	res := make([]string, 0, 10)
	prefix := r.prefix
	if r.urlParams != nil && len(r.urlParams) > 0 {
		prefix += "    args:" + strings.Join(r.urlParams, ",")
	}
	if r.mainHandler != nil {
		if r.allowedMethods != nil && len(r.allowedMethods) > 0 {
			prefix = prefix + "    ( "
			for i := range r.allowedMethods {
				prefix = prefix + i + " "
			}
			prefix += ")"
		}
	}
	prefix += "    " + r.caller
	res = append(res, prefix)
	for _, item := range r.subRouters {
		subs := item.string()
		for _, sub := range subs {
			res = append(res, "    "+sub)
		}
	}
	return res
}

// NewRouter allowedMethods 为空时默认所有方法皆允许
func NewMainRouter(_app core.AppInfo) core.Router {
	if baseRouter == nil {
		baseRouter = &route{
			prefix:   "/",
			fullPath: "/",
		}
		addRouter(baseRouter)
	}
	app = _app
	cfg = _app.Config()
	logger = app.Logger()
	return baseRouter
}

func (r *route) Set(prefix string, fc func() core.Handler, allowedMethods ...rfc.Method) {
	var urlParams []string
	if (prefix == "/" || prefix == "") && r.root != nil {
		urlParams, prefix = parsePrefix(r.root.urlParams, prefix)
	} else {
		urlParams, prefix = parsePrefix(r.urlParams, prefix)
	}
	var tmp *route
	if prefix == "/" {
		tmp = r
		tmp.caller = utils.CallPath(1)
	} else {
		fullPath := joinFragment(r.fullPath, prefix)
		tmp = getRouter(fullPath)
		if tmp == nil {
			tmp = &route{
				prefix:    prefix,
				root:      r,
				fullPath:  fullPath,
				urlParams: urlParams,
				caller:    utils.CallPath(1),
			}
			addRouter(tmp)
			r.subRouters = append(r.subRouters, tmp)
		}
	}
	if tmp.mainHandler != nil {
		// will raise panic
		logger.Fatal(oerr.UrlDefinedDuplicate.AttachStr(tmp.fullPath + " has defined at " + tmp.caller).Error())
		return
	}
	tmp.mainHandler = hpool.NewHandlerPool(fc, app)
	if len(allowedMethods) == 0 {
		allowedMethods = []rfc.Method{rfc.MethodAll}
	}
	tmp.enableMethod(allowedMethods...)
	//r.subRouters = append(r.subRouters, tmp)
}

func (r *route) SubRouter(prefix string) core.Router {
	var urlParams []string
	urlParams, prefix = parsePrefix(r.urlParams, prefix)
	fullPath := joinFragment(r.fullPath, prefix)
	tmp := getRouter(fullPath)
	if tmp == nil {
		tmp = &route{
			prefix:    prefix,
			root:      r,
			urlParams: urlParams,
			fullPath:  fullPath,
			caller:    utils.CallPath(1),
		}
		addRouter(tmp)
		r.subRouters = append(r.subRouters, tmp)
	}
	return tmp
}

func (r *route) parsePrefixToFragment() {
	r.piece = make([]urlFragment, 0, 8)
	piece := strings.Split(r.prefix, "/")
	index := 0
	for _, item := range piece[1:] {
		var fragment = urlFragment{
			IsComplex: false,
			Type:      "",
			Var:       "",
		}
		if strings.HasPrefix(item, "@") {
			fragment.IsComplex = true
			fragment.Type = strings.TrimPrefix(item, "@")
			if ok, _ := utils.ListContainStr(urlParamType, fragment.Type); !ok {
				logger.Fatal(oerr.UrlPatternNotSupport.AttachStr(strings.Join(urlParamType, ", ") + " is supported, except " + fragment.Type + " of " + r.prefix).Error())
			}
			if r.root != nil && r.root.urlParams != nil {
				fragment.Var = r.urlParams[index+len(r.root.urlParams)]
			} else {
				fragment.Var = r.urlParams[index]
			}
			index++
		} else {
			fragment.Var = item
		}
		r.piece = append(r.piece, fragment)
	}
}

func (r *route) Match(url string) (*route, map[string]interface{}) {
	// TODO:: 是否有必要对路由查询结果做缓存
	pieces := strings.Split(url, "/")
	if url == "/" {
		pieces = pieces[1:]
	}
	params := map[string]interface{}{}
	return r.match(pieces, params), params
}

func (r *route) match(pieces []string, params map[string]interface{}) *route {
	// TODO:: params 被污染
	var index int = 0
	var v urlFragment
	for index, v = range r.piece {
		if v.IsComplex {
			switch v.Type {
			case "":
				params[v.Var] = pieces[index]
			case "str":
				params[v.Var] = pieces[index]
			case "int":
				if t, err := strconv.Atoi(pieces[index]); err != nil {
					return nil
				} else {
					params[v.Var] = t
				}
			case "float":
				if t, err := strconv.ParseFloat(pieces[index], 64); err != nil {
					return nil
				} else {
					params[v.Var] = t
				}
			default:
				return nil
			}
		} else {
			if pieces[index] != v.Var {
				return nil
			}
		}
	}
	if index == len(pieces)-1 {
		return r
	}
	var tmp *route
	pieces = pieces[index+1:]
	for _, item := range r.subRouters {
		tmp = item.match(pieces, params)
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
	//logger.Debug(req.RequestURI + " <- " + req.RemoteAddr)
	now := time.Now()
	urlPath := req.URL.Path
	if cfg.DisableRouterPathCaseSensitive {
		urlPath = strings.ToLower(urlPath)
	}
	if !cfg.DisableRouterPathCorrection && len(urlPath) > 1 {
		urlPath = strings.TrimSuffix(urlPath, "/")
	}
	req.URL.Path = urlPath
	hp, params := r.Match(urlPath)
	fmt.Printf("%p %+v\n", hp, params)
	m := meta.Acquire(w, req, params, app)
	if hp != nil {
		handler := hp.mainHandler.Acquire()
		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic error", zap.Any("err", err))
					m.SetStatus(rfc.StatusInternalServerError)
				}
			}()
			requestCircle(handler, m)
		}()
		hp.mainHandler.Release(handler)
		r.fireOnStatus(m.Status(), m)

	} else {
		m.SetStatus(rfc.StatusNotFound)
		r.fireOnStatus(rfc.StatusNotFound, m)
	}
	m.Flush()
	meta.Release(m)
	logger.Debug(req.Method+":"+req.URL.Path,
		zap.Int64("delta/us", time.Now().Sub(now).Microseconds()),
		zap.String("addr", req.RemoteAddr))
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
