package core

import "go.uber.org/zap"

type AppInfo interface {
	Logger() *zap.Logger
}
