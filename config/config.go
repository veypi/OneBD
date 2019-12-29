package config

import (
	"crypto/tls"
	"github.com/lightjiang/OneBD/core"
	"github.com/lightjiang/OneBD/utils/log"
	"go.uber.org/zap"
)

type Config struct {
	// 服务监听地址
	Host          string
	Debug         bool
	LoggerPath    string
	LoggerLevel   log.Level
	Logger        *zap.Logger
	Charset       string `json:"charset,omitempty"`
	TimeFormat    string `json:"time_format,omitempty"`
	PostMaxMemory int64
	TlsCfg        *tls.Config
	// 最大连接数量 为非正数 则不限制
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

func (c *Config) BuildLogger() {
	if c.Logger != nil {
		return
	}
	log.SetLevel(c.LoggerLevel)
	if c.LoggerPath != "" {
		log.EnableFileLog(c.LoggerPath)
	}
	c.Logger = log.Build()
}

func DefaultConfig() *Config {
	c := &Config{}
	return c.IsValid()
}
