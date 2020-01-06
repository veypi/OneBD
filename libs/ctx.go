package libs

import (
	ctx "context"
	"github.com/lightjiang/OneBD/core"
	"net/http"
)

type context struct {
	parent     ctx.Context
	ctx        ctx.Context
	cancelFunc ctx.CancelFunc
	writer     http.ResponseWriter
	request    *http.Request
	app        *Application
}

func NewContext(app *Application) core.Context {
	c := &context{app: app, parent: ctx.Background()}
	c.ctx, c.cancelFunc = ctx.WithCancel(c.parent)
	return c
}

func (c *context) Start(w http.ResponseWriter, r *http.Request) {
	c.writer = w
	c.request = r
}
func (c *context) Stop() {

}

func (c *context) ResetCtx() {
	c.writer = nil
	c.request = nil
}
