package meta

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/utils"
	"github.com/lightjiang/utils/log"
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
	mu             utils.FastLocker
	initTime       time.Time
	empty          utils.SafeBool
	writer         http.ResponseWriter
	request        *http.Request
	buf            bytes.Buffer
	status         rfc.Status
	ifSetStatus    utils.SafeBool
	ifRead         utils.SafeBool
	ifFlush        utils.SafeBool
	params         map[string]uint
	paramsIndex    map[uint]string
	resolvedParams map[string]string
}

func (p *payLoad) Init(w http.ResponseWriter, r *http.Request, params map[string]uint) {
	p.initTime = time.Now()
	p.TryReset()
	p.writer = w
	p.request = r
	p.status = rfc.StatusOK
	p.params = params
	p.resolvedParams = map[string]string{}
	p.empty.ForceSetFalse()
	p.ifSetStatus.ForceSetFalse()
	p.ifRead.ForceSetFalse()
	p.ifFlush.ForceSetFalse()
}

func (p *payLoad) TryReset() {
	if p.empty.SetTrue() {
		p.writer = nil
		p.request = nil
		p.params = nil
		p.paramsIndex = nil
		p.resolvedParams = nil
		p.ResetBuf()
	}
}

func (p *payLoad) RemoteAddr() string {
	return p.request.RemoteAddr
}

func (p *payLoad) RequestPath() string {
	return p.request.URL.Path
}

func (p *payLoad) UnmarshalBody(ptr interface{}, fc func([]byte, interface{}) error) error {
	p.request.ParseForm()
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

func (p *payLoad) Header(key string) string {
	return p.request.Header.Get(key)
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

func (p *payLoad) SetStatus(status rfc.Status) {
	p.status = status
}

func (p *payLoad) Status() rfc.Status {
	return p.status
}

func (p *payLoad) SetHeader(key, value string) {
	if p.ifSetStatus.IfTrue() {
		log.Warn().Msg("try to set header failed, must be called before flush")
		return
	}
	if value == "" {
		p.writer.Header().Del(key)
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
	if p.ifRead.SetTrue() {
		io.Copy(wrt, p.request.Body)
	} else {
		log.Warn().Msg("request body has been read")
	}
}

func (p *payLoad) StreamWrite(src io.Reader) {
	if p.ifFlush.SetTrue() {
		p.flushStatus()
		io.Copy(p.writer, src)
	} else {
		log.Warn().Msg("response context has been written")
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

func Acquire(w http.ResponseWriter, r *http.Request, params map[string]uint) core.Meta {
	m := pool.Get().(core.Meta)
	m.Init(w, r, params)
	return m
}

func Release(m core.Meta) {
	m.TryReset()
	pool.Put(m)
}
