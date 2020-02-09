package core

import (
	"crypto/tls"
	"github.com/lightjiang/utils/log"
)

type Config struct {
	// 服务监听地址
	Host  string
	Debug bool
	// log file path
	LoggerPath    string
	LoggerLevel   log.Level
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
	log.SetLevel(c.LoggerLevel)
	if c.LoggerPath != "" {
		log.SetLogger(log.FileLogger(c.LoggerPath))
	} else {
		log.SetLogger(log.ConsoleLogger())
	}
}

func DefaultConfig() *Config {
	c := &Config{}
	return c.IsValid()
}
