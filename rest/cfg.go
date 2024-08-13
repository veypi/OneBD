//
// cfg.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-12 12:08
// Distributed under terms of the MIT license.
//

package rest

import (
	"crypto/tls"
	"errors"
	"fmt"
	"regexp"
)

var ipv4Regex = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$`)

type RestConf struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	// log file path
	LoggerPath     string `json:"logger_path,omitempty"`
	LoggerLevel    string `json:"logger_level,omitempty"`
	PrettyLog      bool   `json:"pretty_log,omitempty"`
	TimeFormat     string `json:"time_format,omitempty"`
	PostMaxMemory  uint
	TlsCfg         *tls.Config
	MaxConnections int
}

func (c *RestConf) Url() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *RestConf) IsValid() error {
	if !ipv4Regex.MatchString(c.Host) {
		return errors.New("invalid host")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port")
	}
	return nil
}
