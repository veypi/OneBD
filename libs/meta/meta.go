package meta

import (
	"bytes"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		var p core.Meta
		p = &payLoad{}
		return p
	},
}

// payLoad 请求基本处理包
type payLoad struct {
	mu             utils.FastLocker
	empty          utils.SafeBool
	app            core.AppInfo
	writer         http.ResponseWriter
	request        *http.Request
	buf            bytes.Buffer
	status         rfc.Status
	ifSetStatus    utils.SafeBool
	ifFlush        utils.SafeBool
	params         map[string]uint
	paramsIndex    map[uint]string
	resolvedParams map[string]string
}

func (p *payLoad) Init(w http.ResponseWriter, r *http.Request, params map[string]uint, app core.AppInfo) {
	p.TryReset()
	p.app = app
	p.writer = w
	p.request = r
	p.status = rfc.StatusOK
	p.params = params
	p.empty.ForceSetFalse()
	p.ifSetStatus.ForceSetFalse()
	p.ifFlush.ForceSetFalse()
}

func (p *payLoad) TryReset() {
	if p.empty.SetTrue() {
		p.writer = nil
		p.request = nil
		p.paramsIndex = nil
		p.resolvedParams = nil
		p.ResetBuf()
	}
}

func (p *payLoad) Logger() *zap.Logger {
	return p.app.Logger()
}

func (p *payLoad) RemoteAddr() string {
	return p.request.RemoteAddr
}

func (p *payLoad) RequestPath() string {
	return p.request.URL.Path
}

// url 后缀参数, 惰性解析
func (p *payLoad) Query(key string) string {
	return ""
}

// url 路径内参数, 由router 给出 惰性解析
func (p *payLoad) Params(key string) string {
	p.mu.Lock()
	if len(p.params) > 0 && len(p.paramsIndex) == 0 {
		p.paramsIndex = make(map[uint]string)
		for i, v := range p.params {
			p.paramsIndex[v] = i
		}
		started := false
		//startedIdx := 2
		for _, v := range p.RequestPath() {
			if v == '/' {
				if started {
					started = false
				} else {
					started = true
				}
			}
		}
	}
	p.mu.Unlock()
	return ""
}

func (p *payLoad) ParamsInt(key string) int {
	v, _ := strconv.Atoi(p.Params(key))
	return v
}

func (p *payLoad) Method() rfc.Method {
	return rfc.Method(p.request.Method)
}

func (p *payLoad) SetStatus(status rfc.Status) {
	p.status = status
}

func (p *payLoad) Status() rfc.Status {
	return p.status
}

func (p *payLoad) SetHeader(key, value string) {
	if p.ifSetStatus.IfTrue() {
		p.app.Logger().Warn("try to set header failed, must be called before flush")
		return
	}
	p.writer.Header().Set(key, value)
}

func (p *payLoad) flushStatus() {
	if p.ifSetStatus.SetTrue() {
		p.writer.WriteHeader(int(p.status))
	}
}

func (p *payLoad) StreamRead(wrt io.Writer) {
	io.Copy(wrt, p.request.Body)
}

func (p *payLoad) StreamWrite(src io.Reader) {
	if p.ifFlush.SetTrue() {
		p.flushStatus()
		io.Copy(p.writer, src)
	}
}

func (p *payLoad) Flush() {
	if p.ifFlush.SetTrue() {
		p.flushStatus()
		p.buf.WriteTo(p.writer)
	} else {
		p.app.Logger().Warn(p.request.URL.Path+":"+p.request.Method+" -> payload has flushed", zap.String("addr", p.RemoteAddr()))
	}
}

func (p *payLoad) Write(wrt []byte) {
	p.buf.Write(wrt)
}

func (p *payLoad) ResetBuf() {
	p.buf.Reset()
}

func Acquire(w http.ResponseWriter, r *http.Request, params map[string]uint, app core.AppInfo) core.Meta {
	m := pool.Get().(core.Meta)
	m.Init(w, r, params, app)
	return m
}

func Release(m core.Meta) {
	m.TryReset()
	pool.Put(m)
}
