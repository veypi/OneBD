package core

import "github.com/rs/zerolog"

type AppInfo interface {
	Logger() *zerolog.Logger
	Config() *Config
}
