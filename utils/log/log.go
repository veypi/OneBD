package log

import (
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func SetLevel(l zerolog.Level) {
	zerolog.SetGlobalLevel(l)
}

var fileHook = lumberjack.Logger{
	Filename:   "",
	MaxSize:    128, // 每个日志文件保存的最大尺寸 单位：M
	MaxBackups: 30,  // 日志文件最多保存多少个备份
	MaxAge:     21,  // 文件最多保存多少天
	LocalTime:  true,
	Compress:   true, // 是否压缩
}

// DefaultLogger just for dev env, low performance but human-friendly
var DefaultLogger = func() *zerolog.Logger {
	l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Timestamp().Logger()
	return &l
}()

// FileLogger for product, height performance
func FileLogger(fileName string) *zerolog.Logger {
	fileHook.Filename = fileName
	l := zerolog.New(&fileHook).With().Caller().Timestamp().Logger()
	return &l
}
