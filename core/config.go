package core

import (
	"crypto/tls"
	"github.com/lightjiang/OneBD/utils/log"
	"github.com/rs/zerolog"
)

type Config struct {
	// 服务监听地址
	Host          string
	Debug         bool
	LoggerPath    string
	LoggerLevel   zerolog.Level
	Logger        *zerolog.Logger
	Charset       string `json:"charset,omitempty"`
	TimeFormat    string `json:"time_format,omitempty"`
	PostMaxMemory int64
	TlsCfg        *tls.Config
	// 最大连接数量 为非正数 则不限制
	MaxConnections int
	Router         Router
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
		c.Logger = log.FileLogger(c.LoggerPath)
	} else {
		c.Logger = log.DefaultLogger
	}
}

func DefaultConfig() *Config {
	c := &Config{}
	return c.IsValid()
}
