package logger

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

type ctxLogger struct {
	*zap.Logger
	ctx iris.Context
}

const traceIdKey = "trace_id"

func WithContext(ctx iris.Context) *ctxLogger {
	return &ctxLogger{
		Logger: defaultLogger,
		ctx:    ctx,
	}
}

func (c *ctxLogger) fields(fields []zap.Field) []zap.Field {
	if c.ctx != nil {
		fields = append(
			[]zap.Field{
				zap.String(traceIdKey, fmt.Sprintf("%v", c.ctx.GetID())),
			},
			fields...)
	}
	return fields
}

func (c *ctxLogger) Debug(msg string, fields ...zap.Field) {
	c.Logger.Debug(msg, c.fields(fields)...)
}

func (c *ctxLogger) Info(msg string, fields ...zap.Field) {
	c.Logger.Info(msg, c.fields(fields)...)
}

func (c *ctxLogger) Warn(msg string, fields ...zap.Field) {
	c.Logger.Warn(msg, c.fields(fields)...)
}

func (c *ctxLogger) Error(msg string, fields ...zap.Field) {
	c.Logger.Error(msg, c.fields(fields)...)
}

func (c *ctxLogger) DPanic(msg string, fields ...zap.Field) {
	c.Logger.DPanic(msg, c.fields(fields)...)
}

func (c *ctxLogger) Panic(msg string, fields ...zap.Field) {
	c.Logger.Panic(msg, c.fields(fields)...)
}

func (c *ctxLogger) Fatal(msg string, fields ...zap.Field) {
	c.Logger.Fatal(msg, c.fields(fields)...)
}
