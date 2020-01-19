package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// A Level is a logging priority. Higher levels are more important.
type Level int8

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any oerr-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel

	_minLevel = DebugLevel
	_maxLevel = FatalLevel
)

var level = zap.NewAtomicLevel()

func SetLevel(l Level) {
	level.SetLevel(zapcore.Level(l))
}

var fileHook = lumberjack.Logger{
	Filename:   "",
	MaxSize:    128, // 每个日志文件保存的最大尺寸 单位：M
	MaxBackups: 30,  // 日志文件最多保存多少个备份
	MaxAge:     21,  // 文件最多保存多少天
	LocalTime:  true,
	Compress:   true, // 是否压缩
}

func EnableFileLog(filePath string) {
	fileHook.Filename = filePath
}

func Build() *zap.Logger {
	jsonEncoder := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.RFC3339TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	consoleEncoder := jsonEncoder
	consoleEncoder.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	consoleCore := zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoder), zapcore.AddSync(os.Stdout), level)
	var core zapcore.Core
	if fileHook.Filename != "" {
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(jsonEncoder),                                                 // 编码器配置
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&fileHook)), // 打印到控制台和文件
			level, // 日志级别
		)
		core = zapcore.NewTee(fileCore, consoleCore)
	} else {
		core = consoleCore
	}

	// 开启开发模式，堆栈跟踪
	trace := zap.AddStacktrace(zapcore.ErrorLevel)
	// 开启文件及行号
	caller := zap.AddCaller()
	//development := zap.Development()
	// 构造日志
	logger := zap.New(core, caller, trace)
	return logger
}
