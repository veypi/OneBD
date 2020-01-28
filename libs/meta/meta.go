package meta

import (
	"bytes"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/OneBD/utils"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	initTime       time.Time
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
	p.initTime = time.Now()
	p.TryReset()
	p.app = app
	p.writer = w
	p.request = r
	p.status = rfc.StatusOK
	p.params = params
	p.resolvedParams = map[string]string{}
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

func (p *payLoad) Logger() *zerolog.Logger {
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
			p.paramsIndex[v-1] = i
		}
		started := false
		var startedID uint = 0
		startedIdx := 0
		tmpPath := p.RequestPath()
		for i, v := range tmpPath {
			if v == '/' {
				if started {
					started = false
					if arg, ok := p.paramsIndex[startedID]; ok {
						if arg[0] == ':' {
							p.resolvedParams[arg[1:]] = tmpPath[startedIdx : i-1]
						} else if arg[0] == '*' {
							p.resolvedParams[arg[1:]] = tmpPath[startedIdx:]
							break
						}
					}
				}
				startedID += 1
				startedIdx = i + 1
				started = true
			}
		}
		if started {
			if arg, ok := p.paramsIndex[startedID]; ok {
				p.resolvedParams[arg[1:]] = tmpPath[startedIdx:]
			}
		}
		p.app.Logger().Info().Interface("p", p.params).Interface("i", p.paramsIndex).Interface("re", p.resolvedParams).Msg(tmpPath)
	}
	p.mu.Unlock()
	return p.resolvedParams[key]
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
		p.app.Logger().Warn().Msg("try to set header failed, must be called before flush")
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
	}
}

func (p *payLoad) Write(wrt []byte) {
	p.buf.Write(wrt)
}

func (p *payLoad) ResetBuf() {
	p.buf.Reset()
}

func (p *payLoad) AliveTime() time.Duration {
	return time.Now().Sub(p.initTime)
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
