package database

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

type gLogger struct {
	*zap.Logger
	level         gormLogger.LogLevel
	TraceSQL      bool
	SlowThreshold time.Duration
}

func (l *gLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	l.level = level
	return l
}

func (l *gLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info("[gorm] " + fmt.Sprintf(msg, data...))
}

func (l *gLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warn("[gorm] " + fmt.Sprintf(msg, data...))
}

func (l *gLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Error("[gorm] " + fmt.Sprintf(msg, data...))
}

func (l *gLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	if !l.TraceSQL && err == nil && !(l.SlowThreshold != 0 && elapsed > l.SlowThreshold) {
		return
	}

	sql, rows := fc()
	logFields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	switch {
	//case err != nil && l.level >= gormLogger.Error:
	case err != nil:
		logFields = append(logFields, zap.Error(err))
		l.Logger.Error("[gorm] Trace Error", logFields...)
	//case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.level >= gormLogger.Warn:
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold:
		l.Logger.Warn("[gorm] Trace Slow SQL", logFields...)
	//case l.level >= gormLogger.Info:
	case l.TraceSQL:
		l.Logger.Info("[gorm] Trace", logFields...)
	}
}
