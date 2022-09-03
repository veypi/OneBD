package meta

import (
	"encoding/json"
	"errors"
	"github.com/veypi/OneBD/core"
	"github.com/veypi/OneBD/rfc"
	"github.com/veypi/utils"
	"io"
	"io/ioutil"
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
	writer         http.ResponseWriter
	mu             utils.FastLocker
	disabled       bool
	initTime       time.Time
	empty          utils.SafeBool
	request        *http.Request
	status         rfc.Status
	ifRead         utils.SafeBool
	ifFlushStatus  utils.SafeBool
	params         map[string]uint
	paramsIndex    map[uint]string
	resolvedParams map[string]string
}

func (p *payLoad) Init(w http.ResponseWriter, r *http.Request, params map[string]uint) {
	p.initTime = time.Now()
	p.writer = w
	p.request = r
	p.status = rfc.StatusOK
	p.params = params
	//p.resolvedParams = make(map[string]string)
	p.empty.ForceSetFalse()
	p.ifRead.ForceSetFalse()
	p.ifFlushStatus.ForceSetFalse()
}

func (p *payLoad) TryReset() {
	p.flushStatue()
	p.disabled = false
	if p.empty.SetTrue() {
		p.writer = nil
		p.request = nil
		p.params = nil
		p.paramsIndex = nil
		p.resolvedParams = nil
	}
}

func (p *payLoad) RemoteAddr() string {
	return p.request.RemoteAddr
}

func (p *payLoad) RequestPath() string {
	return p.request.URL.Path
}

func (p *payLoad) UnmarshalBody(ptr interface{}, fc func([]byte, interface{}) error) error {
	if p.request.Body == nil {
		return errors.New("empty body")
	}
	if p.ifRead.SetTrue() {
		data, err := ioutil.ReadAll(p.request.Body)
		if err != nil {
			return err
		}
		return fc(data, ptr)
	}
	return errors.New("request body has been read")
}

func (p *payLoad) ReadJson(ptr interface{}) error {
	return p.UnmarshalBody(ptr, json.Unmarshal)
}

// url 后缀参数, 惰性解析
func (p *payLoad) Query(key string) string {
	return p.request.URL.Query().Get(key)
}

func (p *payLoad) GetHeader(key string) string {
	return p.request.Header.Get(key)
}

// url 路径内参数, 由router 给出 惰性解析
func (p *payLoad) Params(key string) string {
	p.mu.Lock()
	if len(p.params) > 0 && len(p.paramsIndex) == 0 {
		p.paramsIndex = make(map[uint]string)
		p.resolvedParams = make(map[string]string)
		for i, v := range p.params {
			p.paramsIndex[v] = i
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
							p.resolvedParams[arg[1:]] = tmpPath[startedIdx:i]
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
	}
	p.mu.Unlock()
	return p.resolvedParams[key]
}

func (p *payLoad) ParamsInt(key string) int {
	v, _ := strconv.Atoi(p.Params(key))
	return v
}

func (p *payLoad) Method() rfc.Method {
	return p.request.Method
}

func (p *payLoad) Status() rfc.Status {
	return p.status
}

func (p *payLoad) SetHeader(key, value string) {
	if value == "" {
		p.writer.Header().Del(key)
		return
	}
	p.writer.Header().Set(key, value)
}

func (p *payLoad) StreamRead(wrt io.Writer) (int64, error) {
	if p.ifRead.SetTrue() {
		return io.Copy(wrt, p.request.Body)
	}
	return 0, errors.New("request body has been read")
}

func (p *payLoad) StreamWrite(src io.Reader) (int64, error) {
	p.flushStatue()
	return io.Copy(p.writer, src)
}

func (p *payLoad) Request() *http.Request {
	return p.request
}

func (p *payLoad) ResponseWriter() http.ResponseWriter {
	return p.writer
}

func (p *payLoad) DisableSelfWriter() {
	p.disabled = true
}

func (p *payLoad) Write(wrt []byte) (int, error) {
	if p.disabled {
		return 0, nil
	}
	p.flushStatue()
	return p.writer.Write(wrt)
}

func (p *payLoad) WriteHeader(status rfc.Status) {
	p.status = status
}

func (p *payLoad) Header() http.Header {
	return p.writer.Header()
}

func (p *payLoad) flushStatue() {
	if !p.disabled && p.ifFlushStatus.SetTrue() {
		p.writer.WriteHeader(p.status)
	}
}

func (p *payLoad) AliveTime() time.Duration {
	return time.Now().Sub(p.initTime)
}

func Acquire(w http.ResponseWriter, r *http.Request, params map[string]uint) core.Meta {
	m := pool.Get().(core.Meta)
	m.Init(w, r, params)
	return m
}

func Release(m core.Meta) {
	m.TryReset()
	pool.Put(m)
}
