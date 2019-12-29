package OneBD

import (
	"github.com/lightjiang/OneBD/config"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Host:           "0.0.0.0:4000",
		Charset:        "",
		TimeFormat:     "",
		PostMaxMemory:  0,
		TlsCfg:         nil,
		MaxConnections: 0,
	}
	cfg.BuildLogger()
	app := New(cfg)
	err := app.Run()
	t.Error(err)
}
