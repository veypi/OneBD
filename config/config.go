package config

import (
	"crypto/tls"
	"github.com/lightjiang/OneBD/core"
)

type Config struct {
	// 服务监听地址
	Host          string
	Charset       string `json:"charset,omitempty"`
	TimeFormat    string `json:"time_format,omitempty"`
	PostMaxMemory int64
	TlsCfg        *tls.Config
	// 最大连接数量 为0 则不限制
	MaxConnections int
	NewCtx         func() core.Context
	CtxPool        core.CtxPool
	Router         core.Router
}

func (c *Config) IsValid() *Config {
	if c.Host == "" {
		c.Host = "0.0.0.0:8000"
	}
	return c
}

func DefaultConfig() *Config {
	c := &Config{}
	return c.IsValid()
}
