package libs

import (
	"github.com/lightjiang/OneBD/core"
	"net/http"
	"runtime"
	"sync/atomic"
)

type payLoad struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

type context struct {
	ctxPool  core.CtxPool
	mu       uint32
	started  uint32
	taskChan chan *payLoad
	payLoad  payLoad
	response http.ResponseWriter
	request  *http.Request
	app      *Application
}

func NewContext(app *Application) core.Context {
	c := &context{app: app, ctxPool: app.ctxPool}
	c.taskChan = make(chan *payLoad, 1)
	return c
}

func (c *context) Start(w http.ResponseWriter, r *http.Request) {
	p := &payLoad{
		Writer:  w,
		Request: r,
	}
	c.taskChan <- p
	if !atomic.CompareAndSwapUint32(&c.started, 0, 1) {
		return
	}
	// TODO: 如何检测等待超时
	// TODO: net/Serve 每次建立连接都会创建goroutine, 是否还有必要在此使用协程复用
	go func() {
		for t := range c.taskChan {
			// nil为goroutine终止条件
			if t == nil {
				atomic.CompareAndSwapUint32(&c.started, 1, 0)
				return
			}
			c.request = t.Request
			c.response = t.Writer
		}
	}()
}

func (c *context) Stop() {
	if atomic.CompareAndSwapUint32(&c.started, 1, 0) {
		c.taskChan <- nil
	}
}

func (c *context) ResetCtx() {
	c.response = nil
	c.request = nil
}

func (c *context) Lock() {
	for !atomic.CompareAndSwapUint32(&c.mu, 0, 1) {
		runtime.Gosched()
	}
}

func (c *context) Unlock() {
	atomic.StoreUint32(&c.mu, 0)
}
