package logger

import (
	"fmt"
	"github.com/qkzsky/gutils/config"
	"go.uber.org/zap/buffer"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	bufferPool = buffer.NewPool()
	logPath    string
	loggerMap  = map[string]*zap.Logger{}
	mu         sync.RWMutex

	defaultLogger  *zap.Logger
	defaultMaxSize = 1 << 10 // 1GB
)

func GetLevel() *zapcore.Level {
	l := new(zapcore.Level)
	if err := l.Set(config.AppMode); err != nil {
		_ = l.Set("info")
	}
	return l
}

func GetEncoder() zapcore.EncoderConfig {
	encoder := zap.NewProductionEncoderConfig()

	//encoder.EncodeTime = TimeEncoder
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	return encoder
}

func InitLogger(directory string, options ...zap.Option) {
	var err error
	if err = os.Mkdir(directory, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}

	logPath = directory
	options = append(options, zap.AddCaller(), zap.AddCallerSkip(1))
	defaultLogger = NewLogger(config.AppName, options...)
}

func GetLogPath() string {
	return logPath
}

func GetDefaultLogger() *zap.Logger {
	return defaultLogger
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func NewLogger(logName string, options ...zap.Option) *zap.Logger {
	mu.RLock()
	logger, ok := loggerMap[logName]
	mu.RUnlock()
	if ok {
		return logger
	}

	mu.Lock()
	defer mu.Unlock()

	fileName := fmt.Sprintf("%s/%s.log", logPath, logName)
	sc := config.Section("log")
	fileWriters := []zapcore.WriteSyncer{zapcore.AddSync(&lumberjack.Logger{
		Filename:  fileName,
		MaxSize:   sc.Key("maxsize").MustInt(defaultMaxSize), // MB
		LocalTime: true,
		Compress:  sc.Key("compress").MustBool(true),
	})}

	var (
		logLevel = GetLevel()
		encoder  = GetEncoder()
		cores    []zapcore.Core
	)

	// 文件日志格式
	switch sc.Key("encode_type").String() {
	case "mis":
		cores = append(cores, zapcore.NewCore(NewMisEncoder(encoder), zap.CombineWriteSyncers(fileWriters...), logLevel))
	case "json":
		fallthrough
	default:
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoder), zap.CombineWriteSyncers(fileWriters...), logLevel))
	}

	// debug 日志输出至日志文件、标准输出
	if config.AppMode == "debug" {
		cores = append(cores, func() zapcore.Core {
			consoleWriter, closeOut, err := zap.Open("stdout")
			if err != nil {
				if closeOut != nil {
					closeOut()
				}
				panic(err)
			}
			encoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
			return zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), zap.CombineWriteSyncers(consoleWriter), logLevel)
		}())
	}

	if len(cores) == 0 {
		panic("not found logger core")
	}

	if len(cores) > 1 {
		logger = zap.New(zapcore.NewTee(cores...), options...)
	} else {
		logger = zap.New(cores[0], options...)
	}
	loggerMap[logName] = logger
	return logger
}

func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	defaultLogger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	defaultLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	defaultLogger.Fatal(msg, fields...)
}

func Sugar() *zap.SugaredLogger {
	return defaultLogger.Sugar()
}
