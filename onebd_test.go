package OneBD

import (
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/libs/handler"
	"github.com/lightjiang/OneBD/libs/hpool"
	"github.com/lightjiang/OneBD/rfc"
	"github.com/lightjiang/utils/log"
	"testing"
)

type testHandler struct {
	handler.Base
}

var response = []byte("response")

func (h *testHandler) Get() (interface{}, error) {
	//h.Meta().SetHeader("a", "1")
	h.Meta().Write(response)
	//h.Meta().StreamWrite(strings.NewReader(h.Meta().Params("uid")))
	return nil, nil
}

func (h *testHandler) Post() (interface{}, error) {
	var m = map[string]interface{}{}
	err := h.Meta().ReadJson(&m)
	if err != nil {
		return nil, err
	}
	log.Warn().Interface("m", m).
		Str("uid", h.Meta().Params("uid")).
		Str("abc", h.Meta().Params("abc")).
		Str("query_a", h.Meta().Query("a")).
		Str("query_b", h.Meta().Query("b")).
		Msg("msg")
	return nil, nil
}

func TestNew(t *testing.T) {
	cfg := &core.Config{
		Host:           "0.0.0.0:8080",
		Charset:        "",
		TimeFormat:     "",
		PostMaxMemory:  0,
		TlsCfg:         nil,
		MaxConnections: 0,
	}
	cfg.LoggerLevel = log.WarnLevel
	cfg.BuildLogger()
	app := New(cfg)
	newH := func() core.Handler {
		log.Info().Msg("creating a handler")
		return &testHandler{}
	}
	newPool := hpool.New(newH)
	hp := hpool.New(newH)
	app.Router().Set("/", hp, rfc.MethodGet)
	app.Router().Set("/s/:uid", hp, rfc.MethodGet)
	app.Router().SubRouter("/asd/sss").Set("/:uid/*abc", newPool, rfc.MethodGet, rfc.MethodPost)
	app.Router().SubRouter("/asd/sss").Set("asd/zzz", newH, rfc.MethodPost)
	app.Router().SubRouter("/sss/asd").Set("/123/:uid/:username", newH)
	app.Router().SetNotFoundFunc(func(m core.Meta) {
		log.Info().Msg("checking 404 status")
		m.Write([]byte(m.RequestPath()))
	})
	app.Router().SetInternalErrorFunc(func(meta core.Meta) {
		log.Info().Msg("checking 500 status")
	})
	err := app.Run()
	t.Error(err)
}
