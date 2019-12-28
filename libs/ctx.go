package libs

import (
	"github.com/lightjiang/OneBD/core"
	"net/http"
)

type context struct {
	writer  http.ResponseWriter
	request *http.Request
	app     *Application
}

func NewContext(app *Application) core.Context {
	return &context{app: app}
}

func (ctx *context) Start(w http.ResponseWriter, r *http.Request) {
	ctx.writer = w
	ctx.request = r
}
func (ctx *context) Stop() {
}

func (ctx *context) ResetCtx() {
	ctx.writer = nil
	ctx.request = nil
}
